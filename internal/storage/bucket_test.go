package storage_test

import (
	"bytes"
	"testing"

	. "github.com/pseudomuto/pacman/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestBucket(t *testing.T) {
	blob, err := NewBucket(t.Context(), "mem://testing")
	require.NoError(t, err)
	require.NoError(t, blob.Write(t.Context(), bytes.NewBufferString("pfft"), "mem://testing/some/path/here"))

	var buf bytes.Buffer
	require.NoError(t, blob.Read(t.Context(), &buf, "mem://testing/some/path/here"))
	require.Equal(t, "pfft", buf.String())
}
