package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/pseudomuto/pacman/internal/types"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
)

type GCS struct {
	bucket string
}

func NewGCS(bucket string) *GCS {
	if !strings.HasPrefix(bucket, "gs://") {
		bucket = "gs://" + bucket
	}

	return &GCS{bucket: bucket}
}

func (s *GCS) Type() types.StorageType {
	return types.GCS
}

func (s *GCS) Read(ctx context.Context, w io.Writer, uri string) error {
	bucket, err := blob.OpenBucket(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to open bucket: %s, %w", s.bucket, err)
	}
	defer closerDefer(bucket)

	r, err := bucket.NewReader(ctx, uri, nil)
	if err != nil {
		return fmt.Errorf("failed to read: %s, %w", uri, err)
	}
	defer closerDefer(r)

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to write blob content: %w", err)
	}

	return nil
}

func (s *GCS) Write(ctx context.Context, r io.Reader, uri string) (string, error) {
	bucket, err := blob.OpenBucket(ctx, s.bucket)
	if err != nil {
		return "", fmt.Errorf("failed to open bucket: %s, %w", s.bucket, err)
	}
	defer closerDefer(bucket)

	w, err := bucket.NewWriter(ctx, uri, nil)
	if err != nil {
		return "", fmt.Errorf("failed to open: %s, %w", uri, err)
	}
	defer closerDefer(w)

	_, err = io.Copy(w, r)
	if err != nil {
		return "", fmt.Errorf("failed to write blob content: %w", err)
	}

	return path.Join(s.bucket, uri), nil
}

func closerDefer(c io.Closer) {
	_ = c.Close()
}
