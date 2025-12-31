package sumdb

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/pseudomuto/pacman/internal/data"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/sumdbhash"
	"github.com/pseudomuto/pacman/internal/ent/sumdbrecord"
	"github.com/pseudomuto/pacman/internal/ent/sumdbtree"
	"github.com/pseudomuto/sumdb"
	"golang.org/x/mod/sumdb/tlog"
)

// Store implements sumdb.Store. It is bound to a particular tree, meaning it ensures that all queries are bounded to
// a specific tree. This allows for multiple tree if desired.
type Store struct {
	tx     *ent.Tx
	client *ent.Client
	id     int
}

// NewStore creates a new Store bounded to the supplied SumDBTree.
func NewStore(id int, db *ent.Client) *Store {
	return &Store{
		client: db,
		id:     id,
	}
}

func (s *Store) WithTx(ctx context.Context, fn func(sumdb.Store) error) error {
	_, err := data.WithTx(ctx, s.client, func(tx *ent.Tx) (*ent.SumDBTree, error) {
		store := &Store{tx: tx, id: s.id}
		if err := fn(store); err != nil {
			return nil, err
		}

		return tx.SumDBTree.Get(ctx, store.id)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) RecordID(ctx context.Context, path, version string) (int64, error) {
	rec, err := s.records().Query().
		Where(
			sumdbrecord.HasTreeWith(sumdbtree.ID(s.id)),
			sumdbrecord.Path(path),
			sumdbrecord.Version(version),
		).
		Only(ctx)
	if err != nil {
		var nfe *ent.NotFoundError
		if errors.As(err, &nfe) {
			return 0, sumdb.ErrNotFound
		}

		return 0, fmt.Errorf("failed looking up record: %s@%s, %w", path, version, err)
	}

	return rec.RecordID, nil
}

func (s *Store) Records(ctx context.Context, id, n int64) ([]*sumdb.Record, error) {
	recs, err := s.records().Query().
		Where(
			sumdbrecord.HasTreeWith(sumdbtree.ID(s.id)),
			sumdbrecord.RecordIDGTE(id),
		).
		Limit(int(n)).
		Order(sumdbrecord.ByRecordID(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}

	res := make([]*sumdb.Record, len(recs))
	for i := range recs {
		res[i] = &sumdb.Record{
			ID:      recs[i].RecordID,
			Path:    recs[i].Path,
			Version: recs[i].Version,
			Data:    recs[i].Data,
		}
	}

	return res, nil
}

func (s *Store) AddRecord(ctx context.Context, r *sumdb.Record) (int64, error) {
	tree, err := s.trees().Get(ctx, s.id)
	if err != nil {
		return 0, fmt.Errorf("failed to get tree: %d, %w", s.id, err)
	}

	rec, err := s.records().Create().
		SetTreeID(s.id).
		SetRecordID(tree.Size).
		SetPath(r.Path).
		SetVersion(r.Version).
		SetData(r.Data).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to create record: %s@%s, %w", r.Path, r.Version, err)
	}

	return rec.RecordID, nil
}

func (s *Store) ReadHashes(ctx context.Context, indexes []int64) ([]tlog.Hash, error) {
	hashes, err := s.hashes().Query().
		Where(
			sumdbhash.HasTreeWith(sumdbtree.ID(s.id)),
			sumdbhash.IndexIn(indexes...),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read hashes: %w", err)
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
		creates[i] = s.hashes().Create().
			SetTreeID(s.id).
			SetIndex(indexes[i]).
			SetHash(hashes[i][:])
	}

	if err := s.hashes().
		CreateBulk(creates...).
		OnConflictColumns("tree_id", "index").
		UpdateNewValues().
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Store) TreeSize(ctx context.Context) (int64, error) {
	tree, err := s.trees().Get(ctx, s.id)
	if err != nil {
		return 0, fmt.Errorf("failed to get tree: %d, %w", s.id, err)
	}

	return tree.Size, nil
}

func (s *Store) SetTreeSize(ctx context.Context, size int64) error {
	if err := s.trees().
		UpdateOneID(s.id).
		SetSize(size).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to update tree size: %w", err)
	}

	return nil
}

func (s *Store) records() *ent.SumDBRecordClient {
	if s.tx != nil {
		return s.tx.SumDBRecord
	}

	return s.client.SumDBRecord
}

func (s *Store) hashes() *ent.SumDBHashClient {
	if s.tx != nil {
		return s.tx.SumDBHash
	}

	return s.client.SumDBHash
}

func (s *Store) trees() *ent.SumDBTreeClient {
	if s.tx != nil {
		return s.tx.SumDBTree
	}

	return s.client.SumDBTree
}
