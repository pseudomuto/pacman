package goproxy

import (
	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/goproxy"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/storage"
	"github.com/pseudomuto/pacman/internal/types"
	"go.uber.org/fx"
)

type (
	Server struct {
		prefix string
		store  *Store
	}

	ServerPool struct {
		fx.Out

		Routers []types.Router `group:"server_routers,flatten"`
		Servers []*Server
	}
)

func NewServerPool(db *ent.Client, trees []*ent.SumDBTree) (ServerPool, error) {
	var pool ServerPool
	pool.Routers = make([]types.Router, len(trees)+1)
	pool.Servers = make([]*Server, len(trees))

	up := NewUpstreamProxy(db, ReaderFunc(storage.Read))
	pool.Routers[0] = up

	for i := range trees {
		svr := NewServer(db, trees[i])
		pool.Routers[i+1] = svr
		pool.Servers[i] = svr
	}

	return pool, nil
}

func NewServer(db *ent.Client, t *ent.SumDBTree) *Server {
	return &Server{
		prefix: "/goproxy/" + t.Name,
		store:  NewStore(db, t.ID, nil),
	}
}

func (s *Server) RegisterRoutes(g *gin.Engine) {
	h := goproxy.NewServer(s.store, goproxy.WithPathPrefix(s.prefix))
	g.GET(s.prefix+"/*action", gin.WrapH(h))
}
