package kv

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")
var ErrStopScan = errors.New("stop scan")

type (
	Key   []byte
	Value []byte
)

type ScanOptions struct {
	Limit  int
	Prefix Key
}

type ScanHandler func(Key, Value) error

type Store interface {
	Set(context.Context, Key, Value) error
	Get(context.Context, Key) (Value, error)
	Delete(context.Context, Key) error
	Scan(context.Context, ScanOptions, ScanHandler) error
}
