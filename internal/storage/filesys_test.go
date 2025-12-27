package storage_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/pseudomuto/pacman/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestFileSys(t *testing.T) {
	dir := t.TempDir()
	s := NewFileSys()

	file := filepath.Join(dir, "test", "sub", "file.txt")
	path, err := s.Write(t.Context(), strings.NewReader("yo"), file)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, s.Read(t.Context(), &buf, path))
	require.Equal(t, "yo", buf.String())
}
