package badgerkv

import (
	"context"
	"errors"
	"kvstore/internal/store/kv"

	"github.com/dgraph-io/badger/v4"
	"github.com/sirupsen/logrus"
)

type Config struct {
	InMem bool
	Root  string
}

type Dependencies struct {
	Log *logrus.Logger
}

type badgerkv struct {
	cfg  Config
	deps Dependencies

	db *badger.DB
}

func New(cfg Config, deps Dependencies) (kv.Store, error) {
	ret := &badgerkv{
		cfg:  cfg,
		deps: deps,
	}

	opts := badger.DefaultOptions(cfg.Root)
	opts.WithInMemory(cfg.InMem)
	opts.Logger = deps.Log

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	ret.db = db
	return ret, nil
}

func (b *badgerkv) Set(_ context.Context, k kv.Key, v kv.Value) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(k, v)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *badgerkv) Get(_ context.Context, k kv.Key) (kv.Value, error) {
	var valCopy []byte

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		valCopy, err = item.ValueCopy(nil)
		return err
	})

	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		b.deps.Log.Errorf("failed to get key=%s: %v", k, err)
		return nil, err
	} else if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, kv.ErrNotFound
	}

	return valCopy, nil
}

func (b *badgerkv) Delete(_ context.Context, k kv.Key) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(k)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *badgerkv) Scan(ctx context.Context, opts kv.ScanOptions, h kv.ScanHandler) error {
	limit := -1
	if opts.Limit != 0 {
		limit = opts.Limit
	}

	err := b.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		it := txn.NewIterator(opt)
		defer it.Close()

		for it.Seek(opts.Prefix); it.ValidForPrefix(opts.Prefix); it.Next() {
			limit--
			item := it.Item()

			key := item.Key()
			err := item.Value(func(val []byte) error {
				return h(key, val)
			})
			if err != nil {
				return err
			}

			if limit == 0 {
				break
			}
		}
		return nil
	})
	if err != nil && !errors.Is(err, kv.ErrStopScan) {
		return err
	}

	return nil
}
