package archive_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/pseudomuto/pacman/internal/archive"
	"github.com/pseudomuto/pacman/internal/fsutil"
	"github.com/stretchr/testify/require"
)

func TestCompress(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("testdata", "archive")

	t.Run("tar file", func(t *testing.T) {
		require.NoError(t, fsutil.WithTempFile(func(f *os.File) error {
			return Compress(f, Tar, dir)
		}))
	})

	t.Run("tar.gz file", func(t *testing.T) {
		require.NoError(t, fsutil.WithTempFile(func(f *os.File) error {
			return Compress(f, TarGz, dir, PrefixComponents("repo", "sub", "dir"))
		}))
	})
}
