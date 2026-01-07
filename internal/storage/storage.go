package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
)

var (
	bucketLock sync.RWMutex

	ErrNoStorageForPath = errors.New("no storage found for path")

	// all registered storage implementations.
	buckets []*Bucket
)

func RegisterBuckets(ctx context.Context, paths ...string) error {
	bucketLock.Lock()
	defer bucketLock.Unlock()

	buckets = make([]*Bucket, len(paths))
	for i, path := range paths {
		blob, err := NewBucket(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to construct storage for: %s, %w", path, err)
		}

		buckets[i] = blob
	}

	// Sort by length (desc) to avoid a scenario which could lead to incorrect bucket selection when one bucket's
	// rootPath is a prefix of another. For example, if buckets are registered as "file:///tmp" and "file:///tmp/sub", a
	// path "file:///tmp/sub/file.txt" would match the first bucket instead of the more specific second one.
	slices.SortFunc(buckets, func(a, b *Bucket) int {
		return len(b.rootPath) - len(a.rootPath)
	})

	return nil
}

func Read(ctx context.Context, w io.Writer, uri string) error {
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	for _, blob := range buckets {
		if strings.HasPrefix(uri, blob.rootPath) {
			return blob.Read(ctx, w, uri)
		}
	}

	return fmt.Errorf("%w: %s", ErrNoStorageForPath, uri)
}

func Write(ctx context.Context, r io.Reader, uri string) error {
	bucketLock.RLock()
	defer bucketLock.RUnlock()

	for _, blob := range buckets {
		if strings.HasPrefix(uri, blob.rootPath) {
			return blob.Write(ctx, r, uri)
		}
	}

	return fmt.Errorf("%w: %s", ErrNoStorageForPath, uri)
}
