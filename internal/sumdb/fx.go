package sumdb

import (
	"log/slog"

	"github.com/pseudomuto/pacman/internal/types"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"sumdb",
	fx.Provide(
		NewSumDBPool,
		fx.Annotate(
			NewHandler,
			fx.As(new(types.Router)),
			fx.ResultTags(types.FXServerRouters),
		),
	),
	// NB: this is a forcing function to trigger NewSumDBPool.
	// This ensures that sumdb trees are created when necessary on startup.
	fx.Invoke(func(log *slog.Logger, sdbs []*SumDB) {
		log.Debug("Initialized sumdbs", "n", len(sdbs))
	}),
)
