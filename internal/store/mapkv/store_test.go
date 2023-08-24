package mapkv

import (
	"context"
	"fmt"
	"kvstore/internal/store"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetGet(t *testing.T) {
	ctx := context.Background()

	s := NewStore()
	err := s.Set(ctx, store.Key("key"), store.Value("val"))
	require.NoError(t, err)

	val, err := s.Get(ctx, store.Key("key"))
	require.NoError(t, err)
	require.Equal(t, val, store.Value("val"))
}

func TestGetNotFound(t *testing.T) {
	ctx := context.Background()

	s := NewStore()
	err := s.Set(ctx, store.Key("key"), store.Value("val"))
	require.NoError(t, err)

	_, err = s.Get(ctx, store.Key("k"))
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	s := NewStore()
	err := s.Set(ctx, store.Key("key"), store.Value("val"))
	require.NoError(t, err)

	err = s.Delete(ctx, store.Key("key"))
	require.NoError(t, err)

	_, err = s.Get(ctx, store.Key("key"))
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestScan(t *testing.T) {
	ctx := context.Background()
	const count = 100

	s := NewStore()
	m := map[string]store.Value{}
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := store.Value(fmt.Sprintf("%s-%d", "value", i))

		m[key] = value
		err := s.Set(ctx, store.Key(key), value)
		require.NoError(t, err)
	}

	r := map[string]store.Value{}
	err := s.Scan(ctx, store.ScanOptions{},
		func(k store.Key, v store.Value) error {
			r[string(k)] = v
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, m, r)
}

func TestStopScan(t *testing.T) {
	ctx := context.Background()
	const count = 100
	const read = 10

	s := NewStore()
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := store.Value(fmt.Sprintf("%s-%d", "value", i))

		err := s.Set(ctx, store.Key(key), value)
		require.NoError(t, err)
	}

	counter := 0
	err := s.Scan(ctx, store.ScanOptions{},
		func(k store.Key, v store.Value) error {
			counter++
			if counter == read {
				return store.ErrStopScan
			}
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, read, counter)
}
