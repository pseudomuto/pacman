package server

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/config"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"server",
	fx.Provide(
		func(c *config.Config) *ServerConfig {
			return &ServerConfig{
				ListenAddr:  c.Addr,
				MetricsAddr: c.MetricsAddr,
				GinMode:     gin.ReleaseMode,
				ShowRoutes:  c.Debug,
			}
		},
	),
	fx.Invoke(func(p ServerParams) {
		svr := New(&p)

		p.Lifecycle.Append(fx.Hook{
			OnStart: func(context.Context) error {
				svr.Start()
				return nil
			},
			OnStop: func(context.Context) error {
				// NB: New context because fx's has a 15s timeout.
				return svr.Stop(context.Background())
			},
		})
	}),
)
