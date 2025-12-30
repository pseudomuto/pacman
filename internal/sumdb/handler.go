package sumdb

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/api/common"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/sumdb/api"
)

// Handler implements the generated api.ServerInterface for the sumdb domain.
type Handler struct {
	db *ent.Client
}

// NewHandler creates a new sumdb API handler.
func NewHandler(db *ent.Client) *Handler {
	return &Handler{db: db}
}

// ListTrees implements api.ServerInterface.
func (h *Handler) ListTrees(ctx *gin.Context) {
	trees, err := h.db.SumDBTree.Query().All(ctx)
	if err != nil {
		common.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	res := make(api.TreeList, len(trees))
	for i := range trees {
		res[i] = api.Tree{
			Name:        trees[i].Name,
			VerifierKey: trees[i].VerifierKey,
			CreatedAt:   trees[i].CreatedAt,
			UpdatedAt:   trees[i].UpdatedAt,
		}
	}

	ctx.JSON(http.StatusOK, res)
}

// RegisterRoutes implements types.Router interface.
func (h *Handler) RegisterRoutes(engine *gin.Engine) {
	api.RegisterHandlers(engine, h)
}
