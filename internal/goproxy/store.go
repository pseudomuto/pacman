package goproxy

import (
	"context"
	"fmt"
	"io"

	"github.com/pseudomuto/goproxy"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/asset"
	"github.com/pseudomuto/pacman/internal/ent/sumdbrecord"
	"github.com/pseudomuto/pacman/internal/ent/sumdbtree"
	"github.com/pseudomuto/pacman/internal/types"
)

type (
	Store struct {
		id  int
		db  *ent.Client
		rdr Reader
	}

	Reader interface {
		Read(context.Context, io.Writer, string) error
	}

	reader struct {
		fn func(context.Context, io.Writer, string) error
	}
)

func NewStore(db *ent.Client, treeID int, r Reader) *Store {
	return &Store{
		db:  db,
		id:  treeID,
		rdr: r,
	}
}

func ReaderFunc(fn func(context.Context, io.Writer, string) error) Reader {
	return &reader{fn: fn}
}

func (s *Store) Get(ctx context.Context, path, version string) (*goproxy.ModuleVersion, error) {
	rec, err := s.db.SumDBRecord.Query().
		Where(
			sumdbrecord.HasTreeWith(sumdbtree.ID(s.id)),
			sumdbrecord.Path(path),
			sumdbrecord.Version(version),
		).
		WithAssets().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, goproxy.ErrModuleNotFound
		}

		return nil, fmt.Errorf("failed to find module: %s@%s, %w", path, version, err)
	}

	modURI, err := rec.
		QueryAssets().
		Where(asset.TypeEQ(types.TextFile)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get go.mod URI for: %s@%s, %w", path, version, err)
	}

	zipURI, err := rec.
		QueryAssets().
		Where(asset.TypeEQ(types.Archive)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get zip URI for: %s@%s, %w", path, version, err)
	}

	mv := toModuleVersion(rec)
	mv.ModURI = modURI.URI
	mv.ZipURI = zipURI.URI
	return mv, nil
}

func (s *Store) GetVersions(ctx context.Context, path string) ([]*goproxy.ModuleVersion, error) {
	recs, err := s.db.SumDBRecord.Query().
		Where(
			sumdbrecord.HasTreeWith(sumdbtree.ID(s.id)),
			sumdbrecord.Path(path),
		).
		WithAssets().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find module versions: %s, %w", path, err)
	}

	mvs := make([]*goproxy.ModuleVersion, len(recs))
	for i := range recs {
		modURI, err := recs[i].
			QueryAssets().
			Where(asset.TypeEQ(types.TextFile)).
			Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get go.mod URI for: %s, %w", path, err)
		}

		zipURI, err := recs[i].
			QueryAssets().
			Where(asset.TypeEQ(types.Archive)).
			Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get zip URI for: %s, %w", path, err)
		}

		mvs[i] = toModuleVersion(recs[i])
		mvs[i].ModURI = modURI.URI
		mvs[i].ZipURI = zipURI.URI
	}

	return mvs, nil
}

func (s *Store) ReadFile(ctx context.Context, w io.Writer, uri string) error {
	if err := s.rdr.Read(ctx, w, uri); err != nil {
		return fmt.Errorf("failed to read file: %s, %w", uri, err)
	}

	return nil
}

func (r *reader) Read(ctx context.Context, w io.Writer, uri string) error {
	return r.fn(ctx, w, uri)
}

func toModuleVersion(r *ent.SumDBRecord) *goproxy.ModuleVersion {
	return &goproxy.ModuleVersion{
		Path:      r.Path,
		Version:   r.Version,
		CreatedAt: r.CreatedAt,
	}
}
