package info_test

import (
	"testdata/module/pkg/info"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	require.Equal(t, "0.1.0-pre", info.Version)
}
