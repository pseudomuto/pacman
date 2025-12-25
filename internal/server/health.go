package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/api"
)

// (GET /healthz)
func (s *Server) GetHealthz(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &api.HealthCheckResponse{
		Status: "OK",
	})
}
