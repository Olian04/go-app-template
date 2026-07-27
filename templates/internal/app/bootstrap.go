// [[ when (modeIs "http") ]]
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"[[.ModulePath]]/internal/config"
	"[[.ModulePath]]/internal/domain/echo"
[[ if modeIs "http" ]]
	"[[.ModulePath]]/internal/observability/metrics"
[[ end ]]
	httptransport "[[.ModulePath]]/internal/transport/http"
[[ if modeIs "http" ]]
	"[[.ModulePath]]/internal/transport/metricshttp"
[[ end ]]
)

func Run(ctx context.Context, cfg config.Config) error {
	cfg = cfg.WithDefaults()
[[ if modeIs "http" ]]
	registry, err := metrics.NewRegistry(
		metrics.WithNamespace(cfg.Metrics.MetricPrefix),
		metrics.WithConstLabels(map[string]string(cfg.Labels)),
	)
	if err != nil {
		return fmt.Errorf("metrics registry: %w", err)
	}
	svc := echo.NewService()
	handler := httptransport.Router(svc, registry)

	n := 1
	metricsOn := cfg.MetricsEnabled()
	if metricsOn {
		n++
	}
	errCh := make(chan error, n)

	go func() { errCh <- serveHTTP(ctx, cfg.HTTP.ListenAddr, handler) }()

	if metricsOn {
		go func() { errCh <- metricshttp.Serve(ctx, cfg.Metrics.ListenAddr, registry.Handler()) }()
		slog.Info("[[.ServiceName]] service started", "http_addr", cfg.HTTP.ListenAddr, "metrics_addr", cfg.Metrics.ListenAddr)
	} else {
		slog.Info("[[.ServiceName]] service started", "http_addr", cfg.HTTP.ListenAddr, "metrics_enabled", false)
	}
[[ else ]]
	svc := echo.NewService()
	handler := httptransport.Router(svc)
	errCh := make(chan error, 1)
	go func() { errCh <- serveHTTP(ctx, cfg.HTTP.ListenAddr, handler) }()
	slog.Info("[[.ServiceName]] service started", "http_addr", cfg.HTTP.ListenAddr)
[[ end ]]

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func serveHTTP(ctx context.Context, addr string, handler http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("http shutdown", "error", err)
		}
	}()
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
