package kvtests

import (
	"bytes"
	"context"
	"fmt"
	"kvstore/internal/store/kv"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func RunTests(t *testing.T, db kv.Store) {
	tests := []func(*testing.T, kv.Store){
		testSetGet,
		testGetNotFound,
		testDelete,
		testScan,
		testStopScan,
		testScanPrefixOption,
		testScanLimitOption,
	}

	for _, test := range tests {
		testName := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
		t.Run(testName, func(t *testing.T) {
			test(t, db)
		})
	}
}

func testSetGet(t *testing.T, s kv.Store) {
	ctx := context.Background()

	err := s.Set(ctx, kv.Key("key"), kv.Value("val"))
	require.NoError(t, err)

	val, err := s.Get(ctx, kv.Key("key"))
	require.NoError(t, err)
	require.Equal(t, val, kv.Value("val"))
}

func testGetNotFound(t *testing.T, s kv.Store) {
	ctx := context.Background()

	err := s.Set(ctx, kv.Key("key"), kv.Value("val"))
	require.NoError(t, err)

	_, err = s.Get(ctx, kv.Key("k"))
	require.ErrorIs(t, err, kv.ErrNotFound)
}

func testDelete(t *testing.T, s kv.Store) {
	ctx := context.Background()

	err := s.Set(ctx, kv.Key("key"), kv.Value("val"))
	require.NoError(t, err)

	err = s.Delete(ctx, kv.Key("key"))
	require.NoError(t, err)

	_, err = s.Get(ctx, kv.Key("key"))
	require.ErrorIs(t, err, kv.ErrNotFound)
}

func testScan(t *testing.T, s kv.Store) {
	ctx := context.Background()
	const count = 100

	m := map[string]kv.Value{}
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := kv.Value(fmt.Sprintf("%s-%d", "value", i))

		m[key] = value
		err := s.Set(ctx, kv.Key(key), value)
		require.NoError(t, err)
	}

	r := map[string]kv.Value{}
	err := s.Scan(ctx, kv.ScanOptions{},
		func(k kv.Key, v kv.Value) error {
			r[string(k)] = v
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, m, r)
}

func testStopScan(t *testing.T, s kv.Store) {
	ctx := context.Background()
	const count = 100
	const read = 10

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := kv.Value(fmt.Sprintf("%s-%d", "value", i))

		err := s.Set(ctx, kv.Key(key), value)
		require.NoError(t, err)
	}

	counter := 0
	err := s.Scan(ctx, kv.ScanOptions{},
		func(k kv.Key, v kv.Value) error {
			counter++
			if counter == read {
				return kv.ErrStopScan
			}
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, read, counter)
}

func testScanPrefixOption(t *testing.T, s kv.Store) {
	ctx := context.Background()
	const count = 100
	var prefixes = []string{"a/", "b/", "c/", "d/"}
	var prefixIdx = 0
	var nextPrefix = func() string {
		prefixIdx++
		return prefixes[prefixIdx%len(prefixes)]
	}
	var scanPrefix = prefixes[0]

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%s-%d", nextPrefix(), "key", i)
		value := kv.Value(fmt.Sprintf("%s-%d", "value", i))

		err := s.Set(ctx, kv.Key(key), value)
		require.NoError(t, err)
	}

	err := s.Scan(ctx, kv.ScanOptions{
		Prefix: []byte(scanPrefix),
	}, func(k kv.Key, v kv.Value) error {
		require.True(t, bytes.HasPrefix(k, []byte(scanPrefix)))
		return nil
	})
	require.NoError(t, err)
}

func testScanLimitOption(t *testing.T, s kv.Store) {
	ctx := context.Background()
	const count = 100
	const read = 10

	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%s-%d", "key", i)
		value := kv.Value(fmt.Sprintf("%s-%d", "value", i))

		err := s.Set(ctx, kv.Key(key), value)
		require.NoError(t, err)
	}

	var counter int
	err := s.Scan(ctx, kv.ScanOptions{
		Limit: read,
	}, func(k kv.Key, v kv.Value) error {
		counter++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, read, counter)
}
