package config

import (
	"os"

	"go.uber.org/fx"
)

// Module provides config dependencies for Fx.
var Module = fx.Module("config", fx.Provide(
	func(path ConfigFilePath) (*Config, error) {
		return LoadFile(path, os.ExpandEnv)
	},
))

// ConfigFilePath is the path to the configuration file.
type ConfigFilePath = string
