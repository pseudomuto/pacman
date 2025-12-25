package storage

import (
	"github.com/pseudomuto/pacman/internal/publisher"
	"go.uber.org/fx"
)

var Module = fx.Module("storage", fx.Provide(
	fx.Annotate(
		NewFileSys,
		fx.As(new(publisher.Uploader)),
		fx.ResultTags(publisher.FXUploaders),
	),
	fx.Annotate(
		NewGCS,
		fx.As(new(publisher.Uploader)),
		fx.ResultTags(publisher.FXUploaders),
	),
))
