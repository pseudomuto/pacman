package packager

import (
	"context"
	"fmt"
	"io"

	"github.com/pseudomuto/pacman/internal/types"
	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
)

type GoModule struct{}

func NewGoModule() *GoModule {
	return new(GoModule)
}

func (g *GoModule) Type() types.ArchiveType {
	return types.GoModule
}

func (g *GoModule) Package(ctx context.Context, w io.Writer, opts types.PackageOptions) error {
	if err := zip.CreateFromDir(w, module.Version{
		Path:    opts.Package,
		Version: opts.Version,
	}, opts.Dir); err != nil {
		return fmt.Errorf("failed to create go archive: %w", err)
	}

	return nil
}
