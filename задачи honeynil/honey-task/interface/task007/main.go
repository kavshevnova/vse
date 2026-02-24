package main

// Задача: InMemoryRepository + CachedRepository с TTL кэшем.

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type User struct {
	ID    int
	Name  string
	Email string
}

type Repository interface {
	Save(user User) error
	GetByID(id int) (User, error)
	GetAll() ([]User, error)
	Delete(id int) error
}

// --- InMemoryRepository ---

type InMemoryRepository struct {
	mu   sync.RWMutex
	data map[int]User
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{data: make(map[int]User)}
}

func (r *InMemoryRepository) Save(user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[user.ID] = user
	return nil
}

func (r *InMemoryRepository) GetByID(id int) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if u, ok := r.data[id]; ok {
		return u, nil
	}
	return User{}, errors.New("user not found")
}

func (r *InMemoryRepository) GetAll() ([]User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]User, 0, len(r.data))
	for _, u := range r.data {
		users = append(users, u)
	}
	return users, nil
}

func (r *InMemoryRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.data[id]; !ok {
		return errors.New("user not found")
	}
	delete(r.data, id)
	return nil
}

// --- CachedRepository ---

type cacheEntry struct {
	user      User
	expiresAt time.Time
}

type CachedRepository struct {
	repo Repository
	ttl  time.Duration
	mu   sync.RWMutex
	cache map[int]cacheEntry
}

func NewCachedRepository(repo Repository, ttl time.Duration) *CachedRepository {
	return &CachedRepository{
		repo:  repo,
		ttl:   ttl,
		cache: make(map[int]cacheEntry),
	}
}

func (c *CachedRepository) Save(user User) error {
	err := c.repo.Save(user)
	if err == nil {
		c.mu.Lock()
		delete(c.cache, user.ID) // инвалидируем кэш
		c.mu.Unlock()
	}
	return err
}

func (c *CachedRepository) GetByID(id int) (User, error) {
	c.mu.RLock()
	if entry, ok := c.cache[id]; ok && time.Now().Before(entry.expiresAt) {
		c.mu.RUnlock()
		return entry.user, nil
	}
	c.mu.RUnlock()

	user, err := c.repo.GetByID(id)
	if err == nil {
		c.mu.Lock()
		c.cache[id] = cacheEntry{user: user, expiresAt: time.Now().Add(c.ttl)}
		c.mu.Unlock()
	}
	return user, err
}

func (c *CachedRepository) GetAll() ([]User, error) {
	return c.repo.GetAll()
}

func (c *CachedRepository) Delete(id int) error {
	err := c.repo.Delete(id)
	if err == nil {
		c.mu.Lock()
		delete(c.cache, id)
		c.mu.Unlock()
	}
	return err
}

func main() {
	repo := NewInMemoryRepository()
	cached := NewCachedRepository(repo, 5*time.Minute)

	cached.Save(User{1, "Alice", "alice@mail.com"})
	u, _ := cached.GetByID(1)
	fmt.Println(u.Name) // Alice (из кэша)
	u2, _ := cached.GetByID(1)
	fmt.Println(u2.Name) // Alice (из кэша)
}
