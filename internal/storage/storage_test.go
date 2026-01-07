package storage_test

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/pseudomuto/pacman/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	dir := t.TempDir()
	rootPaths := []string{
		"file://" + dir + "?create_dir=1&no_tmp_dir=1",
		"mem://testing",
	}

	// Register some storage roots
	require.NoError(t, RegisterBuckets(t.Context(), rootPaths...))

	tests := []struct {
		path    string
		content string
	}{
		{path: "file.txt", content: "content"},
		{path: "sub/dir/file.txt", content: "sub dir content"},
	}

	for _, tt := range tests {
		for _, base := range rootPaths {
			path := strings.Join([]string{base, tt.path}, "/")

			require.NoError(t, Write(
				t.Context(),
				bytes.NewBufferString(tt.content),
				path,
			))

			var buf bytes.Buffer
			require.NoError(t, Read(t.Context(), &buf, path))
			require.Equal(t, tt.content, buf.String())
		}
	}

	t.Run("unmatched paths", func(t *testing.T) {
		t.Parallel()

		require.ErrorIs(t, Read(t.Context(), nil, "wasistdas"), ErrNoStorageForPath)
		require.ErrorIs(t, Read(t.Context(), nil, "s3://whoops/nope"), ErrNoStorageForPath)
	})
}
