package server

import (
	"context"

	"go.uber.org/fx"
)

const (
	FXProxies = `group:"proxies"`
)

var Module = fx.Module(
	"server",
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
