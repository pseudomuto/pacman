package proxy

import (
	"github.com/pseudomuto/pacman/internal/server"
	"go.uber.org/fx"
)

var Module = fx.Module("proxy", fx.Provide(
	fx.Annotate(
		NewSumDBProxy,
		fx.As(new(server.Proxy)),
		fx.ResultTags(server.FXProxies),
	),
))
