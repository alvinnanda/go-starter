package cache

import (
	"sync"
	"time"
)

// MemoryStore provides an in-memory cache implementation
type MemoryStore struct {
	items      map[string]item
	mu         sync.RWMutex
	gcInterval time.Duration
	defaultTTL time.Duration
	stopGC     chan bool
}

// item represents a cached item
type item struct {
	value      []byte
	expiration int64 // Unix timestamp
}

// NewMemoryStore creates a new memory store
func NewMemoryStore(config *Config) Store {
	ttl := config.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	gcInterval := time.Minute * 10

	ms := &MemoryStore{
		items:      make(map[string]item),
		gcInterval: gcInterval,
		defaultTTL: ttl,
		stopGC:     make(chan bool),
	}

	// Start garbage collector
	go ms.startGC()

	return ms
}

// Get retrieves an item from the cache by key
func (s *MemoryStore) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	itm, found := s.items[key]
	if !found {
		return nil, false
	}

	// Check expiration
	if itm.expiration > 0 && itm.expiration < time.Now().UnixNano() {
		return nil, false
	}

	return itm.value, true
}

// Set adds an item to the cache with an expiration time
func (s *MemoryStore) Set(key string, value []byte, expiration time.Duration) {
	var exp int64

	if expiration == 0 {
		expiration = s.defaultTTL
	}

	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = item{
		value:      value,
		expiration: exp,
	}
}

// Delete removes an item from the cache
func (s *MemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
}

// Reset clears the entire cache
func (s *MemoryStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string]item)
}

// Close stops the garbage collector
func (s *MemoryStore) Close() {
	s.stopGC <- true
}

// startGC starts the garbage collector to clean expired items
func (s *MemoryStore) startGC() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.deleteExpired()
		case <-s.stopGC:
			return
		}
	}
}

// deleteExpired removes expired items
func (s *MemoryStore) deleteExpired() {
	now := time.Now().UnixNano()

	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.items {
		if v.expiration > 0 && v.expiration < now {
			delete(s.items, k)
		}
	}
}
