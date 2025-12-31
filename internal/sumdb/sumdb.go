package sumdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/config"
	"github.com/pseudomuto/pacman/internal/crypto"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/sumdbtree"
	"github.com/pseudomuto/pacman/internal/types"
	"github.com/pseudomuto/sumdb"
	"go.uber.org/fx"
	ogdb "golang.org/x/mod/sumdb"
)

type (
	SumDB struct {
		name  string
		sumdb *sumdb.SumDB
	}

	SumDBPool struct {
		fx.Out

		Routers []types.Router `group:"server_routers,flatten"`
		SumDBs  []*SumDB
	}
)

func NewSumDBPool(c *config.Config, db *ent.Client) (SumDBPool, error) {
	var pool SumDBPool
	creates := make([]*ent.SumDBTreeCreate, len(c.Go.SumDBs))
	for i, name := range c.Go.SumDBs {
		cr, err := mkTree(db, name)
		if err != nil {
			return pool, err
		}

		creates[i] = cr
	}

	ctx := context.Background()
	if err := db.SumDBTree.
		CreateBulk(creates...).
		OnConflictColumns("name").
		Ignore().
		Exec(ctx); err != nil {
		return pool, fmt.Errorf("failed to create one or more sumdb trees: %w", err)
	}

	trees, err := db.SumDBTree.Query().
		Where(sumdbtree.NameIn(c.Go.SumDBs...)).
		All(ctx)
	if err != nil {
		return pool, fmt.Errorf("failed to query sumdb trees: %w", err)
	}

	pool.Routers = make([]types.Router, len(trees))
	pool.SumDBs = make([]*SumDB, len(trees))
	for i := range trees {
		sdb, err := NewSumDB(trees[i], db)
		if err != nil {
			return pool, fmt.Errorf("failed to create SumDB: %s, %w", trees[i].Name, err)
		}

		pool.Routers[i] = sdb
		pool.SumDBs[i] = sdb
	}

	return pool, nil
}

func NewSumDB(t *ent.SumDBTree, db *ent.Client, opts ...sumdb.Option) (*SumDB, error) {
	sdb, err := sumdb.New(
		t.Name,
		string(t.SignerKey),
		append(opts, sumdb.WithStore(NewStore(t, db)))...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize sumdb: %s, %w", t.Name, err)
	}

	return &SumDB{
		name:  t.Name,
		sumdb: sdb,
	}, nil
}

func (s *SumDB) RegisterRoutes(g *gin.Engine) {
	h := s.sumdb.Handler()
	gh := func(ctx *gin.Context) {
		// NB: The underlying handler checks URL paths for prefixes.
		// Rewrite the paths accordingly by stripping /sumdb/<name>.
		ctx.Request.URL.Path = strings.TrimPrefix(
			ctx.Request.URL.Path,
			"/sumdb/"+s.name,
		)

		h.ServeHTTP(ctx.Writer, ctx.Request)
	}

	group := g.Group("/sumdb/" + s.name)
	for _, path := range ogdb.ServerPaths {
		if path == "/latest" {
			group.GET(path, gh)
			continue
		}

		group.GET(path+"/*data", gh)
	}
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
