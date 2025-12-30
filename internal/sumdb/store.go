package sumdb

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/sumdbhash"
	"github.com/pseudomuto/pacman/internal/ent/sumdbrecord"
	"github.com/pseudomuto/sumdb"
	"golang.org/x/mod/sumdb/tlog"
)

// Store implements sumdb.Store. It is bound to a particular tree, meaning it ensures that all queries are bounded to
// a specific tree. This allows for multiple tree if desired.
type Store struct {
	client *ent.Client
	tree   *ent.SumDBTree
}

// NewStore creates a new Store bounded to the supplied SumDBTree.
func NewStore(tree *ent.SumDBTree, db *ent.Client) *Store {
	return &Store{
		client: db,
		tree:   tree,
	}
}

func (s *Store) RecordID(ctx context.Context, path, version string) (int64, error) {
	id, err := s.tree.QueryRecords().
		Where(
			sumdbrecord.Path(path),
			sumdbrecord.Version(version),
		).
		OnlyID(ctx)
	if err != nil {
		var nfe *ent.NotFoundError
		if errors.As(err, &nfe) {
			return 0, sumdb.ErrNotFound
		}

		return 0, fmt.Errorf("failed looking up record: %s@%s, %w", path, version, err)
	}

	return int64(id), nil
}

func (s *Store) Records(ctx context.Context, id, n int64) ([]*sumdb.Record, error) {
	recs, err := s.tree.QueryRecords().
		Where(sumdbrecord.IDGTE(int(id))).
		Limit(int(n)).
		Order(sumdbrecord.ByID(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}

	res := make([]*sumdb.Record, len(recs))
	for i := range recs {
		res[i] = &sumdb.Record{
			ID:      int64(recs[i].ID),
			Path:    recs[i].Path,
			Version: recs[i].Version,
			Data:    recs[i].Data,
		}
	}

	return res, nil
}

func (s *Store) AddRecord(ctx context.Context, r *sumdb.Record) (int64, error) {
	rec, err := s.client.SumDBRecord.Create().
		SetTree(s.tree).
		SetPath(r.Path).
		SetVersion(r.Version).
		SetData(r.Data).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to create record: %s@%s, %w", r.Path, r.Version, err)
	}

	return int64(rec.ID), nil
}

func (s *Store) ReadHashes(ctx context.Context, indexes []int64) ([]tlog.Hash, error) {
	hashes, err := s.tree.QueryHashes().
		Where(sumdbhash.IndexIn(indexes...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read hashes: %s, %w", s.tree.Name, err)
	}

	res := make([]tlog.Hash, len(hashes))
	for i := range hashes {
		res[i] = tlog.Hash(hashes[i].Hash)
	}

	return res, nil
}

func (s *Store) WriteHashes(ctx context.Context, indexes []int64, hashes []tlog.Hash) error {
	creates := make([]*ent.SumDBHashCreate, len(indexes))
	for i := range indexes {
		creates[i] = s.client.SumDBHash.Create().
			SetTree(s.tree).
			SetIndex(indexes[i]).
			SetHash(hashes[i][:])
	}

	if err := s.client.SumDBHash.
		CreateBulk(creates...).
		OnConflictColumns("tree_id", "index").
		UpdateNewValues().
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Store) TreeSize(ctx context.Context) (int64, error) {
	n, err := s.tree.QueryRecords().Count(ctx)
	if err != nil {
		return 0, errors.New("failed to count records")
	}

	return int64(n), nil
}

func (s *Store) SetTreeSize(ctx context.Context, size int64) error {
	return nil
}
