package singleflight

import "sync"

type call struct {
	waitGroup sync.WaitGroup
	val       interface{}
	err       error
}

type Group struct {
	mu      sync.Mutex // protects m
	hashMap map[string]*call
}

/* fn will be only called once no matter how many times Do() is called within interval */
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.hashMap == nil {
		g.hashMap = make(map[string]*call)
	}
	if c, exists := g.hashMap[key]; exists {
		g.mu.Unlock()
		c.waitGroup.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.waitGroup.Add(1)
	g.hashMap[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.waitGroup.Done()

	g.mu.Lock()
	delete(g.hashMap, key)
	g.mu.Unlock()

	return c.val, c.err
}
