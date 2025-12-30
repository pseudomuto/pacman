package server

import (
	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/api"
)

func (s *Server) GetSumDBs(ctx *gin.Context) {
	trees, err := s.db.SumDBTree.Query().All(ctx)
	if err != nil {
		ctx.JSON(500, ctx.Error(err))
		return
	}

	res := make(api.SumDBTreeList, len(trees))
	for i := range trees {
		res[i] = api.SumDBTree{
			Name:        trees[i].Name,
			VerifierKey: trees[i].VerifierKey,
			CreatedAt:   trees[i].CreatedAt,
			UpdatedAt:   trees[i].UpdatedAt,
		}
	}

	ctx.JSON(200, res)
}
