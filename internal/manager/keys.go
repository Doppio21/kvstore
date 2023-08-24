package manager

import "errors"

var keyPrefix = []byte("data/")

func wrapDataKey(key []byte) []byte {
	return append(keyPrefix, key...)
}

func unwrapDataKey(key []byte) ([]byte, error) {
	if len(key) <= len(keyPrefix) {
		return nil, errors.New("corrupted key")
	}

	return key[len(keyPrefix):], nil
}
