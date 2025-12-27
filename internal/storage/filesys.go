package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pseudomuto/pacman/internal/types"
)

type FileSys struct{}

func NewFileSys() *FileSys {
	return new(FileSys)
}

func (s *FileSys) Type() types.StorageType {
	return types.FileSystem
}

func (s *FileSys) Read(ctx context.Context, w io.Writer, uri string) error {
	f, err := os.Open(uri)
	if err != nil {
		return fmt.Errorf("failed to open file: %s, %w", uri, err)
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(w, f)
	if err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	return nil
}

func (s *FileSys) Write(ctx context.Context, r io.Reader, uri string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(uri), os.FileMode(0o777)); err != nil {
		return "", fmt.Errorf("failed to create parent dir for: %s, %w", uri, err)
	}

	f, err := os.Create(uri)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %s, %w", uri, err)
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, r)
	if err != nil {
		return "", fmt.Errorf("failed to write content: %w", err)
	}

	return uri, nil
}
