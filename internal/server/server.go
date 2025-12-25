package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pseudomuto/pacman/internal/api"
	"go.uber.org/fx"
)

const (
	defaultListenAddr      = ":8080"
	defaultMetricsAddr     = ":9090"
	defaultReadTimeout     = 30 * time.Second
	defaultWriteTimeout    = 30 * time.Second
	defaultShutdownTimeout = 30 * time.Second
)

type (
	ServerParams struct {
		fx.In

		Config       *ServerConfig
		Lifecycle    fx.Lifecycle
		Logger       *slog.Logger `optional:"true"`
		PromRegistry *prometheus.Registry
		Proxies      []Proxy `group:"proxies"`
	}

	// ServerConfig provides configuration options for the server.
	// All fields are optional and will use defaults if not specified.
	ServerConfig struct {
		// ListenAddr is the address for the HTTP server to listen on.
		// Default: ":8080"
		ListenAddr string
		// MetricsAddr is the address for the metrics server to listen on.
		// Default: ":9090"
		MetricsAddr string

		// Server timeouts
		// ReadTimeout is the maximum duration for reading the entire request.
		// Default: 30s
		ReadTimeout time.Duration
		// WriteTimeout is the maximum duration before timing out writes.
		// Default: 30s
		WriteTimeout time.Duration
		// ShutdownTimeout is the maximum duration to wait for graceful shutdown.
		// Default: 30s
		ShutdownTimeout time.Duration

		// MetricsLabels provides optional constant labels for Prometheus metrics.
		// These labels are added to all metrics to help differentiate between
		// different applications or instances.
		// Example: map[string]string{"service": "user-api", "version": "v1.0.0"}
		MetricsLabels map[string]string

		// GinMode sets the gin framework mode (debug, release, test).
		// Default: release
		GinMode string
	}

	Server struct {
		svr     *http.Server
		cfg     *ServerConfig
		log     *slog.Logger
		promReg *prometheus.Registry
	}

	Proxy interface {
		RegisterRoutes(*gin.Engine)
	}
)

func New(p *ServerParams) *Server {
	p.handleDefaults()

	// Set gin mode
	gin.SetMode(p.Config.GinMode)

	// Create gin engine and metrics middleware
	engine := gin.New()
	engine.Use(createRecoveryMiddleware(p.Logger))
	engine.Use(createMetricsMiddleware(p.PromRegistry, p.Config.MetricsLabels))

	for _, pr := range p.Proxies {
		pr.RegisterRoutes(engine)
	}

	for _, route := range engine.Routes() {
		p.Logger.Info(
			"route",
			"path", route.Path,
			"method", route.Method,
			"handler", route.Handler,
		)
	}

	svr := &Server{
		log:     p.Logger.With("module", "server"),
		cfg:     p.Config,
		promReg: p.PromRegistry,
		svr: &http.Server{
			Addr:         p.Config.ListenAddr,
			Handler:      engine,
			ReadTimeout:  p.Config.ReadTimeout,
			WriteTimeout: p.Config.WriteTimeout,
		},
	}

	// Register generated OpenAPI handlers.
	api.RegisterHandlers(engine, svr)

	return svr
}

func (s *Server) Start() {
	if s.promReg != nil {
		s.log.Info("Starting metrics server", "host", s.cfg.MetricsAddr)
		go func() {
			mux := http.NewServeMux()
			mux.Handle("/metrics", promhttp.HandlerFor(s.promReg, promhttp.HandlerOpts{}))
			_ = http.ListenAndServe(s.cfg.MetricsAddr, mux) // nolint:gosec // don't care about timeout for Prom.
		}()
	}

	s.log.Info("Starting HTTP server", "host", s.cfg.ListenAddr)
	go func() {
		if err := s.svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("HTTP server error", "error", err)
		}
	}()
}

func (s *Server) Stop(ctx context.Context) error {
	// Create shutdown context
	shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()

	return s.svr.Shutdown(shutdownCtx)
}

func (p *ServerParams) handleDefaults() {
	if p.Logger == nil {
		p.Logger = slog.Default()
	}

	if p.Config.ListenAddr == "" {
		p.Config.ListenAddr = defaultListenAddr
	}

	if p.Config.MetricsAddr == "" {
		p.Config.MetricsAddr = defaultMetricsAddr
	}

	if p.Config.ReadTimeout == 0 {
		p.Config.ReadTimeout = defaultReadTimeout
	}

	if p.Config.WriteTimeout == 0 {
		p.Config.WriteTimeout = defaultWriteTimeout
	}

	if p.Config.ShutdownTimeout == 0 {
		p.Config.ShutdownTimeout = defaultShutdownTimeout
	}

	if p.Config.GinMode == "" {
		p.Config.GinMode = gin.ReleaseMode
	}
}
