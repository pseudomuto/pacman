package goproxy_test

import (
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/ent"
	. "github.com/pseudomuto/pacman/internal/goproxy"
	"github.com/stretchr/testify/require"
)

func TestNewServerPool(t *testing.T) {
	pool, err := NewServerPool(nil, []*ent.SumDBTree{
		{ID: 1, Name: "tree1"},
		{ID: 2, Name: "tree2"},
	})
	require.NoError(t, err)
	require.Len(t, pool.Servers, len(pool.Routers)-1)

	engine := gin.New()
	for i, svr := range pool.Servers {
		svr.RegisterRoutes(engine)

		route := engine.Routes()[i]
		require.Equal(t, "/goproxy/tree"+strconv.Itoa(i+1)+"/*action", route.Path)
		require.NotNil(t, route.Handler)
	}
}
