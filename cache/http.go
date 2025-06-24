package cache

import (
	"bytes"
	"distributed-cache/cache/consistenthash"
	pb "distributed-cache/cache/pb"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
)

const (
	defaultPath              = "/dcache/"
	defaultReplicas          = 100 // Number of vnodes
	defaultReplicationFactor = 3   // number of replicated data
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*HTTPGetter
}

type HTTPGetter struct {
	baseURL string
}

func (h *HTTPGetter) Set(in *pb.SetRequest, out *pb.EmptyResponse) error {
	// Build the POST URL: <baseURL>/<group>/<key>
	u := fmt.Sprintf(
		"%s%s/%s",
		h.baseURL,
		url.PathEscape(in.Group),
		url.PathEscape(in.Key),
	)

	// Marshal the SetRequest as protobuf
	body, err := proto.Marshal(in)
	if err != nil {
		return fmt.Errorf("marshaling SetRequest: %w", err)
	}

	// Fire the POST with the protobuf payload
	resp, err := http.Post(u, "application/octet-stream", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("POST to %s failed: %w", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("peer %s returned status %s", u, resp.Status)
	}
	return nil
}

// HTTP Pool
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultPath,
	}
}

func (pool *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", pool.self, fmt.Sprintf(format, v...))
}

func (pool *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, pool.basePath) {
		panic("Unexpected Path: " + r.URL.Path)
	}
	pool.Log("%s %s", r.Method, r.URL.Path)
	// original: /<basepath>/<groupName>/<key>
	// routes: /<groupName>/<Key>
	subpath := r.URL.Path[len(pool.basePath):]
	routes := strings.SplitN(subpath, "/", 2)
	if len(routes) < 2 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	groupName := routes[0]
	key := routes[1]

	switch r.Method {
	case http.MethodGet:
		group := GetGroup(groupName)
		if group == nil {
			http.Error(w, "Group Not Found: "+groupName, http.StatusNotFound)
			return
		}
		bv, err := group.Get(key) // ByteView, Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		body, err := proto.Marshal(&pb.Response{Value: bv.Bytes()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(body)

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var req pb.SetRequest
		if err := proto.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		group := GetGroup(req.Group)
		if group == nil {
			http.Error(w, "Group Not Found: "+groupName, http.StatusNotFound)
			return
		}
		// Add
		group.cache.Add(req.Key, ByteView{bytes: []byte(req.Value)})
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.NewMap(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*HTTPGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &HTTPGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerClient, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	replicas := p.peers.GetReplicas(key, defaultReplicationFactor)
	for _, peer := range replicas {
		if peer != "" && peer != p.self {
			p.Log("Picked peer %s", peer)
			return p.httpGetters[peer], true
		}
	}

	return nil, false
}

// HTTP Getter
func (h *HTTPGetter) Get(in *pb.GetRequest, out *pb.Response) error {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading resbonse body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

// Compile time assertion
var _ PeerClient = (*HTTPGetter)(nil)
