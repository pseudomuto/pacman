package types

import "github.com/gin-gonic/gin"

const (
	FXServerRouters = `group:"server_routers"`
)

type Router interface {
	RegisterRoutes(*gin.Engine)
}
