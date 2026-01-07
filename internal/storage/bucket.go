package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob" // file://
	_ "gocloud.dev/blob/gcsblob"  // gs://
	_ "gocloud.dev/blob/memblob"  // mem://
	_ "gocloud.dev/blob/s3blob"   // s3://
)

type Bucket struct {
	rootPath string
	bucket   *blob.Bucket
}

func NewBucket(ctx context.Context, baseURL string) (*Bucket, error) {
	bucket, err := blob.OpenBucket(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket: %s, %w", baseURL, err)
	}

	// NB: openers support query string params.
	// We need to allow them for the OpenBucket call above, but remove them for the rootPath.
	path := baseURL
	if idx := strings.IndexByte(path, '?'); idx != -1 {
		path = path[:idx]
	}

	return &Bucket{
		bucket:   bucket,
		rootPath: path,
	}, nil
}

func (s *Bucket) Read(ctx context.Context, w io.Writer, uri string) error {
	path := strings.TrimPrefix(uri, s.rootPath)
	r, err := s.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return fmt.Errorf("failed to read: %s, %w", uri, err)
	}
	defer func() { _ = r.Close() }()

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to write blob content: %w", err)
	}

	return nil
}

func (s *Bucket) Write(ctx context.Context, r io.Reader, uri string) error {
	path := strings.TrimPrefix(uri, s.rootPath)
	w, err := s.bucket.NewWriter(ctx, path, nil)
	if err != nil {
		return fmt.Errorf("failed to open: %s, %w", uri, err)
	}
	defer func() { _ = w.Close() }()

	_, err = io.Copy(w, r)
	if err != nil {
		return fmt.Errorf("failed to write blob content: %w", err)
	}

	return nil
}
