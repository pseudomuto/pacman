package packager

import (
	"github.com/pseudomuto/pacman/internal/publisher"
	"go.uber.org/fx"
)

var Module = fx.Module("packager", fx.Provide(
	fx.Annotate(
		NewGoModule,
		fx.As(new(publisher.Packager)),
		fx.ResultTags(publisher.FXPackagers),
	),
))
