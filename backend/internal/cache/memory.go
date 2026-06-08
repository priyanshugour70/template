package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type entry struct {
	value     string
	expiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]entry
}

func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{store: make(map[string]entry)}
	go mc.cleanup()
	return mc
}

func (m *MemoryCache) Get(_ context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.store[key]
	if !ok || (!e.expiresAt.IsZero() && time.Now().After(e.expiresAt)) {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return e.value, nil
}

func (m *MemoryCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	m.store[key] = entry{value: fmt.Sprintf("%v", value), expiresAt: exp}
	return nil
}

func (m *MemoryCache) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, key)
	return nil
}

func (m *MemoryCache) Exists(_ context.Context, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.store[key]
	if !ok || (!e.expiresAt.IsZero() && time.Now().After(e.expiresAt)) {
		return false, nil
	}
	return true, nil
}

func (m *MemoryCache) cleanup() {
	for {
		time.Sleep(time.Minute)
		m.mu.Lock()
		now := time.Now()
		for k, e := range m.store {
			if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
				delete(m.store, k)
			}
		}
		m.mu.Unlock()
	}
}
