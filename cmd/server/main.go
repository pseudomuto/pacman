package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/pseudomuto/pacman/internal/config"
	"github.com/pseudomuto/pacman/internal/packager"
	"github.com/pseudomuto/pacman/internal/proxy"
	"github.com/pseudomuto/pacman/internal/publisher"
	"github.com/pseudomuto/pacman/internal/server"
	"github.com/pseudomuto/pacman/internal/storage"
	"github.com/urfave/cli/v3"
	"go.uber.org/fx"
)

func main() {
	slog.SetDefault(slog.
		New(slog.NewJSONHandler(os.Stderr, nil)).
		With("app", "pacman"),
	)

	app := &cli.Command{
		Name:  "pacman",
		Usage: "Runs the pacman API server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "config",
				Aliases:   []string{"c"},
				Usage:     "The path to the config file",
				Sources:   cli.EnvVars("PACMAN_CONFIG"),
				TakesFile: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			app := fx.New(
				fx.Supply(config.ConfigFilePath(cmd.String("config"))),
				fx.Provide(
					slog.Default,
					promRegistry,
					serverConfig,
				),
				config.Module,
				packager.Module,
				proxy.Module,
				publisher.Module,
				server.Module,
				storage.Module,
				fx.NopLogger,
			)

			if err := app.Err(); err != nil {
				return err
			}

			app.Run()
			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		slog.Error("failed running server", "err", err)
	}
}

func promRegistry() *prometheus.Registry {
	return prometheus.DefaultRegisterer.(*prometheus.Registry)
}

func serverConfig(c *config.Config) *server.ServerConfig {
	return &server.ServerConfig{
		ListenAddr:  c.Addr,
		MetricsAddr: c.MetricsAddr,
		GinMode:     gin.ReleaseMode,
	}
}
