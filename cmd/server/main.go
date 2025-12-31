package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/pseudomuto/pacman/internal/config"
	"github.com/pseudomuto/pacman/internal/crypto"
	"github.com/pseudomuto/pacman/internal/data"
	"github.com/pseudomuto/pacman/internal/packager"
	"github.com/pseudomuto/pacman/internal/publisher"
	"github.com/pseudomuto/pacman/internal/server"
	"github.com/pseudomuto/pacman/internal/storage"
	"github.com/pseudomuto/pacman/internal/sumdb"
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
			&cli.BoolFlag{
				Name:  "keygen",
				Usage: "Generate a new Tink keyset",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("keygen") {
				kh, err := crypto.CreateKey("keys.bin")
				if err != nil {
					return err
				}

				fmt.Fprintln(cmd.Writer, "Generated Tink keyset:")
				fmt.Fprintln(cmd.Writer, kh.String())
				return nil
			}

			app := fx.New(
				fx.Supply(config.ConfigFilePath(cmd.String("config"))),
				fx.Provide(
					slog.Default,
					func() *prometheus.Registry {
						return prometheus.DefaultRegisterer.(*prometheus.Registry)
					},
				),
				config.Module,
				crypto.Module,
				data.Module,
				packager.Module,
				publisher.Module,
				server.Module,
				storage.Module,
				sumdb.Module,
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
