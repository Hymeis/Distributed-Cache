package cache

import (
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
	defaultPath     = "/dcache/"
	defaultReplicas = 100
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

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Picked peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// HTTP Getter
func (h *HTTPGetter) Get(in *pb.Request, out *pb.Response) error {
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
var _ PeerGetter = (*HTTPGetter)(nil)
