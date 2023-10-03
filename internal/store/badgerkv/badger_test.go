package badgerkv

import (
	"kvstore/internal/store/kv"
	"kvstore/internal/store/kvtests"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func setupTestSuite(t *testing.T) (kv.Store, error) {
	return New(
		Config{Root: t.TempDir(), InMem: true},
		Dependencies{Log: logrus.StandardLogger()})
}

func TestBadger(t *testing.T) {
	s, err := setupTestSuite(t)
	require.NoError(t, err)
	kvtests.RunTests(t, s)
}
