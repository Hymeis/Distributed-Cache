package cache

import (
	lru "distributed-cache/cache/lru_cache"
	"sync"
)

type Cache struct {
	mu        sync.Mutex
	lruCache  *lru.LRUCache
	cacheSize int64
}

func NewCache(size int64, onEvicted func(key string, value lru.Value)) *Cache {
	return &Cache{
		lruCache:  lru.New(size, onEvicted),
		cacheSize: size,
	}
}

func (c *Cache) Get(key string) (value ByteView, exists bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if value, exists := c.lruCache.Get(key); exists {
		return value.(ByteView), exists
	}
	return
}

func (c *Cache) Add(key string, value lru.Value) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lruCache.Add(key, value)
}
