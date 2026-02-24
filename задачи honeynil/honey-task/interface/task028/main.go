package main

// Задача: CDN Edge Cache — multi-tier, coherence, geo-routing, prefetch, origin shield.

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// --- Types ---

type Content struct {
	Key          string
	Data         []byte
	ContentType  string
	ETag         string
	LastModified time.Time
	CacheControl CacheControl
	Metadata     map[string]string
}

type CacheControl struct {
	MaxAge               time.Duration
	SMaxAge              time.Duration
	MustRevalidate       bool
	NoCache              bool
	NoStore              bool
	Public               bool
	Private              bool
	Immutable            bool
	StaleWhileRevalidate time.Duration
}

type CacheTier string

const (
	TierEdge     CacheTier = "edge"
	TierRegional CacheTier = "regional"
	TierOrigin   CacheTier = "origin"
)

type Location struct {
	Latitude   float64
	Longitude  float64
	Region     string
	DataCenter string
}

type CacheStatus string

const (
	StatusHit         CacheStatus = "hit"
	StatusMiss        CacheStatus = "miss"
	StatusStale       CacheStatus = "stale"
	StatusRevalidated CacheStatus = "revalidated"
	StatusBypassed    CacheStatus = "bypassed"
)

type RequestOptions struct {
	IfNoneMatch     string
	IfModifiedSince *time.Time
	Range           *RangeSpec
	ClientLocation  *Location
	AcceptEncoding  []string
}

type RangeSpec struct {
	Start int64
	End   int64
}

type TierStats struct {
	Tier       CacheTier
	HitRate    float64
	Size       int64
	ItemCount  int64
	Evictions  int64
	AvgLatency time.Duration
}

type InvalidationMessage struct {
	Type      InvalidationType
	Keys      []string
	Pattern   string
	Timestamp time.Time
	NodeID    string
}

type InvalidationType string

const (
	InvalidateKey     InvalidationType = "key"
	InvalidatePattern InvalidationType = "pattern"
	InvalidatePurge   InvalidationType = "purge"
)

type NodeStatus string

const (
	NodeStatusHealthy   NodeStatus = "healthy"
	NodeStatusDegraded  NodeStatus = "degraded"
	NodeStatusUnhealthy NodeStatus = "unhealthy"
)

type NodeInfo struct {
	ID          string
	Location    Location
	Tier        CacheTier
	Status      NodeStatus
	Load        float64
	Latency     time.Duration
	LastContact time.Time
}

type ClusterState struct {
	Nodes        []NodeInfo
	TotalSize    int64
	TotalItems   int64
	Synchronized bool
}

type AccessPattern struct {
	Key          string
	AccessCount  int64
	LastAccessed time.Time
	Pattern      []time.Time
	Predicted    []string
}

// --- InMemoryTierCache ---

type cacheItem struct {
	content   Content
	storedAt  time.Time
	hits      int64
}

type InMemoryTierCache struct {
	mu   sync.RWMutex
	data map[string]*cacheItem
	tier CacheTier
	hits int64
	miss int64
}

func NewInMemoryTierCache(tier CacheTier) *InMemoryTierCache {
	return &InMemoryTierCache{data: make(map[string]*cacheItem), tier: tier}
}

func (c *InMemoryTierCache) get(key string) (*Content, CacheStatus) {
	c.mu.RLock()
	item, ok := c.data[key]
	c.mu.RUnlock()
	if !ok {
		c.miss++
		return nil, StatusMiss
	}
	maxAge := item.content.CacheControl.MaxAge
	if maxAge > 0 && time.Since(item.storedAt) > maxAge {
		swr := item.content.CacheControl.StaleWhileRevalidate
		if swr > 0 && time.Since(item.storedAt) <= maxAge+swr {
			c.hits++
			item.hits++
			return &item.content, StatusStale
		}
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		c.miss++
		return nil, StatusMiss
	}
	c.hits++
	item.hits++
	return &item.content, StatusHit
}

func (c *InMemoryTierCache) put(content Content) {
	c.mu.Lock()
	c.data[content.Key] = &cacheItem{content: content, storedAt: time.Now()}
	c.mu.Unlock()
}

func (c *InMemoryTierCache) invalidate(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

func (c *InMemoryTierCache) invalidatePattern(pattern string) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return
	}
	c.mu.Lock()
	for k := range c.data {
		if re.MatchString(k) {
			delete(c.data, k)
		}
	}
	c.mu.Unlock()
}

func (c *InMemoryTierCache) stats() TierStats {
	c.mu.RLock()
	count := int64(len(c.data))
	c.mu.RUnlock()
	total := c.hits + c.miss
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}
	return TierStats{Tier: c.tier, HitRate: hitRate, ItemCount: count}
}

// --- MultiTierEdgeCache ---

type MultiTierEdgeCache struct {
	tiers  map[CacheTier]*InMemoryTierCache
	order  []CacheTier // from fastest to slowest
	cohMu  sync.RWMutex
	subs   []func(InvalidationMessage) error
}

func NewMultiTierEdgeCache() *MultiTierEdgeCache {
	return &MultiTierEdgeCache{
		tiers: map[CacheTier]*InMemoryTierCache{
			TierEdge:     NewInMemoryTierCache(TierEdge),
			TierRegional: NewInMemoryTierCache(TierRegional),
			TierOrigin:   NewInMemoryTierCache(TierOrigin),
		},
		order: []CacheTier{TierEdge, TierRegional, TierOrigin},
	}
}

func (m *MultiTierEdgeCache) Get(ctx context.Context, key string, opts RequestOptions) (*Content, CacheStatus, error) {
	for i, tier := range m.order {
		c, status := m.tiers[tier].get(key)
		if status == StatusHit || status == StatusStale {
			// promote to faster tiers
			if i > 0 {
				m.tiers[m.order[0]].put(*c)
			}
			// ETag conditional
			if opts.IfNoneMatch != "" && opts.IfNoneMatch == c.ETag {
				return nil, StatusRevalidated, nil
			}
			return c, status, nil
		}
	}
	return nil, StatusMiss, nil
}

func (m *MultiTierEdgeCache) Put(_ context.Context, content Content) error {
	if content.CacheControl.NoStore {
		return nil
	}
	for _, tier := range m.tiers {
		tier.put(content)
	}
	return nil
}

func (m *MultiTierEdgeCache) Invalidate(_ context.Context, key string) error {
	for _, tier := range m.tiers {
		tier.invalidate(key)
	}
	m.broadcast(InvalidationMessage{Type: InvalidateKey, Keys: []string{key}, Timestamp: time.Now()})
	return nil
}

func (m *MultiTierEdgeCache) InvalidatePattern(_ context.Context, pattern string) error {
	for _, tier := range m.tiers {
		tier.invalidatePattern(pattern)
	}
	m.broadcast(InvalidationMessage{Type: InvalidatePattern, Pattern: pattern, Timestamp: time.Now()})
	return nil
}

func (m *MultiTierEdgeCache) Purge(ctx context.Context, key string) error {
	return m.Invalidate(ctx, key)
}

func (m *MultiTierEdgeCache) GetFromTier(_ context.Context, tier CacheTier, key string) (*Content, error) {
	c, _ := m.tiers[tier].get(key)
	if c == nil {
		return nil, fmt.Errorf("not found in tier %s", tier)
	}
	return c, nil
}

func (m *MultiTierEdgeCache) Promote(ctx context.Context, key string, fromTier, toTier CacheTier) error {
	c, err := m.GetFromTier(ctx, fromTier, key)
	if err != nil {
		return err
	}
	m.tiers[toTier].put(*c)
	return nil
}

func (m *MultiTierEdgeCache) GetTierStats(_ context.Context, tier CacheTier) (TierStats, error) {
	t, ok := m.tiers[tier]
	if !ok {
		return TierStats{}, fmt.Errorf("unknown tier %s", tier)
	}
	return t.stats(), nil
}

func (m *MultiTierEdgeCache) broadcast(msg InvalidationMessage) {
	m.cohMu.RLock()
	subs := append([]func(InvalidationMessage) error{}, m.subs...)
	m.cohMu.RUnlock()
	for _, sub := range subs {
		sub(msg)
	}
}

func (m *MultiTierEdgeCache) Subscribe(_ context.Context, handler func(InvalidationMessage) error) error {
	m.cohMu.Lock()
	m.subs = append(m.subs, handler)
	m.cohMu.Unlock()
	return nil
}

func (m *MultiTierEdgeCache) GetClusterState(_ context.Context) (ClusterState, error) {
	var total int64
	for _, t := range m.tiers {
		t.mu.RLock()
		total += int64(len(t.data))
		t.mu.RUnlock()
	}
	return ClusterState{TotalItems: total, Synchronized: true}, nil
}

// --- GeoRouter ---

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

type GeoRouter struct {
	mu    sync.RWMutex
	nodes []NodeInfo
}

func NewGeoRouter() *GeoRouter { return &GeoRouter{} }

func (r *GeoRouter) AddNode(node NodeInfo) {
	r.mu.Lock()
	r.nodes = append(r.nodes, node)
	r.mu.Unlock()
}

func (r *GeoRouter) Route(_ context.Context, _ string, client Location) (*NodeInfo, error) {
	nodes, err := r.GetClosestNodes(context.Background(), client, 1)
	if err != nil || len(nodes) == 0 {
		return nil, fmt.Errorf("no healthy nodes")
	}
	return &nodes[0], nil
}

func (r *GeoRouter) GetClosestNodes(_ context.Context, loc Location, count int) ([]NodeInfo, error) {
	r.mu.RLock()
	alive := make([]NodeInfo, 0, len(r.nodes))
	for _, n := range r.nodes {
		if n.Status == NodeStatusHealthy {
			alive = append(alive, n)
		}
	}
	r.mu.RUnlock()
	sort.Slice(alive, func(i, j int) bool {
		di := haversine(loc.Latitude, loc.Longitude, alive[i].Location.Latitude, alive[i].Location.Longitude)
		dj := haversine(loc.Latitude, loc.Longitude, alive[j].Location.Latitude, alive[j].Location.Longitude)
		return di < dj
	})
	if count > len(alive) {
		count = len(alive)
	}
	return alive[:count], nil
}

func (r *GeoRouter) UpdateRouting(_ context.Context) error { return nil }

// --- OriginShieldLayer (singleflight) ---

type shieldCall struct {
	wg      sync.WaitGroup
	content *Content
	err     error
}

type OriginShieldLayer struct {
	mu       sync.Mutex
	inflight map[string]*shieldCall
}

func NewOriginShieldLayer() *OriginShieldLayer {
	return &OriginShieldLayer{inflight: make(map[string]*shieldCall)}
}

func (s *OriginShieldLayer) Shield(_ context.Context, key string, fetcher func() (*Content, error)) (*Content, error) {
	s.mu.Lock()
	if call, ok := s.inflight[key]; ok {
		s.mu.Unlock()
		call.wg.Wait()
		return call.content, call.err
	}
	call := &shieldCall{}
	call.wg.Add(1)
	s.inflight[key] = call
	s.mu.Unlock()

	call.content, call.err = fetcher()
	call.wg.Done()

	s.mu.Lock()
	delete(s.inflight, key)
	s.mu.Unlock()
	return call.content, call.err
}

func (s *OriginShieldLayer) CollapseRequests(ctx context.Context, key string) (*Content, error) {
	return s.Shield(ctx, key, func() (*Content, error) {
		return nil, fmt.Errorf("no fetcher provided")
	})
}

// --- CompressionMiddleware ---

type CompressionMiddleware struct{}

func (c *CompressionMiddleware) Compress(_ context.Context, content []byte, encoding string) ([]byte, error) {
	if encoding != "gzip" {
		return content, nil
	}
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(content); err != nil {
		return nil, err
	}
	w.Close()
	return buf.Bytes(), nil
}

func (c *CompressionMiddleware) Decompress(_ context.Context, content []byte, encoding string) ([]byte, error) {
	if encoding != "gzip" {
		return content, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func (c *CompressionMiddleware) NegotiateEncoding(accepted []string) string {
	for _, enc := range accepted {
		if strings.Contains(enc, "gzip") {
			return "gzip"
		}
	}
	return "identity"
}

func main() {
	cdn := NewMultiTierEdgeCache()
	ctx := context.Background()

	content := Content{
		Key:         "/images/logo.png",
		Data:        []byte("fake-image-data"),
		ContentType: "image/png",
		ETag:        `"abc123"`,
		CacheControl: CacheControl{
			MaxAge: 10 * time.Minute,
			Public: true,
		},
	}
	cdn.Put(ctx, content)

	c, status, _ := cdn.Get(ctx, "/images/logo.png", RequestOptions{})
	fmt.Printf("get: status=%s, key=%s\n", status, c.Key)

	// ETag conditional
	_, status, _ = cdn.Get(ctx, "/images/logo.png", RequestOptions{IfNoneMatch: `"abc123"`})
	fmt.Println("conditional get:", status)

	cdn.Invalidate(ctx, "/images/logo.png")
	_, status, _ = cdn.Get(ctx, "/images/logo.png", RequestOptions{})
	fmt.Println("after invalidate:", status)

	// Geo routing
	router := NewGeoRouter()
	router.AddNode(NodeInfo{
		ID: "edge-eu", Tier: TierEdge, Status: NodeStatusHealthy,
		Location: Location{Latitude: 51.5, Longitude: -0.1, Region: "eu-west"},
	})
	router.AddNode(NodeInfo{
		ID: "edge-us", Tier: TierEdge, Status: NodeStatusHealthy,
		Location: Location{Latitude: 37.8, Longitude: -122.4, Region: "us-west"},
	})
	node, _ := router.Route(ctx, "/images/logo.png", Location{Latitude: 48.8, Longitude: 2.3}) // Paris
	fmt.Println("routed to:", node.ID)

	// Compression
	comp := &CompressionMiddleware{}
	compressed, _ := comp.Compress(ctx, []byte("hello world"), "gzip")
	decompressed, _ := comp.Decompress(ctx, compressed, "gzip")
	fmt.Println("compression roundtrip:", string(decompressed))
}
