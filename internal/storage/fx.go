package storage

import (
	"context"

	"github.com/pseudomuto/pacman/internal/config"
	"go.uber.org/fx"
)

var Module = fx.Module("storage",
	fx.Invoke(func(ctx context.Context, c *config.Config) error {
		return RegisterBuckets(ctx, c.StorageBuckets...)
	}),
)
