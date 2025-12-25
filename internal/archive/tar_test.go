package archive_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/pseudomuto/pacman/internal/archive"
	"github.com/stretchr/testify/require"
)

func TestExtract(t *testing.T) {
	t.Parallel()

	t.Run("tgz", func(t *testing.T) {
		t.Parallel()

		tar, err := makeGzippedTar(map[string][]byte{
			"my-package/bin/executable":    []byte("#!/usr/bin/env bash\necho yo"),
			"my-package/lib/share/thing.o": []byte("some binary content"),
			"my-package/README.md":         []byte("# Details about this package"),
		})
		require.NoError(t, err)

		t.Run("untar", func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, Extract(bytes.NewReader(tar), TarGz, dir))
			require.FileExists(t, filepath.Join(dir, "my-package", "bin", "executable"))
			require.FileExists(t, filepath.Join(dir, "my-package", "lib", "share", "thing.o"))
			require.FileExists(t, filepath.Join(dir, "my-package", "README.md"))

			info, err := os.Stat(filepath.Join(dir, "my-package", "bin", "executable"))
			require.NoError(t, err)
			require.NotZero(t, info.Mode().Perm()&0o111) //nolint:gofumpt // executable bit set
		})

		t.Run("strip components", func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, Extract(bytes.NewReader(tar), TarGz, dir, StripComponents(1)))
			require.FileExists(t, filepath.Join(dir, "bin", "executable"))
			require.FileExists(t, filepath.Join(dir, "lib", "share", "thing.o"))
			require.FileExists(t, filepath.Join(dir, "README.md"))
		})
	})
}

func makeGzippedTar(files map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer

	gzipW := gzip.NewWriter(&buf)
	tarW := tar.NewWriter(gzipW)

	for name, data := range files {
		mode := int64(0o644)
		if strings.HasSuffix(name, "executable") {
			mode = 0o755
		}

		hdr := &tar.Header{
			Name: name,
			Mode: mode,
			Size: int64(len(data)),
		}
		if err := tarW.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tarW.Write(data); err != nil {
			return nil, err
		}
	}

	_ = tarW.Close()
	_ = gzipW.Close()
	return buf.Bytes(), nil
}
