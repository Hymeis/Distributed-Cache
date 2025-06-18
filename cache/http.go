package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultPath = "/dcache"

type HTTPPool struct {
	self     string
	basePath string
}

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
	if strings.HasPrefix(r.URL.Path, pool.basePath) {
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

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(bv.Bytes())
}
