package sumdb

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/ent"
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

func NewSumDBPool(db *ent.Client, trees []*ent.SumDBTree) (SumDBPool, error) {
	var pool SumDBPool
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
		append(opts, sumdb.WithStore(NewStore(t.ID, db)))...,
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
