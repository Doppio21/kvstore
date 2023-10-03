package mapkv

import (
	"kvstore/internal/store/kvtests"
	"testing"
)

func TestMapKV(t *testing.T) {
	kvtests.RunTests(t, NewStore())
}
