package manager

import (
	"context"
	"errors"
	"kvstore/internal/store"

	"github.com/golang/snappy"
	"github.com/sirupsen/logrus"
)

var ErrNotFound = store.ErrNotFound

type KeyValuePair struct {
	Key   string
	Value string
}

type GetResult struct {
	KeyValuePair
}

type ScanOptions struct {
}

type ScanResult struct {
	List []KeyValuePair
}

type Manager interface {
	Set(_ context.Context, key []byte, value []byte) error
	Get(_ context.Context, key []byte) (GetResult, error)
	Delete(_ context.Context, key []byte) error
	Scan(context.Context, ScanOptions) (ScanResult, error)
}

type Config struct {
	UseCompression bool
}

type Dependencies struct {
	Store store.Store
	Log   *logrus.Logger
}

type manager struct {
	deps Dependencies
	cfg  Config

	log *logrus.Entry
}

func New(cfg Config, deps Dependencies) Manager {
	return &manager{
		deps: deps,
		cfg:  cfg,
		log:  deps.Log.WithField("component", "manager"),
	}
}

func (m *manager) Set(ctx context.Context, key []byte, value []byte) error {
	data := value
	if m.cfg.UseCompression {
		encodedLen := snappy.MaxEncodedLen(len(value))
		data = make([]byte, encodedLen)
		data = snappy.Encode(data, value)
		m.log.Infof("data compressed %d -> %d", len(value), len(data))
	}
	key = wrapDataKey(key)

	return m.deps.Store.Set(ctx, key, data)
}

func (m *manager) Get(ctx context.Context, key []byte) (GetResult, error) {
	key = wrapDataKey(key)
	res, err := m.deps.Store.Get(ctx, key)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		m.log.Errorf("failed to get key=%s: %v", key, err)
		return GetResult{}, err
	} else if errors.Is(err, store.ErrNotFound) {
		return GetResult{}, ErrNotFound
	}

	data := res
	if m.cfg.UseCompression {
		decodedLen, err := snappy.DecodedLen(res)
		if err != nil {
			return GetResult{}, err
		}

		data = make([]byte, decodedLen)
		data, err = snappy.Decode(data, res)
		if err != nil {
			return GetResult{}, err
		}
	}
	key, err = unwrapDataKey(key)
	if err != nil {
		return GetResult{}, err
	}
	return GetResult{
		KeyValuePair: KeyValuePair{
			Key:   string(key),
			Value: string(data),
		},
	}, nil
}

func (m *manager) Delete(ctx context.Context, key []byte) error {
	key = wrapDataKey(key)
	return m.deps.Store.Delete(ctx, key)
}

func (m *manager) Scan(ctx context.Context, opts ScanOptions) (ScanResult, error) {
	const preallocListSize = 64
	list := make([]KeyValuePair, 0, preallocListSize)
	err := m.deps.Store.Scan(ctx, store.ScanOptions{},
		func(k store.Key, v store.Value) error {
			data := v
			key, err := unwrapDataKey(k)
			if err != nil {
				return err
			}
			if m.cfg.UseCompression {
				decodedLen, err := snappy.DecodedLen(v)
				if err != nil {
					return err
				}
				data = make([]byte, decodedLen)
				data, err = snappy.Decode(data, v)
				if err != nil {
					return err
				}
			}
			list = append(list, KeyValuePair{
				Key:   string(key), 
				Value: string(data),
			})
			return nil
		})
	if err != nil {
		return ScanResult{}, err
	}

	return ScanResult{
		List: list,
	}, nil 
} 
