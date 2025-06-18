package cache
import (
	"sync"
	"fmt"
)

type Getter interface {
	// Get retrieves the data for a given key.
	Get(key string) ([]byte, error)
}

// Encapsulation
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group represents a cache group with a name and a cache.
type Group struct {
	name   	string
	getter 	Getter // Function to get data if not found in cache
	cache  	*Cache
}

var (
	groups = make(map[string]*Group) // All groups
	mu sync.RWMutex 
)

func NewGroup(name string, cacheSize int64, getter Getter) *Group {
	if name == "" {
		panic("Group name cannot be empty")
	}
	if getter == nil {
		panic("Getter cannot be nil")
	}

	mu.Lock()
	defer mu.Unlock()

	group := &Group{
		name:   name,
		getter: getter,
		cache:  NewCache(cacheSize, nil),
	}
	groups[name] = group
	return group
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	group, exists := groups[name]
	if !exists {
		return nil
	}
	return group
}

/*
1. If the key is cached, return the cached value.
2. If the key is not cached, use the getter function to retrieve the value.

TBD: get cached value from other nodes in a distributed cache setup.
*/
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key cannot be empty")
	}

	if value, exists := g.cache.Get(key); exists {
		return value, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	return g.localLoad(key)
}

func (g *Group) localLoad(key string) (ByteView, error) {
	if g.getter == nil {
		return ByteView{}, fmt.Errorf("no getter function defined for group %s", g.name)
	}

	data, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	if len(data) > int(g.cache.cacheSize) {
		return ByteView{}, fmt.Errorf("data size exceeds cache size")
	}

	value := ByteView{bytes: data}
	g.cache.Add(key, value)
	return value, nil
}