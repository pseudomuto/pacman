package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/api/common"
)

// recoveryMiddleware returns a gin middleware for panic recovery.
func recoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered in HTTP handler",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", recovered,
		)

		common.JSONError(
			c,
			http.StatusInternalServerError,
			errors.New("internal server error"),
		)
	})
}
