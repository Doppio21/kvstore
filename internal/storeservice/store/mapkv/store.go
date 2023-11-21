package mapkv

import (
	"context"
	"kvstore/internal/storeservice/store/kv"
	"strings"
	"sync"
)

type Store struct {
	m  map[string]kv.Value
	mu sync.RWMutex
}

func NewStore() kv.Store {
	return &Store{
		m: make(map[string]kv.Value),
	}
}

func (s *Store) Set(_ context.Context, k kv.Key, v kv.Value) error {
	st := string(k)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[st] = v
	return nil
}

func (s *Store) Get(_ context.Context, k kv.Key) (kv.Value, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.m[string(k)]; !ok {
		return nil, kv.ErrNotFound
	} else {
		return v, nil
	}
}

func (s *Store) Delete(_ context.Context, k kv.Key) error {
	st := string(k)

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, st)
	return nil
}

func (s *Store) Scan(_ context.Context, opts kv.ScanOptions, f kv.ScanHandler) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := -1
	if opts.Limit != 0 {
		limit = opts.Limit
	}

	for k, v := range s.m {
		if strings.HasPrefix(k, string(opts.Prefix)) {
			limit--
			if err := f(kv.Key(k), v); err == kv.ErrStopScan {
				break
			}
			if limit == 0 {
				break
			}
		}
	}
	return nil
}
