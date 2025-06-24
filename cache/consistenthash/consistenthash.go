package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys     []int          // virtual nodes
	hashMap  map[int]string // vitrual node, actual key
}

func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add "real" nodes (replicas)
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			offset := strconv.Itoa(i)
			hash := int(m.hash([]byte(offset + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Given a key, return the primary node the key belong to
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}

func (m *Map) GetReplicas(key string, rf int) []string {
	if len(m.keys) == 0 || rf <= 0 {
		return nil
	}
	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	var (
		replicas []string
		seen     = make(map[string]bool)
	)
	for i := 0; len(replicas) < rf; i++ {
		node := m.hashMap[m.keys[(idx+i)%len(m.keys)]]
		if !seen[node] {
			replicas = append(replicas, node)
			seen[node] = true
		}
	}

	return replicas
}
