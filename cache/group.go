package cache

import (
	pb "distributed-cache/cache/pb"
	"distributed-cache/cache/singleflight"
	"fmt"
	"sync"
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
	name   string
	getter Getter // Function to get data if not found in cache
	cache  *Cache
	peers  PeerPicker
	loader *singleflight.Group
}

var (
	groups = make(map[string]*Group) // All groups
	mu     sync.RWMutex
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
		loader: &singleflight.Group{},
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
1. If the key is cached locally, return the cached value.
2. If the key is cached in peer, get the value from the peer.
3. If the key is not cached, use the getter function to retrieve the value.
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

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) loadFn(key string) func() (interface{}, error) {
	return func() (interface{}, error) {
		// Local Load
		if val, err := g.localLoad(key); err == nil {
			return val, nil
		}
		// Peer Load
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if val, err := g.peerLoad(peer, key); err == nil {
					return val, nil
				}
			}
		}
		// Impossible
		return ByteView{}, fmt.Errorf("key %q not found locally or on peers", key)
	}
}

func (g *Group) load(key string) (value ByteView, err error) {
	viewInterface, err := g.loader.Do(key, g.loadFn(key))
	if err == nil {
		return viewInterface.(ByteView), nil
	}
	return
}

func (g *Group) peerLoad(peer PeerClient, key string) (ByteView, error) {
	req := &pb.GetRequest{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes: res.Value}, nil
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

	// Local Add
	g.cache.Add(key, value)

	// Replicas Add
	if httpPool, ok := g.peers.(*HTTPPool); ok {
		go func() {
			peers := httpPool.peers.GetReplicas(key, defaultReplicationFactor)
			req := &pb.SetRequest{Group: g.name, Key: key, Value: data}
			var empty pb.EmptyResponse
			for _, addr := range peers {
				if addr == httpPool.self {
					continue
				}
				getter := httpPool.httpGetters[addr]
				_ = getter.Set(req, &empty)
			}
		}()
	}
	return value, nil
}
