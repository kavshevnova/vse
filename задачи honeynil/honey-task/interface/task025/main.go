package main

// Задача: Graph Database — in-memory с транзакциями, алгоритмами и индексами.

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

type Vertex struct {
	ID         string
	Label      string
	Properties map[string]interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Edge struct {
	ID         string
	Label      string
	FromVertex string
	ToVertex   string
	Properties map[string]interface{}
	Weight     float64
	CreatedAt  time.Time
}

type Path struct {
	Vertices []Vertex
	Edges    []Edge
	Length   int
	Cost     float64
}

type Direction string

const (
	DirectionIn   Direction = "in"
	DirectionOut  Direction = "out"
	DirectionBoth Direction = "both"
)

type TxOptions struct {
	ReadOnly  bool
	Isolation IsolationLevel
}

type IsolationLevel string

const (
	ReadUncommitted IsolationLevel = "read_uncommitted"
	ReadCommitted   IsolationLevel = "read_committed"
	RepeatableRead  IsolationLevel = "repeatable_read"
	Serializable    IsolationLevel = "serializable"
)

type Transaction interface {
	AddVertex(vertex Vertex) error
	GetVertex(id string) (*Vertex, error)
	UpdateVertex(id string, properties map[string]interface{}) error
	DeleteVertex(id string) error
	AddEdge(edge Edge) error
	GetEdge(id string) (*Edge, error)
	DeleteEdge(id string) error
	Commit() error
	Rollback() error
}

type GraphDB interface {
	BeginTx(ctx context.Context, opts TxOptions) (Transaction, error)
	AddVertex(ctx context.Context, vertex Vertex) error
	GetVertex(ctx context.Context, id string) (*Vertex, error)
	UpdateVertex(ctx context.Context, id string, properties map[string]interface{}) error
	DeleteVertex(ctx context.Context, id string) error
	AddEdge(ctx context.Context, edge Edge) error
	GetEdge(ctx context.Context, id string) (*Edge, error)
	DeleteEdge(ctx context.Context, id string) error
	GetNeighbors(ctx context.Context, vertexID string, direction Direction) ([]Vertex, error)
	GetEdges(ctx context.Context, vertexID string, direction Direction) ([]Edge, error)
}

// --- InMemoryGraphDB ---

type graphState struct {
	vertices map[string]Vertex
	edges    map[string]Edge
	// adjacency: vertexID -> edgeIDs
	out map[string][]string
	in  map[string][]string
}

func newGraphState() *graphState {
	return &graphState{
		vertices: make(map[string]Vertex),
		edges:    make(map[string]Edge),
		out:      make(map[string][]string),
		in:       make(map[string][]string),
	}
}

type InMemoryGraphDB struct {
	mu    sync.RWMutex
	state *graphState
}

func NewInMemoryGraphDB() *InMemoryGraphDB {
	return &InMemoryGraphDB{state: newGraphState()}
}

func (db *InMemoryGraphDB) AddVertex(_ context.Context, v Vertex) error {
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	db.mu.Lock()
	db.state.vertices[v.ID] = v
	db.mu.Unlock()
	return nil
}

func (db *InMemoryGraphDB) GetVertex(_ context.Context, id string) (*Vertex, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	v, ok := db.state.vertices[id]
	if !ok {
		return nil, fmt.Errorf("vertex %q not found", id)
	}
	return &v, nil
}

func (db *InMemoryGraphDB) UpdateVertex(_ context.Context, id string, props map[string]interface{}) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	v, ok := db.state.vertices[id]
	if !ok {
		return fmt.Errorf("vertex %q not found", id)
	}
	for k, val := range props {
		v.Properties[k] = val
	}
	v.UpdatedAt = time.Now()
	db.state.vertices[id] = v
	return nil
}

func (db *InMemoryGraphDB) DeleteVertex(_ context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.state.vertices, id)
	for _, eid := range db.state.out[id] {
		e := db.state.edges[eid]
		db.removeFromSlice(db.state.in, e.ToVertex, eid)
		delete(db.state.edges, eid)
	}
	for _, eid := range db.state.in[id] {
		e := db.state.edges[eid]
		db.removeFromSlice(db.state.out, e.FromVertex, eid)
		delete(db.state.edges, eid)
	}
	delete(db.state.out, id)
	delete(db.state.in, id)
	return nil
}

func (db *InMemoryGraphDB) removeFromSlice(m map[string][]string, key, val string) {
	s := m[key]
	for i, v := range s {
		if v == val {
			m[key] = append(s[:i], s[i+1:]...)
			return
		}
	}
}

func (db *InMemoryGraphDB) AddEdge(_ context.Context, e Edge) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	db.mu.Lock()
	db.state.edges[e.ID] = e
	db.state.out[e.FromVertex] = append(db.state.out[e.FromVertex], e.ID)
	db.state.in[e.ToVertex] = append(db.state.in[e.ToVertex], e.ID)
	db.mu.Unlock()
	return nil
}

func (db *InMemoryGraphDB) GetEdge(_ context.Context, id string) (*Edge, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	e, ok := db.state.edges[id]
	if !ok {
		return nil, fmt.Errorf("edge %q not found", id)
	}
	return &e, nil
}

func (db *InMemoryGraphDB) DeleteEdge(_ context.Context, id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	e, ok := db.state.edges[id]
	if !ok {
		return nil
	}
	db.removeFromSlice(db.state.out, e.FromVertex, id)
	db.removeFromSlice(db.state.in, e.ToVertex, id)
	delete(db.state.edges, id)
	return nil
}

func (db *InMemoryGraphDB) GetEdges(_ context.Context, vertexID string, dir Direction) ([]Edge, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var eids []string
	switch dir {
	case DirectionOut:
		eids = db.state.out[vertexID]
	case DirectionIn:
		eids = db.state.in[vertexID]
	case DirectionBoth:
		seen := make(map[string]struct{})
		for _, eid := range append(db.state.out[vertexID], db.state.in[vertexID]...) {
			if _, ok := seen[eid]; !ok {
				seen[eid] = struct{}{}
				eids = append(eids, eid)
			}
		}
	}
	result := make([]Edge, 0, len(eids))
	for _, eid := range eids {
		if e, ok := db.state.edges[eid]; ok {
			result = append(result, e)
		}
	}
	return result, nil
}

func (db *InMemoryGraphDB) GetNeighbors(ctx context.Context, vertexID string, dir Direction) ([]Vertex, error) {
	edges, err := db.GetEdges(ctx, vertexID, dir)
	if err != nil {
		return nil, err
	}
	db.mu.RLock()
	defer db.mu.RUnlock()
	seen := make(map[string]struct{})
	var result []Vertex
	for _, e := range edges {
		nid := e.ToVertex
		if dir == DirectionIn {
			nid = e.FromVertex
		} else if dir == DirectionBoth {
			if e.FromVertex == vertexID {
				nid = e.ToVertex
			} else {
				nid = e.FromVertex
			}
		}
		if _, ok := seen[nid]; !ok {
			seen[nid] = struct{}{}
			if v, ok := db.state.vertices[nid]; ok {
				result = append(result, v)
			}
		}
	}
	return result, nil
}

func (db *InMemoryGraphDB) BeginTx(_ context.Context, _ TxOptions) (Transaction, error) {
	return &memTx{db: db}, nil
}

// --- Simple transaction (no true isolation, just collects ops) ---

type txOp struct{ fn func() error }

type memTx struct {
	db  *InMemoryGraphDB
	ops []txOp
}

func (t *memTx) AddVertex(v Vertex) error {
	t.ops = append(t.ops, txOp{fn: func() error { return t.db.AddVertex(context.Background(), v) }})
	return nil
}
func (t *memTx) GetVertex(id string) (*Vertex, error) {
	return t.db.GetVertex(context.Background(), id)
}
func (t *memTx) UpdateVertex(id string, props map[string]interface{}) error {
	t.ops = append(t.ops, txOp{fn: func() error { return t.db.UpdateVertex(context.Background(), id, props) }})
	return nil
}
func (t *memTx) DeleteVertex(id string) error {
	t.ops = append(t.ops, txOp{fn: func() error { return t.db.DeleteVertex(context.Background(), id) }})
	return nil
}
func (t *memTx) AddEdge(e Edge) error {
	t.ops = append(t.ops, txOp{fn: func() error { return t.db.AddEdge(context.Background(), e) }})
	return nil
}
func (t *memTx) GetEdge(id string) (*Edge, error) { return t.db.GetEdge(context.Background(), id) }
func (t *memTx) DeleteEdge(id string) error {
	t.ops = append(t.ops, txOp{fn: func() error { return t.db.DeleteEdge(context.Background(), id) }})
	return nil
}
func (t *memTx) Commit() error {
	for _, op := range t.ops {
		if err := op.fn(); err != nil {
			return err
		}
	}
	return nil
}
func (t *memTx) Rollback() error { t.ops = nil; return nil }

// --- Graph Algorithms ---

type GraphAlgorithmsImpl struct {
	db *InMemoryGraphDB
}

type dijkstraItem struct {
	id   string
	cost float64
	idx  int
}

type dijkstraHeap []*dijkstraItem

func (h dijkstraHeap) Len() int            { return len(h) }
func (h dijkstraHeap) Less(i, j int) bool  { return h[i].cost < h[j].cost }
func (h dijkstraHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i]; h[i].idx = i; h[j].idx = j }
func (h *dijkstraHeap) Push(x interface{}) { item := x.(*dijkstraItem); item.idx = len(*h); *h = append(*h, item) }
func (h *dijkstraHeap) Pop() interface{}   { old := *h; n := len(old); x := old[n-1]; *h = old[:n-1]; return x }

func (g *GraphAlgorithmsImpl) ShortestPath(ctx context.Context, fromID, toID string) (*Path, error) {
	dist := map[string]float64{fromID: 0}
	prev := map[string]string{}
	prevEdge := map[string]string{}
	pq := &dijkstraHeap{{id: fromID, cost: 0}}
	heap.Init(pq)

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*dijkstraItem)
		if item.id == toID {
			break
		}
		if item.cost > dist[item.id] {
			continue
		}
		edges, _ := g.db.GetEdges(ctx, item.id, DirectionOut)
		for _, e := range edges {
			newCost := dist[item.id] + e.Weight
			if d, ok := dist[e.ToVertex]; !ok || newCost < d {
				dist[e.ToVertex] = newCost
				prev[e.ToVertex] = item.id
				prevEdge[e.ToVertex] = e.ID
				heap.Push(pq, &dijkstraItem{id: e.ToVertex, cost: newCost})
			}
		}
	}

	if _, ok := dist[toID]; !ok {
		return nil, fmt.Errorf("no path from %s to %s", fromID, toID)
	}

	var vids []string
	cur := toID
	for cur != "" {
		vids = append([]string{cur}, vids...)
		cur = prev[cur]
	}
	path := &Path{Cost: dist[toID], Length: len(vids) - 1}
	for _, vid := range vids {
		if v, err := g.db.GetVertex(ctx, vid); err == nil {
			path.Vertices = append(path.Vertices, *v)
		}
	}
	return path, nil
}

func (g *GraphAlgorithmsImpl) BFS(ctx context.Context, startID string, visitor func(Vertex) bool) error {
	visited := map[string]bool{startID: true}
	queue := []string{startID}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		v, err := g.db.GetVertex(ctx, id)
		if err != nil {
			continue
		}
		if !visitor(*v) {
			return nil
		}
		neighbors, _ := g.db.GetNeighbors(ctx, id, DirectionOut)
		for _, n := range neighbors {
			if !visited[n.ID] {
				visited[n.ID] = true
				queue = append(queue, n.ID)
			}
		}
	}
	return nil
}

func (g *GraphAlgorithmsImpl) DFS(ctx context.Context, startID string, visitor func(Vertex) bool) error {
	visited := map[string]bool{}
	var dfs func(id string) bool
	dfs = func(id string) bool {
		if visited[id] {
			return true
		}
		visited[id] = true
		v, err := g.db.GetVertex(ctx, id)
		if err != nil {
			return true
		}
		if !visitor(*v) {
			return false
		}
		neighbors, _ := g.db.GetNeighbors(ctx, id, DirectionOut)
		for _, n := range neighbors {
			if !dfs(n.ID) {
				return false
			}
		}
		return true
	}
	dfs(startID)
	return nil
}

func (g *GraphAlgorithmsImpl) DetectCycle(ctx context.Context) (bool, []string, error) {
	g.db.mu.RLock()
	vids := make([]string, 0, len(g.db.state.vertices))
	for id := range g.db.state.vertices {
		vids = append(vids, id)
	}
	g.db.mu.RUnlock()

	color := make(map[string]int) // 0=white, 1=gray, 2=black
	var cycle []string
	var dfs func(id string) bool
	dfs = func(id string) bool {
		color[id] = 1
		edges, _ := g.db.GetEdges(ctx, id, DirectionOut)
		for _, e := range edges {
			if color[e.ToVertex] == 1 {
				cycle = []string{e.ToVertex, id}
				return true
			}
			if color[e.ToVertex] == 0 && dfs(e.ToVertex) {
				return true
			}
		}
		color[id] = 2
		return false
	}
	for _, id := range vids {
		if color[id] == 0 && dfs(id) {
			return true, cycle, nil
		}
	}
	return false, nil, nil
}

func (g *GraphAlgorithmsImpl) ConnectedComponents(ctx context.Context) ([][]string, error) {
	g.db.mu.RLock()
	vids := make([]string, 0, len(g.db.state.vertices))
	for id := range g.db.state.vertices {
		vids = append(vids, id)
	}
	g.db.mu.RUnlock()

	visited := make(map[string]bool)
	var components [][]string

	var dfs func(id string, comp *[]string)
	dfs = func(id string, comp *[]string) {
		visited[id] = true
		*comp = append(*comp, id)
		neighbors, _ := g.db.GetNeighbors(ctx, id, DirectionBoth)
		for _, n := range neighbors {
			if !visited[n.ID] {
				dfs(n.ID, comp)
			}
		}
	}
	for _, id := range vids {
		if !visited[id] {
			var comp []string
			dfs(id, &comp)
			components = append(components, comp)
		}
	}
	return components, nil
}

func (g *GraphAlgorithmsImpl) PageRank(_ context.Context, iterations int, dampingFactor float64) (map[string]float64, error) {
	g.db.mu.RLock()
	n := len(g.db.state.vertices)
	if n == 0 {
		g.db.mu.RUnlock()
		return nil, nil
	}
	ranks := make(map[string]float64, n)
	for id := range g.db.state.vertices {
		ranks[id] = 1.0 / float64(n)
	}
	g.db.mu.RUnlock()

	for i := 0; i < iterations; i++ {
		newRanks := make(map[string]float64, n)
		for id := range ranks {
			newRanks[id] = (1 - dampingFactor) / float64(n)
		}
		g.db.mu.RLock()
		for _, e := range g.db.state.edges {
			outDeg := len(g.db.state.out[e.FromVertex])
			if outDeg > 0 {
				newRanks[e.ToVertex] += dampingFactor * ranks[e.FromVertex] / float64(outDeg)
			}
		}
		g.db.mu.RUnlock()
		ranks = newRanks
	}
	return ranks, nil
}

func (g *GraphAlgorithmsImpl) AllPaths(ctx context.Context, fromID, toID string, maxDepth int) ([]Path, error) {
	var paths []Path
	var dfs func(cur string, visited map[string]bool, currentPath []string, depth int)
	dfs = func(cur string, visited map[string]bool, currentPath []string, depth int) {
		if depth > maxDepth {
			return
		}
		currentPath = append(currentPath, cur)
		if cur == toID {
			var verts []Vertex
			for _, id := range currentPath {
				v, _ := g.db.GetVertex(ctx, id)
				if v != nil {
					verts = append(verts, *v)
				}
			}
			paths = append(paths, Path{Vertices: verts, Length: len(verts) - 1})
			return
		}
		visited[cur] = true
		edges, _ := g.db.GetEdges(ctx, cur, DirectionOut)
		for _, e := range edges {
			if !visited[e.ToVertex] {
				dfs(e.ToVertex, visited, currentPath, depth+1)
			}
		}
		delete(visited, cur)
	}
	dfs(fromID, make(map[string]bool), nil, 0)
	return paths, nil
}

// --- Simple HashIndex ---

type HashIndex struct {
	mu   sync.RWMutex
	data map[string][]string
}

func NewHashIndex() *HashIndex { return &HashIndex{data: make(map[string][]string)} }

func (idx *HashIndex) Add(key, vertexID string) error {
	idx.mu.Lock()
	idx.data[key] = append(idx.data[key], vertexID)
	idx.mu.Unlock()
	return nil
}

func (idx *HashIndex) Remove(key, vertexID string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	s := idx.data[key]
	for i, v := range s {
		if v == vertexID {
			idx.data[key] = append(s[:i], s[i+1:]...)
			return nil
		}
	}
	return nil
}

func (idx *HashIndex) Search(key string) ([]string, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return append([]string{}, idx.data[key]...), nil
}

func (idx *HashIndex) RangeSearch(from, to string) ([]string, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var result []string
	for k, ids := range idx.data {
		if strings.Compare(k, from) >= 0 && strings.Compare(k, to) <= 0 {
			result = append(result, ids...)
		}
	}
	return result, nil
}

func main() {
	db := NewInMemoryGraphDB()
	ctx := context.Background()

	// Построим граф: A -1-> B -2-> C -1-> D, A -4-> D
	for _, id := range []string{"A", "B", "C", "D"} {
		db.AddVertex(ctx, Vertex{ID: id, Label: "node", Properties: map[string]interface{}{"name": id}})
	}
	db.AddEdge(ctx, Edge{ID: "e1", FromVertex: "A", ToVertex: "B", Weight: 1})
	db.AddEdge(ctx, Edge{ID: "e2", FromVertex: "B", ToVertex: "C", Weight: 2})
	db.AddEdge(ctx, Edge{ID: "e3", FromVertex: "C", ToVertex: "D", Weight: 1})
	db.AddEdge(ctx, Edge{ID: "e4", FromVertex: "A", ToVertex: "D", Weight: 4})

	algo := &GraphAlgorithmsImpl{db: db}
	path, _ := algo.ShortestPath(ctx, "A", "D")
	fmt.Printf("Shortest path A->D: cost=%.1f, vertices=%d\n", path.Cost, len(path.Vertices))

	var visited []string
	algo.BFS(ctx, "A", func(v Vertex) bool {
		visited = append(visited, v.ID)
		return true
	})
	fmt.Println("BFS:", visited)

	components, _ := algo.ConnectedComponents(ctx)
	fmt.Println("Connected components:", len(components))

	hasCycle, cycle, _ := algo.DetectCycle(ctx)
	fmt.Println("Has cycle:", hasCycle, cycle)

	pr, _ := algo.PageRank(ctx, 20, 0.85)
	_ = math.NaN()
	fmt.Printf("PageRank D: %.4f\n", pr["D"])
}
