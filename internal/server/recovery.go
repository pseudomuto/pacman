package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// recoveryMiddleware returns a gin middleware for panic recovery.
func recoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered in HTTP handler",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"panic", recovered,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}

// createRecoveryMiddleware creates a recovery middleware function.
// This is used by the server to add recovery with proper ordering.
func createRecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return recoveryMiddleware(logger)
}
