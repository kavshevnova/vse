package main

// Задача: Storage — файловая система, память, составное хранилище.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Metadata struct {
	Size         int64
	LastModified time.Time
	ContentType  string
	Custom       map[string]string
}

type Storage interface {
	Put(key string, data []byte, metadata Metadata) error
	Get(key string) ([]byte, Metadata, error)
	Delete(key string) error
	Exists(key string) (bool, error)
	List(prefix string) ([]string, error)
}

// --- MemoryStorage ---

type entry struct {
	data     []byte
	metadata Metadata
}

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string]entry
}

func NewMemoryStorage() *MemoryStorage { return &MemoryStorage{data: make(map[string]entry)} }

func (m *MemoryStorage) Put(key string, data []byte, meta Metadata) error {
	meta.Size = int64(len(data))
	meta.LastModified = time.Now()
	m.mu.Lock()
	m.data[key] = entry{data: data, metadata: meta}
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) Get(key string) ([]byte, Metadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[key]
	if !ok {
		return nil, Metadata{}, fmt.Errorf("key %q not found", key)
	}
	return e.data, e.metadata, nil
}

func (m *MemoryStorage) Delete(key string) error {
	m.mu.Lock()
	delete(m.data, key)
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) Exists(key string) (bool, error) {
	m.mu.RLock()
	_, ok := m.data[key]
	m.mu.RUnlock()
	return ok, nil
}

func (m *MemoryStorage) List(prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var keys []string
	for k := range m.data {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// --- FileStorage ---

type FileStorage struct {
	root string
}

func NewFileStorage(root string) (*FileStorage, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &FileStorage{root: root}, nil
}

func (f *FileStorage) path(key string) string { return filepath.Join(f.root, key) }

func (f *FileStorage) Put(key string, data []byte, _ Metadata) error {
	p := f.path(key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func (f *FileStorage) Get(key string) ([]byte, Metadata, error) {
	p := f.path(key)
	info, err := os.Stat(p)
	if err != nil {
		return nil, Metadata{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, Metadata{}, err
	}
	return data, Metadata{Size: info.Size(), LastModified: info.ModTime()}, nil
}

func (f *FileStorage) Delete(key string) error { return os.Remove(f.path(key)) }

func (f *FileStorage) Exists(key string) (bool, error) {
	_, err := os.Stat(f.path(key))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (f *FileStorage) List(prefix string) ([]string, error) {
	var keys []string
	err := filepath.Walk(f.root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(f.root, p)
		if strings.HasPrefix(rel, prefix) {
			keys = append(keys, rel)
		}
		return nil
	})
	return keys, err
}

// --- TieredStorage (fast=memory, slow=file) ---

type TieredStorage struct {
	fast Storage
	slow Storage
}

func NewTieredStorage(fast, slow Storage) *TieredStorage { return &TieredStorage{fast: fast, slow: slow} }

func (t *TieredStorage) Put(key string, data []byte, meta Metadata) error {
	if err := t.fast.Put(key, data, meta); err != nil {
		return err
	}
	return t.slow.Put(key, data, meta)
}

func (t *TieredStorage) Get(key string) ([]byte, Metadata, error) {
	if data, meta, err := t.fast.Get(key); err == nil {
		return data, meta, nil
	}
	data, meta, err := t.slow.Get(key)
	if err == nil {
		_ = t.fast.Put(key, data, meta) // promote
	}
	return data, meta, err
}

func (t *TieredStorage) Delete(key string) error {
	_ = t.fast.Delete(key)
	return t.slow.Delete(key)
}

func (t *TieredStorage) Exists(key string) (bool, error) {
	if ok, _ := t.fast.Exists(key); ok {
		return true, nil
	}
	return t.slow.Exists(key)
}

func (t *TieredStorage) List(prefix string) ([]string, error) {
	slow, err := t.slow.List(prefix)
	if err != nil {
		return t.fast.List(prefix)
	}
	seen := make(map[string]struct{})
	for _, k := range slow {
		seen[k] = struct{}{}
	}
	fast, _ := t.fast.List(prefix)
	for _, k := range fast {
		if _, ok := seen[k]; !ok {
			slow = append(slow, k)
		}
	}
	return slow, nil
}

func main() {
	mem := NewMemoryStorage()
	mem.Put("hello", []byte("world"), Metadata{ContentType: "text/plain"})
	data, meta, _ := mem.Get("hello")
	fmt.Printf("mem: %s, size=%d\n", data, meta.Size)

	tiered := NewTieredStorage(NewMemoryStorage(), mem)
	tiered.Put("key1", []byte("value1"), Metadata{})
	d, _, _ := tiered.Get("key1")
	fmt.Println("tiered:", string(d))
}
