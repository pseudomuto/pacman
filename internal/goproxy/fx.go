package goproxy

import (
	"log/slog"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"goproxy",
	fx.Provide(
		NewServerPool,
	),
	// NB: this is a forcing function to trigger NewServerPool.
	fx.Invoke(func(log *slog.Logger, svrs []*Server) {
		log.Debug("Initialized goproxy servers", "n", len(svrs))
	}),
)
