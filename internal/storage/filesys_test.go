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
	require.NoError(t, s.Write(t.Context(), strings.NewReader("yo"), file))

	var buf bytes.Buffer
	require.NoError(t, s.Read(t.Context(), &buf, file))
	require.Equal(t, "yo", buf.String())
}
