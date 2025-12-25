package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func WithTempDir(fn func(string) error) error {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	return fn(dir)
}

func WithTempFile(fn func(*os.File) error) error {
	return WithTempDir(func(dir string) error {
		fname := filepath.Join(dir, "file.tmp")
		f, err := os.Create(fname)
		if err != nil {
			return fmt.Errorf("failed to create file: %s, %w", fname, err)
		}
		defer func() { _ = f.Close() }()

		return fn(f)
	})
}
