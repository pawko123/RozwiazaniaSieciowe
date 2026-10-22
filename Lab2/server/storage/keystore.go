package storage

import (
	"sync"
)

type KeyStore struct {
	mu   sync.RWMutex
	keys map[string]struct{}
}

func NewKeyStore() *KeyStore {
	return &KeyStore{
		keys: make(map[string]struct{}),
	}
}

func (storage *KeyStore) Contains(id string) bool {
	storage.mu.RLock()
	defer storage.mu.RUnlock()
	_, exists := storage.keys[id]
	return exists
}

func (storage *KeyStore) Add(id string) bool {
	storage.mu.Lock()
	defer storage.mu.Unlock()
	if _, exists := storage.keys[id]; exists {
		return false
	}
	storage.keys[id] = struct{}{}
	return true
}
