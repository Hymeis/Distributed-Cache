package lru

import (
	"container/list"
)

// LRUCache is a structure that implements a Least Recently Used (LRU) cache.
type LRUCache struct {
	capacity int64
	size     int64
	cache    map[string]*list.Element
	list     *list.List
	OnEvicted func(key string, value Value) // Called when an entry is evicted
}

type entry struct {
	key   string
	value Value
}

type Value interface{
	Len() int
}

func New(capacity int64, onEvicted func(key string, value Value)) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element, capacity),
		list:     list.New(),
		OnEvicted: onEvicted,
	}
}

func (c *LRUCache) Get(key string) (value Value, exists bool) {
	if element, exists := c.cache[key]; exists {
		c.list.MoveToFront(element)
		kv := element.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *LRUCache) RemoveOldest() {
	element := c.list.Back()
	if element != nil {
		c.list.Remove(element)
		kv := element.Value.(*entry)
		delete(c.cache, kv.key)
		c.size -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *LRUCache) Add(key string, value Value) {
	if element, exists := c.cache[key]; exists {
		c.list.MoveToFront(element)
		kv := element.Value.(*entry)
		c.size -= int64(len(kv.key)) + int64(kv.value.Len())
		kv.value = value
	} else {
		element := c.list.PushFront(&entry{key: key, value: value})
		c.cache[key] = element
		c.size += int64(len(key)) + int64(value.Len())
	}
	for c.capacity != 0 && c.size > c.capacity {
		c.RemoveOldest()
	}
}

func (c *LRUCache) Len() int {
	return c.list.Len()
}