package config_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/pseudomuto/pacman/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	env := func(s string) string {
		if s == "$DATABASE_URL" {
			return "sqlite://open_string"
		}

		return s
	}

	r := strings.NewReader(`
addr: ":8080"
metricsAddr: ":9200"
db:
  dialect: sqlite
  dsn: $DATABASE_URL
debug: true`)

	cfg, err := Load(r, env)
	require.NoError(t, err)

	require.Equal(t, ":8080", cfg.Addr)
	require.Equal(t, ":9200", cfg.MetricsAddr)
	require.Equal(t, "sqlite://open_string", cfg.DB.DSN)
	require.True(t, cfg.Debug)
}

func TestLoadFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	f, err := os.Create(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	r := strings.NewReader("addr: :8080\nmetricsAddr: :9200")
	_, err = io.Copy(f, r)
	require.NoError(t, err)

	cfg, err := LoadFile(path, func(s string) string {
		return os.Expand(s, func(key string) string { return key })
	})
	require.NoError(t, err)

	require.Equal(t, ":8080", cfg.Addr)
	require.Equal(t, ":9200", cfg.MetricsAddr)
	require.False(t, cfg.Debug)
}
