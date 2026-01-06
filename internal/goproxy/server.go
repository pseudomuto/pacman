package goproxy

import (
	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/goproxy"
	"github.com/pseudomuto/pacman/internal/ent"
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
	pool.Routers = make([]types.Router, len(trees))
	pool.Servers = make([]*Server, len(trees))

	for i := range trees {
		svr := NewServer(db, trees[i])
		pool.Routers[i] = svr
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
