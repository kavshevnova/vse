package main

// Задача: Temporal Database — версионирование, time-travel queries, GC, MVCC-like.

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Operation string

const (
	OperationInsert Operation = "insert"
	OperationUpdate Operation = "update"
	OperationDelete Operation = "delete"
)

type Version struct {
	ID            int64
	RecordID      string
	Data          map[string]interface{}
	ValidFrom     time.Time
	ValidTo       *time.Time
	CreatedBy     string
	Operation     Operation
	TransactionID string
}

type TimeQuery struct {
	AsOf    *time.Time
	From    *time.Time
	To      *time.Time
	Version *int64
}

type RetentionPolicy struct {
	KeepVersions  int
	KeepDuration  time.Duration
	CompactAfter  time.Duration
	KeepSnapshots int
}

type CompactionStats struct {
	VersionsRemoved int64
	SpaceFreed      int64
	Duration        time.Duration
}

type SnapshotInfo struct {
	ID          string
	Timestamp   time.Time
	Size        int64
	RecordCount int64
}

// --- VersionStore ---

var globalVersionSeq int64

type VersionStore struct {
	mu       sync.RWMutex
	versions map[string][]Version // recordID -> sorted versions
}

func NewVersionStore() *VersionStore {
	return &VersionStore{versions: make(map[string][]Version)}
}

func (s *VersionStore) add(id string, v Version) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Close current head
	vids := s.versions[id]
	now := time.Now()
	if len(vids) > 0 {
		last := &vids[len(vids)-1]
		if last.ValidTo == nil {
			last.ValidTo = &now
			vids[len(vids)-1] = *last
		}
	}
	v.ID = atomic.AddInt64(&globalVersionSeq, 1)
	v.ValidFrom = now
	s.versions[id] = append(vids, v)
}

func (s *VersionStore) current(id string) (*Version, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vs := s.versions[id]
	if len(vs) == 0 {
		return nil, false
	}
	last := vs[len(vs)-1]
	if last.Operation == OperationDelete {
		return nil, false
	}
	return &last, last.ValidTo == nil
}

func (s *VersionStore) asOf(id string, ts time.Time) (*Version, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vs := s.versions[id]
	for i := len(vs) - 1; i >= 0; i-- {
		v := vs[i]
		if !v.ValidFrom.After(ts) && (v.ValidTo == nil || v.ValidTo.After(ts)) {
			if v.Operation == OperationDelete {
				return nil, false
			}
			return &v, true
		}
	}
	return nil, false
}

func (s *VersionStore) history(id string) []Version {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Version{}, s.versions[id]...)
}

// --- TemporalDatabase ---

type TemporalDatabase struct {
	store      *VersionStore
	mu         sync.RWMutex
	retention  RetentionPolicy
	snapshots  map[string]map[string][]Version
}

func NewTemporalDatabase() *TemporalDatabase {
	return &TemporalDatabase{
		store:     NewVersionStore(),
		snapshots: make(map[string]map[string][]Version),
		retention: RetentionPolicy{KeepVersions: 100, KeepDuration: 30 * 24 * time.Hour},
	}
}

func (db *TemporalDatabase) Insert(_ context.Context, id string, data map[string]interface{}) error {
	if _, ok := db.store.current(id); ok {
		return fmt.Errorf("record %q already exists", id)
	}
	db.store.add(id, Version{RecordID: id, Data: data, Operation: OperationInsert})
	return nil
}

func (db *TemporalDatabase) Update(_ context.Context, id string, data map[string]interface{}) error {
	db.store.add(id, Version{RecordID: id, Data: data, Operation: OperationUpdate})
	return nil
}

func (db *TemporalDatabase) Delete(_ context.Context, id string) error {
	db.store.add(id, Version{RecordID: id, Data: nil, Operation: OperationDelete})
	return nil
}

func (db *TemporalDatabase) Get(_ context.Context, id string) (map[string]interface{}, error) {
	v, ok := db.store.current(id)
	if !ok {
		return nil, fmt.Errorf("record %q not found", id)
	}
	return v.Data, nil
}

func (db *TemporalDatabase) GetAsOf(_ context.Context, id string, ts time.Time) (map[string]interface{}, error) {
	v, ok := db.store.asOf(id, ts)
	if !ok {
		return nil, fmt.Errorf("record %q not found at %v", id, ts)
	}
	return v.Data, nil
}

func (db *TemporalDatabase) GetVersion(_ context.Context, id string, versionID int64) (*Version, error) {
	for _, v := range db.store.history(id) {
		if v.ID == versionID {
			vCopy := v
			return &vCopy, nil
		}
	}
	return nil, fmt.Errorf("version %d not found for record %q", versionID, id)
}

func (db *TemporalDatabase) Query(_ context.Context, q TimeQuery, filter map[string]interface{}) ([]map[string]interface{}, error) {
	db.store.mu.RLock()
	ids := make([]string, 0, len(db.store.versions))
	for id := range db.store.versions {
		ids = append(ids, id)
	}
	db.store.mu.RUnlock()

	var result []map[string]interface{}
	for _, id := range ids {
		var data map[string]interface{}
		if q.AsOf != nil {
			if v, ok := db.store.asOf(id, *q.AsOf); ok {
				data = v.Data
			}
		} else {
			if v, ok := db.store.current(id); ok {
				data = v.Data
			}
		}
		if data == nil {
			continue
		}
		match := true
		for k, val := range filter {
			if data[k] != val {
				match = false
				break
			}
		}
		if match {
			result = append(result, data)
		}
	}
	return result, nil
}

// --- VersionHistory ---

func (db *TemporalDatabase) GetHistory(_ context.Context, id string) ([]Version, error) {
	return db.store.history(id), nil
}

func (db *TemporalDatabase) GetChanges(_ context.Context, from, to time.Time) ([]Version, error) {
	db.store.mu.RLock()
	defer db.store.mu.RUnlock()
	var result []Version
	for _, vs := range db.store.versions {
		for _, v := range vs {
			if !v.ValidFrom.Before(from) && !v.ValidFrom.After(to) {
				result = append(result, v)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ValidFrom.Before(result[j].ValidFrom) })
	return result, nil
}

func (db *TemporalDatabase) Diff(_ context.Context, id string, v1ID, v2ID int64) (map[string]interface{}, error) {
	var v1, v2 *Version
	for _, v := range db.store.history(id) {
		vCopy := v
		if v.ID == v1ID {
			v1 = &vCopy
		}
		if v.ID == v2ID {
			v2 = &vCopy
		}
	}
	if v1 == nil || v2 == nil {
		return nil, fmt.Errorf("version not found")
	}
	diff := make(map[string]interface{})
	for k, newVal := range v2.Data {
		if oldVal, ok := v1.Data[k]; !ok || oldVal != newVal {
			diff[k] = map[string]interface{}{"old": oldVal, "new": newVal}
		}
	}
	for k, oldVal := range v1.Data {
		if _, ok := v2.Data[k]; !ok {
			diff[k] = map[string]interface{}{"old": oldVal, "new": nil}
		}
	}
	return diff, nil
}

// --- GarbageCollector ---

type GarbageCollector struct {
	store    *VersionStore
	policy   RetentionPolicy
}

func NewGarbageCollector(store *VersionStore, policy RetentionPolicy) *GarbageCollector {
	return &GarbageCollector{store: store, policy: policy}
}

func (gc *GarbageCollector) Compact(_ context.Context) (CompactionStats, error) {
	start := time.Now()
	gc.store.mu.Lock()
	defer gc.store.mu.Unlock()
	var removed int64
	for id, vs := range gc.store.versions {
		if gc.policy.KeepVersions > 0 && len(vs) > gc.policy.KeepVersions {
			toRemove := len(vs) - gc.policy.KeepVersions
			gc.store.versions[id] = vs[toRemove:]
			removed += int64(toRemove)
		}
		if gc.policy.KeepDuration > 0 {
			cutoff := time.Now().Add(-gc.policy.KeepDuration)
			filtered := vs[:0]
			for _, v := range vs {
				if v.ValidFrom.After(cutoff) || v.ValidTo == nil {
					filtered = append(filtered, v)
				} else {
					removed++
				}
			}
			gc.store.versions[id] = filtered
		}
	}
	return CompactionStats{VersionsRemoved: removed, Duration: time.Since(start)}, nil
}

func main() {
	db := NewTemporalDatabase()
	ctx := context.Background()

	db.Insert(ctx, "user:1", map[string]interface{}{"name": "Alice", "age": 25})
	t1 := time.Now()
	time.Sleep(1 * time.Millisecond)

	db.Update(ctx, "user:1", map[string]interface{}{"name": "Alice Smith", "age": 26})
	time.Sleep(1 * time.Millisecond)

	current, _ := db.Get(ctx, "user:1")
	fmt.Println("current:", current)

	historical, _ := db.GetAsOf(ctx, "user:1", t1)
	fmt.Println("as of t1:", historical)

	history, _ := db.GetHistory(ctx, "user:1")
	fmt.Printf("versions: %d\n", len(history))

	if len(history) >= 2 {
		diff, _ := db.Diff(ctx, "user:1", history[0].ID, history[1].ID)
		fmt.Println("diff:", diff)
	}

	db.Delete(ctx, "user:1")
	_, err := db.Get(ctx, "user:1")
	fmt.Println("after delete:", err)

	gc := NewGarbageCollector(db.store, RetentionPolicy{KeepVersions: 1})
	stats, _ := gc.Compact(ctx)
	fmt.Println("compaction removed:", stats.VersionsRemoved)
}
