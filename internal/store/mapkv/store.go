package mapkv

import (
	"context"
	"kvstore/internal/store"
	"sync"
)

type Store struct {
	m  map[string]store.Value
	mu sync.RWMutex
}

func NewStore() store.Store {
	return &Store{
		m: make(map[string]store.Value),
	}
}

func (s *Store) Set(_ context.Context, k store.Key, v store.Value) error {
	st := string(k)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[st] = v
	return nil
}

func (s *Store) Get(_ context.Context, k store.Key) (store.Value, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.m[string(k)]; !ok {
		return nil, store.ErrNotFound
	} else {
		return v, nil
	}
}

func (s *Store) Delete(_ context.Context, k store.Key) error {
	st := string(k)

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, st)
	return nil
}

func (s *Store) Scan(_ context.Context, opts store.ScanOptions, f store.ScanHandler) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.m {
		if err := f(store.Key(k), v); err == store.ErrStopScan {
			break
		}
	}
	return nil
}
