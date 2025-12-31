package sumdb

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/api/common"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/sumdbhash"
	"github.com/pseudomuto/pacman/internal/ent/sumdbrecord"
	"github.com/pseudomuto/pacman/internal/ent/sumdbtree"
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
			Size:        trees[i].Size,
			VerifierKey: trees[i].VerifierKey,
			CreatedAt:   trees[i].CreatedAt,
			UpdatedAt:   trees[i].UpdatedAt,
		}
	}

	ctx.JSON(http.StatusOK, res)
}

func (h *Handler) ListTreeHashes(ctx *gin.Context, name string) {
	hashes, err := h.db.SumDBHash.Query().
		Where(sumdbhash.HasTreeWith(sumdbtree.NameEqualFold(name))).
		Order(sumdbhash.ByIndex()).
		All(ctx)
	if err != nil {
		common.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	res := make(api.HashList, len(hashes))
	for i := range hashes {
		res[i] = api.Hash{
			Index: hashes[i].Index,
			Hash:  string(hashes[i].Hash),
		}
	}

	ctx.JSON(http.StatusOK, res)
}

func (h *Handler) ListTreeRecords(ctx *gin.Context, name string) {
	records, err := h.db.SumDBRecord.Query().
		Where(
			sumdbrecord.HasTreeWith(sumdbtree.NameEqualFold(name)),
		).
		Order(sumdbrecord.ByPath(), sumdbrecord.ByVersion()).
		All(ctx)
	if err != nil {
		common.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}

	res := make(api.RecordList, len(records))
	for i := range records {
		res[i] = api.Record{
			Id:        records[i].RecordID,
			Path:      records[i].Path,
			Version:   records[i].Version,
			Data:      string(records[i].Data),
			CreatedAt: records[i].CreatedAt,
			UpdatedAt: records[i].UpdatedAt,
		}
	}

	ctx.JSON(http.StatusOK, res)
}

// RegisterRoutes implements types.Router interface.
func (h *Handler) RegisterRoutes(engine *gin.Engine) {
	api.RegisterHandlers(engine, h)
}
