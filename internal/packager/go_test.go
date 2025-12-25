package packager_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/pseudomuto/pacman/internal/packager"
	"github.com/pseudomuto/pacman/internal/types"
	"github.com/stretchr/testify/require"
)

func TestGoModule(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "mod.zip")

	archive, err := os.Create(archivePath)
	require.NoError(t, err)

	mod := NewGoModule()

	// Create the archive
	require.NoError(t, mod.Package(t.Context(), archive, types.PackageOptions{
		Dir:     filepath.Join("..", "..", "testdata", "gomodule"),
		Package: "testdata.io/gomodule",
		Version: "v0.9.1",
	}))
	require.NoError(t, archive.Close())

	// Verify archive contents
	newPackageUtil(t, archivePath).
		HasFile("testdata.io/gomodule@v0.9.1/go.mod").
		HasFile("testdata.io/gomodule@v0.9.1/cmd/server.go").
		HasFile("testdata.io/gomodule@v0.9.1/pkg/info/info.go").
		HasFile("testdata.io/gomodule@v0.9.1/pkg/info/info_test.go").
		NotHasFile("testdata.io/gomodule@v0.9.1/.git/config")
}
