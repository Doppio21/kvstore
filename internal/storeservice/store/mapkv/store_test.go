package mapkv

import (
	"kvstore/internal/storeservice/store/kvtests"
	"testing"
)

func TestMapKV(t *testing.T) {
	kvtests.RunTests(t, NewStore())
}
