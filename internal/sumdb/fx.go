package sumdb

import (
	"github.com/pseudomuto/pacman/internal/types"
	"go.uber.org/fx"
)

var Module = fx.Module("sumdb", fx.Provide(
	NewSumDBPool,
	fx.Annotate(
		NewSumDBProxy,
		fx.As(new(types.Router)),
		fx.ResultTags(types.FXServerRouters),
	),
))
