package main

import (
	cache "distributed-cache/cache"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Hymeis": "uwu",
	"Lang":   "🐺",
	"Gou":    "🐕",
}

func main() {
	cache.NewGroup("messages", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[Cache] Searching key ", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("key %s does not exist", key)
		}))

	addr := "localhost:8080"
	peers := cache.NewHTTPPool(addr)
	log.Println("Cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
