package manager

import (
	"context"
	"kvstore/internal/storeservice/store/mapkv"
	"sort"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSetGet(t *testing.T) {
	var value = []byte("test-value")
	type test struct {
		name           string
		setKey         string
		getKey         string
		wantErr        error
		useCompression bool
	}

	cases := []test{
		{
			name:    "not_found",
			setKey:  "key1",
			getKey:  "key2",
			wantErr: ErrNotFound,
		},
		{
			name:   "success",
			setKey: "key1",
			getKey: "key1",
		},
		{
			name:           "success_with_compression",
			setKey:         "key1",
			getKey:         "key1",
			useCompression: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := mapkv.NewStore()
			mgr := New(Config{
				UseCompression: c.useCompression,
			}, Dependencies{
				Store: store,
				Log:   logrus.StandardLogger(),
			})

			ctx := context.Background()

			err := mgr.Set(ctx, []byte(c.setKey), value)
			require.NoError(t, err)

			res, err := mgr.Get(ctx, []byte(c.getKey))
			if c.wantErr != nil {
				require.ErrorIs(t, err, c.wantErr)
				return
			}

			require.Equal(t, c.getKey, res.Key)
			require.Equal(t, string(value), res.Value)
		})
	}
}

func TestDelete(t *testing.T) {
	var (
		value = []byte("test-value")
		key   = []byte("test-key")
	)
	store := mapkv.NewStore()
	mgr := New(Config{},
		Dependencies{
			Store: store,
			Log:   logrus.StandardLogger(),
		})

	ctx := context.Background()

	err := mgr.Set(ctx, key, value)
	require.NoError(t, err)

	err = mgr.Delete(ctx, key)
	require.NoError(t, err)

	_, err = mgr.Get(ctx, key)
	require.ErrorIs(t, err, ErrNotFound)

}

func TestScan(t *testing.T) {
	var value = []byte("test-value")
	type test struct {
		name           string
		input          []KeyValuePair
		result         ScanResult
		useCompression bool
	}

	cases := []test{
		{
			name: "success",
			input: []KeyValuePair{
				{Key: "key", Value: string(value)},
				{Key: "key-1", Value: string(value)},
			},
			result: ScanResult{
				List: []KeyValuePair{
					{Key: "key", Value: string(value)},
					{Key: "key-1", Value: string(value)},
				},
			},
		},
		{
			name: "success_with_compression",
			input: []KeyValuePair{
				{Key: "key", Value: string(value)},
				{Key: "key-1", Value: string(value)},
			},
			result: ScanResult{
				List: []KeyValuePair{
					{Key: "key", Value: string(value)},
					{Key: "key-1", Value: string(value)},
				},
			},
			useCompression: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := mapkv.NewStore()
			mgr := New(Config{
				UseCompression: c.useCompression,
			}, Dependencies{
				Store: store,
				Log:   logrus.StandardLogger(),
			})

			ctx := context.Background()

			for _, kv := range c.input {
				err := mgr.Set(ctx, []byte(kv.Key), []byte(kv.Value))
				require.NoError(t, err)
			}

			res, err := mgr.Scan(ctx, ScanOptions{})
			require.NoError(t, err)

			sort.Slice(res.List, func(i, j int) bool {
				return strings.Compare(res.List[i].Key, res.List[j].Key) < 0
			})

			sort.Slice(c.result.List, func(i, j int) bool {
				return strings.Compare(c.result.List[i].Key, c.result.List[j].Key) < 0
			})

			require.Equal(t, c.result, res)
		})
	}
}
