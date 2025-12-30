package data

import (
	"context"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pseudomuto/pacman/internal/config"
	"github.com/pseudomuto/pacman/internal/ent"
	"go.uber.org/fx"
)

type DSN struct {
	Dialect          string
	ConnectionString string
}

var Module = fx.Module("data", fx.Provide(
	NewRepo,
	func(c *config.Config) (*ent.Client, error) {
		client, err := ent.Open(c.DB.Dialect, c.DB.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s connection: %w", c.DB.Dialect, err)
		}

		if err := client.Schema.Create(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to run DB migrations: %w", err)
		}

		return client, nil
	},
))
