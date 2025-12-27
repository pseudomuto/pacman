package config

import (
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

type (
	// EnvExpander will replace environment variables in the input.
	// NB: This is satisfied by os.ExpandEnv.
	EnvExpander func(string) string

	// Config holds the configuration for the clickhouse-api service.
	Config struct {
		Addr        string   `yaml:"addr"`
		MetricsAddr string   `yaml:"metricsAddr"`
		Debug       bool     `yaml:"debug"`
		DB          Database `yaml:"db"`
		Go          Go       `yaml:"go"`
	}

	Database struct {
		Dialect string `yaml:"dialect"`
		DSN     string `yaml:"dsn"`
	}

	Go struct {
		NoSumPatterns []string `yaml:"noSumPatterns,omitempty"`
	}
)

// Load reads configuration from an io.Reader, expanding environment variables as needed.
func Load(r io.Reader, exp EnvExpander) (*Config, error) {
	var c Config
	if err := yaml.NewDecoder(r).Decode(&c); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	c.Addr = exp(c.Addr)
	c.MetricsAddr = exp(c.MetricsAddr)
	c.DB.Dialect = exp(c.DB.Dialect)
	c.DB.DSN = exp(c.DB.DSN)

	return &c, nil
}

// LoadFile reads configuration from the file at path.
func LoadFile(path string, env EnvExpander) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s, %w", path, err)
	}
	defer func() { _ = f.Close() }()

	return Load(f, env)
}
