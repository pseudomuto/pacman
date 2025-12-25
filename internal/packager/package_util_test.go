package packager_test

import (
	"archive/zip"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type packageUtil struct {
	t     *testing.T
	files map[string]*zip.File
}

func newPackageUtil(t *testing.T, path string) *packageUtil {
	t.Helper()

	r, err := zip.OpenReader(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })

	files := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		files[f.Name] = f
	}

	return &packageUtil{t: t, files: files}
}

func (a *packageUtil) HasFile(path string) *packageUtil {
	a.t.Helper()
	_, ok := a.files[path]
	require.True(a.t, ok, "expected file %q to exist in archive", path)
	return a
}

func (a *packageUtil) NotHasFile(path string) *packageUtil {
	a.t.Helper()
	_, ok := a.files[path]
	require.False(a.t, ok, "expected file %q to NOT exist in archive", path)
	return a
}

func (a *packageUtil) HasContent(path, expected string) *packageUtil {
	a.t.Helper()

	f, ok := a.files[path]
	require.True(a.t, ok, "expected file %q to exist in archive", path)

	rc, err := f.Open()
	require.NoError(a.t, err)
	defer rc.Close()

	content, err := io.ReadAll(rc)
	require.NoError(a.t, err)
	require.Equal(a.t, expected, string(content))

	return a
}
