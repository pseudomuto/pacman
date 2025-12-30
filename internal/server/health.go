package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/api/common"
)

// GetHealthz handles the health check endpoint (GET /healthz).
func (s *Server) GetHealthz(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &common.HealthCheckResponse{
		Status: "OK",
	})
}

// RegisterRoutes implements types.Router interface for health endpoint.
func (s *Server) RegisterRoutes(engine *gin.Engine) {
	engine.GET("/healthz", s.GetHealthz)
}
