package boot

import (
	"context"
	"fmt"

	"github.com/pseudomuto/pacman/internal/config"
	"github.com/pseudomuto/pacman/internal/crypto"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/sumdbtree"
	"github.com/pseudomuto/sumdb"
	"go.uber.org/fx"
)

type SumDB struct {
	fx.Out

	Trees []*ent.SumDBTree
}

func InitSumDBs(c *config.Config, db *ent.Client) (SumDB, error) {
	var data SumDB

	creates := make([]*ent.SumDBTreeCreate, len(c.Go.SumDBs))
	for i, name := range c.Go.SumDBs {
		cr, err := mkTree(db, name)
		if err != nil {
			return data, err
		}

		creates[i] = cr
	}

	ctx := context.Background()
	if err := db.SumDBTree.
		CreateBulk(creates...).
		OnConflictColumns("name").
		Ignore().
		Exec(ctx); err != nil {
		return data, fmt.Errorf("failed to create one or more sumdb trees: %w", err)
	}

	trees, err := db.SumDBTree.Query().
		Where(sumdbtree.NameIn(c.Go.SumDBs...)).
		All(ctx)
	if err != nil {
		return data, fmt.Errorf("failed to query sumdb trees: %w", err)
	}

	data.Trees = trees
	return data, nil
}

func mkTree(db *ent.Client, name string) (*ent.SumDBTreeCreate, error) {
	skey, vkey, err := sumdb.GenerateKeys(name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signing keys for tree: %s, %w", name, err)
	}

	return db.SumDBTree.Create().
		SetName(name).
		SetSize(0).
		SetSignerKey(crypto.Secret(skey)).
		SetVerifierKey(vkey), nil
}
