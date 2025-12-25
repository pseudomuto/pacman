package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/pseudomuto/pacman/internal/types"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
)

type GCS struct{}

func NewGCS() *GCS {
	return new(GCS)
}

func (s *GCS) Type() types.StorageType {
	return types.GCS
}

func (s *GCS) Read(ctx context.Context, w io.Writer, uri string) error {
	url, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %s, %w", uri, err)
	}

	bucket, err := blob.OpenBucket(ctx, fmt.Sprintf("%s://%s", url.Scheme, url.Host))
	if err != nil {
		return fmt.Errorf("failed to open bucket: %s, %w", url, err)
	}
	defer closerDefer(bucket)

	r, err := bucket.NewReader(ctx, url.Path, nil)
	if err != nil {
		return fmt.Errorf("failed to read: %s, %w", url.Path, err)
	}
	defer closerDefer(r)

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to write blob content: %w", err)
	}

	return nil
}

func (s *GCS) Write(ctx context.Context, r io.Reader, uri string) error {
	url, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %s, %w", uri, err)
	}

	bucket, err := blob.OpenBucket(ctx, fmt.Sprintf("%s://%s", url.Scheme, url.Host))
	if err != nil {
		return fmt.Errorf("failed to open bucket: %s, %w", url, err)
	}
	defer closerDefer(bucket)

	w, err := bucket.NewWriter(ctx, url.Path, nil)
	if err != nil {
		return fmt.Errorf("failed to open: %s, %w", url.Path, err)
	}
	defer closerDefer(w)

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to write blob content: %w", err)
	}

	return nil
}

func closerDefer(c io.Closer) {
	_ = c.Close()
}
