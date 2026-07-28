// [[ when (modeIs "http") ]]
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"[[.ModulePath]]/internal/config"
	"[[.ModulePath]]/internal/domain/echo"
	"[[.ModulePath]]/internal/observability/metrics"
	httptransport "[[.ModulePath]]/internal/transport/http"
	"[[.ModulePath]]/internal/transport/metricshttp"
)

// server is one listener supervised by Run.
type server struct {
	name  string
	serve func(context.Context) error
}

func Run(ctx context.Context, cfg config.Config) error {
	cfg = cfg.WithDefaults()

	registry, err := metrics.NewRegistry(
		metrics.WithNamespace(cfg.Metrics.MetricPrefix),
		metrics.WithConstLabels(map[string]string(cfg.Labels)),
	)
	if err != nil {
		return fmt.Errorf("metrics registry: %w", err)
	}

	handler := httptransport.Router(echo.NewService(), registry, cfg.HTTP.MaxBodyBytes)

	servers := []server{{
		name: "http",
		serve: func(ctx context.Context) error {
			return serve(ctx, "http", cfg.HTTP.ListenAddr, handler, cfg.HTTP)
		},
	}}
	if cfg.MetricsEnabled() {
		servers = append(servers, server{
			name: "metrics",
			serve: func(ctx context.Context) error {
				return serve(ctx, "metrics", cfg.Metrics.ListenAddr, metricshttp.Handler(registry.Handler()), cfg.HTTP)
			},
		})
		slog.Info("[[.ServiceName]] service started", "http_addr", cfg.HTTP.ListenAddr, "metrics_addr", cfg.Metrics.ListenAddr)
	} else {
		slog.Info("[[.ServiceName]] service started", "http_addr", cfg.HTTP.ListenAddr, "metrics_enabled", false)
	}

	// Cancelling runCtx stops the siblings when any one listener fails, so a
	// partial failure drains the whole process instead of leaving it half up.
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg      sync.WaitGroup
		errOnce sync.Once
		runErr  error
	)
	for _, s := range servers {
		wg.Add(1)
		go func(s server) {
			defer wg.Done()
			if err := s.serve(runCtx); err != nil {
				errOnce.Do(func() { runErr = fmt.Errorf("%s: %w", s.name, err) })
				cancel()
			}
		}(s)
	}

	// Wait for every listener to drain before returning, so the caller can tear
	// down logging knowing no handler is still running.
	wg.Wait()

	if runErr != nil {
		return runErr
	}
	slog.Info("[[.ServiceName]] shutdown complete")
	return nil
}

// serve runs one HTTP listener with the given timeouts and shuts it down when
// ctx is cancelled, waiting up to ShutdownTimeout for in-flight requests.
func serve(ctx context.Context, name, addr string, handler http.Handler, tuning config.HTTPSection) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       tuning.ReadTimeout,
		ReadHeaderTimeout: tuning.ReadHeaderTimeout,
		WriteTimeout:      tuning.WriteTimeout,
		IdleTimeout:       tuning.IdleTimeout,
		MaxHeaderBytes:    tuning.MaxHeaderBytes,
	}

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		// WithoutCancel: the grace period must outlive the already-cancelled ctx.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), tuning.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown", "server", name, "error_message", err.Error())
		}
	}()

	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		// Shutdown was requested; let the drain finish before reporting success.
		<-shutdownDone
		return nil
	}
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	<-shutdownDone
	return nil
}
