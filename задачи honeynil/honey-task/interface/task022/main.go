package main

// Задача: Distributed Cache — consistent hashing, replication, single-flight, tiered cache.

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"
)

// --- Types from task ---

type CacheEntry struct {
	Key        string
	Value      interface{}
	TTL        time.Duration
	CreatedAt  time.Time
	AccessedAt time.Time
	Version    int64
}

type EvictionPolicy string

const (
	EvictionLRU  EvictionPolicy = "lru"
	EvictionLFU  EvictionPolicy = "lfu"
	EvictionFIFO EvictionPolicy = "fifo"
)

type CacheStrategy string

const (
	WriteThrough CacheStrategy = "write_through"
	WriteBehind  CacheStrategy = "write_behind"
	WriteAround  CacheStrategy = "write_around"
)

type DistributedCache interface {
	Get(ctx context.Context, key string) (interface{}, bool, error)
	GetOrLoad(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	InvalidatePattern(ctx context.Context, pattern string) error
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetMulti(ctx context.Context, entries map[string]interface{}, ttl time.Duration) error
}

type Node interface {
	ID() string
	IsAlive() bool
	Get(ctx context.Context, key string) (interface{}, bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

type ConsistentHash interface {
	Add(node Node) error
	Remove(nodeID string) error
	GetNode(key string) (Node, error)
	GetNodes(key string, count int) ([]Node, error)
}

// --- Local in-memory Node ---

type localNode struct {
	id  string
	mu  sync.RWMutex
	kv  map[string]CacheEntry
}

func newLocalNode(id string) *localNode { return &localNode{id: id, kv: make(map[string]CacheEntry)} }
func (n *localNode) ID() string         { return n.id }
func (n *localNode) IsAlive() bool      { return true }

func (n *localNode) Get(_ context.Context, key string) (interface{}, bool, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	e, ok := n.kv[key]
	if !ok {
		return nil, false, nil
	}
	if e.TTL > 0 && time.Since(e.CreatedAt) > e.TTL {
		return nil, false, nil
	}
	return e.Value, true, nil
}

func (n *localNode) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	n.mu.Lock()
	n.kv[key] = CacheEntry{Key: key, Value: value, TTL: ttl, CreatedAt: time.Now()}
	n.mu.Unlock()
	return nil
}

// --- ConsistentHashRing ---

const defaultVirtualNodes = 150

type virtualNode struct {
	hash   uint64
	nodeID string
}

type ConsistentHashRing struct {
	mu      sync.RWMutex
	vnodes  []virtualNode
	nodes   map[string]Node
	vncount int
}

func NewConsistentHashRing(virtualNodes int) *ConsistentHashRing {
	if virtualNodes <= 0 {
		virtualNodes = defaultVirtualNodes
	}
	return &ConsistentHashRing{nodes: make(map[string]Node), vncount: virtualNodes}
}

func hashKey(key string) uint64 {
	h := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint64(h[:8])
}

func (r *ConsistentHashRing) Add(node Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes[node.ID()] = node
	for i := 0; i < r.vncount; i++ {
		vk := fmt.Sprintf("%s#%d", node.ID(), i)
		r.vnodes = append(r.vnodes, virtualNode{hash: hashKey(vk), nodeID: node.ID()})
	}
	sort.Slice(r.vnodes, func(i, j int) bool { return r.vnodes[i].hash < r.vnodes[j].hash })
	return nil
}

func (r *ConsistentHashRing) Remove(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.nodes, nodeID)
	filtered := r.vnodes[:0]
	for _, vn := range r.vnodes {
		if vn.nodeID != nodeID {
			filtered = append(filtered, vn)
		}
	}
	r.vnodes = filtered
	return nil
}

func (r *ConsistentHashRing) GetNode(key string) (Node, error) {
	nodes, err := r.GetNodes(key, 1)
	if err != nil || len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	return nodes[0], nil
}

func (r *ConsistentHashRing) GetNodes(key string, count int) ([]Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.vnodes) == 0 {
		return nil, fmt.Errorf("ring is empty")
	}
	h := hashKey(key)
	idx := sort.Search(len(r.vnodes), func(i int) bool { return r.vnodes[i].hash >= h })
	seen := make(map[string]struct{})
	var result []Node
	for i := 0; len(result) < count && len(seen) < len(r.nodes); i++ {
		vn := r.vnodes[(idx+i)%len(r.vnodes)]
		if _, ok := seen[vn.nodeID]; ok {
			continue
		}
		seen[vn.nodeID] = struct{}{}
		if n, ok := r.nodes[vn.nodeID]; ok && n.IsAlive() {
			result = append(result, n)
		}
	}
	return result, nil
}

// --- ReplicatedCache ---

type ReplicatedCache struct {
	ring     ConsistentHash
	replicas int
}

func NewReplicatedCache(ring ConsistentHash, replicas int) *ReplicatedCache {
	return &ReplicatedCache{ring: ring, replicas: replicas}
}

func (c *ReplicatedCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	nodes, err := c.ring.GetNodes(key, c.replicas)
	if err != nil {
		return nil, false, err
	}
	for _, n := range nodes {
		if v, ok, err := n.Get(ctx, key); err == nil && ok {
			return v, true, nil
		}
	}
	return nil, false, nil
}

func (c *ReplicatedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	nodes, err := c.ring.GetNodes(key, c.replicas)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if err := n.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

func (c *ReplicatedCache) Delete(ctx context.Context, key string) error {
	nodes, err := c.ring.GetNodes(key, c.replicas)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		n.Set(ctx, key, nil, -1) // mark as deleted via nil
	}
	return nil
}

func (c *ReplicatedCache) InvalidatePattern(ctx context.Context, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	_ = re
	return nil
}

func (c *ReplicatedCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(keys))
	for _, k := range keys {
		if v, ok, _ := c.Get(ctx, k); ok {
			result[k] = v
		}
	}
	return result, nil
}

func (c *ReplicatedCache) GetOrLoad(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error) {
	if v, ok, _ := c.Get(ctx, key); ok {
		return v, nil
	}
	val, err := loader()
	if err != nil {
		return nil, err
	}
	c.Set(ctx, key, val, 5*time.Minute)
	return val, nil
}

func (c *ReplicatedCache) SetMulti(ctx context.Context, entries map[string]interface{}, ttl time.Duration) error {
	for k, v := range entries {
		if err := c.Set(ctx, k, v, ttl); err != nil {
			return err
		}
	}
	return nil
}

// --- SingleFlightCache (cache stampede protection) ---

type inflightCall struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type SingleFlightCache struct {
	DistributedCache
	mu       sync.Mutex
	inflight map[string]*inflightCall
}

func NewSingleFlightCache(inner DistributedCache) *SingleFlightCache {
	return &SingleFlightCache{DistributedCache: inner, inflight: make(map[string]*inflightCall)}
}

func (s *SingleFlightCache) GetOrLoad(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error) {
	if v, ok, _ := s.DistributedCache.Get(ctx, key); ok {
		return v, nil
	}
	s.mu.Lock()
	if call, ok := s.inflight[key]; ok {
		s.mu.Unlock()
		call.wg.Wait()
		return call.val, call.err
	}
	call := &inflightCall{}
	call.wg.Add(1)
	s.inflight[key] = call
	s.mu.Unlock()

	call.val, call.err = loader()
	if call.err == nil {
		s.DistributedCache.Set(ctx, key, call.val, 5*time.Minute)
	}
	call.wg.Done()
	s.mu.Lock()
	delete(s.inflight, key)
	s.mu.Unlock()
	return call.val, call.err
}

func main() {
	ring := NewConsistentHashRing(50)
	ring.Add(newLocalNode("node-1"))
	ring.Add(newLocalNode("node-2"))
	ring.Add(newLocalNode("node-3"))

	cache := NewReplicatedCache(ring, 2)
	ctx := context.Background()

	cache.Set(ctx, "user:1", map[string]string{"name": "Alice"}, 10*time.Minute)
	v, ok, _ := cache.Get(ctx, "user:1")
	fmt.Println("get user:1:", v, ok)

	sf := NewSingleFlightCache(cache)
	val, err := sf.GetOrLoad(ctx, "user:2", func() (interface{}, error) {
		fmt.Println("loading from DB")
		return map[string]string{"name": "Bob"}, nil
	})
	fmt.Println("GetOrLoad:", val, err)
	// Second call — served from cache, no loader call
	val, err = sf.GetOrLoad(ctx, "user:2", func() (interface{}, error) {
		fmt.Println("this should NOT print")
		return nil, nil
	})
	fmt.Println("GetOrLoad (cached):", val, err)
}
