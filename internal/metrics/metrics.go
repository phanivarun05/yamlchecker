package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/phanivarun05/yamlchecker/checker"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	endpointUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "yamlchecker_endpoint_up",
		Help: "Current status of the Endpoint",
	}, []string{"name"})
)
var (
	checksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "yamlchecker_checks_total",
		Help: "The total number of checks done",
	}, []string{"name", "result"})
)
var (
	checkDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "yamlchecker_check_duration_seconds",
		Help:    "Duration describing how long checks took",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
	}, []string{"name"})
)

func Init(targets []checker.Target) {
	outcomes := []string{checker.OutcomeHealthy, checker.OutcomeUnhealthy, checker.OutcomeUnreachable}
	for _, target := range targets {
		for _, outcome := range outcomes {
			checksTotal.WithLabelValues(target.Name, outcome)
		}
	}
}

func Record(name, result string, d time.Duration) {
	var statusVal float64 = 0
	if result == "healthy" {
		statusVal = 1
	}
	endpointUp.WithLabelValues(name).Set(statusVal)
	checksTotal.WithLabelValues(name, result).Inc()
	checkDuration.WithLabelValues(name).Observe(d.Seconds())
}

func Serve(ctx context.Context, addr string) (<-chan struct{}, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen at %s: %w", addr, err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		err := server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			slog.Error("failed to serve the port", "err", err)
		}
	}()
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-ctx.Done()
		slog.Info("Shutting down the metrics server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown with error", "err", err)
		} else {
			slog.Info("server exited gracefully")
		}
	}()
	slog.Info("metrics server listening", "address", addr)
	return done, nil
}
