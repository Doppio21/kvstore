package badgerkv

import (
	"bytes"
	"context"
	"fmt"
	"kvstore/internal/store"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func setupTestSuite(t *testing.T) (store.Store, error) {
	return New(
		Config{Root: t.TempDir(), InMem: true},
		Dependencies{Log: logrus.StandardLogger()})
}

func TestSetGet(t *testing.T) {
	ctx := context.Background()

	s, err := setupTestSuite(t)
	require.NoError(t, err)

	err = s.Set(ctx, store.Key("key"), store.Value("val"))
	require.NoError(t, err)

	val, err := s.Get(ctx, store.Key("key"))
	require.NoError(t, err)
	require.Equal(t, val, store.Value("val"))
}

func TestGetNotFound(t *testing.T) {
	ctx := context.Background()

	s, err := setupTestSuite(t)
	require.NoError(t, err)

	err = s.Set(ctx, store.Key("key"), store.Value("val"))
	require.NoError(t, err)

	_, err = s.Get(ctx, store.Key("k"))
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	s, err := setupTestSuite(t)
	require.NoError(t, err)

	err = s.Set(ctx, store.Key("key"), store.Value("val"))
	require.NoError(t, err)

	err = s.Delete(ctx, store.Key("key"))
	require.NoError(t, err)

	_, err = s.Get(ctx, store.Key("key"))
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestScan(t *testing.T) {
	ctx := context.Background()
	const count = 100

	s, err := setupTestSuite(t)
	require.NoError(t, err)

	m := map[string]store.Value{}
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := store.Value(fmt.Sprintf("%s-%d", "value", i))

		m[key] = value
		err := s.Set(ctx, store.Key(key), value)
		require.NoError(t, err)
	}

	r := map[string]store.Value{}
	err = s.Scan(ctx, store.ScanOptions{},
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

	s, err := setupTestSuite(t)
	require.NoError(t, err)

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := store.Value(fmt.Sprintf("%s-%d", "value", i))

		err := s.Set(ctx, store.Key(key), value)
		require.NoError(t, err)
	}

	counter := 0
	err = s.Scan(ctx, store.ScanOptions{},
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

func TestScanPrefixOption(t *testing.T) {
	ctx := context.Background()
	const count = 100
	var prefixes = []string{"a/", "b/", "c/", "d/"}
	var prefixIdx = 0
	var nextPrefix = func() string {
		prefixIdx++
		return prefixes[prefixIdx%len(prefixes)]
	}
	var scanPrefix = prefixes[0]

	s, err := setupTestSuite(t)
	require.NoError(t, err)

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%s-%d", nextPrefix(), "key", i)
		value := store.Value(fmt.Sprintf("%s-%d", "value", i))

		err := s.Set(ctx, store.Key(key), value)
		require.NoError(t, err)
	}

	err = s.Scan(ctx, store.ScanOptions{
		Prefix: []byte(scanPrefix),
	}, func(k store.Key, v store.Value) error {
		require.True(t, bytes.HasPrefix(k, []byte(scanPrefix)))
		return nil
	})
	require.NoError(t, err)
}

func TestScanLimitOption(t *testing.T) {
	ctx := context.Background()
	const count = 100
	const read = 10

	s, err := setupTestSuite(t)
	require.NoError(t, err)

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := store.Value(fmt.Sprintf("%s-%d", "value", i))

		err := s.Set(ctx, store.Key(key), value)
		require.NoError(t, err)
	}

	var counter int
	err = s.Scan(ctx, store.ScanOptions{
		Limit: read,
	}, func(k store.Key, v store.Value) error {
		counter++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, read, counter)

}