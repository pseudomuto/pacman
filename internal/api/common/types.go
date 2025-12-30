// Package common provides shared types used across API domains.
package common

import "github.com/gin-gonic/gin"

// Error represents a standard API error response.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HealthCheckResponse represents the health check endpoint response.
type HealthCheckResponse struct {
	Status string `json:"status"`
}

// JSONError writes a standardized error response to the gin context.
func JSONError(ctx *gin.Context, code int, err error) {
	ctx.JSON(code, &Error{
		Code:    code,
		Message: err.Error(),
	})

	ctx.Abort()
}
