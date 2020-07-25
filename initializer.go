package rk_prom

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

var (
	logger = zap.NewNop()
	// Why 1608? It is the year of first telescope was invented
	DefaultPort = "1608"
	DefaultPath = "/metrics"
)

// Register collectors in default registry
func RegisterCollectors(c ...prometheus.Collector) {
	prometheus.MustRegister(c...)
}

func StartProm(port, path string) *http.Server {
	// Trim space by default
	port = strings.TrimSpace(port)
	path = strings.TrimSpace(path)

	if len(port) < 1 {
		logger.Warn(fmt.Sprintf("port is empty, using default port:%s", DefaultPort))
		port = DefaultPort
	}

	if len(path) < 1 || !strings.HasPrefix(path, "/") {
		// Invalid path, use default one
		logger.Warn(fmt.Sprintf("invalid path, using default path:%s", DefaultPath))
		path = DefaultPath
	}

	// Register by default
	err := prometheus.Register(ProcessCollector)
	if err != nil {
		logger.Warn(fmt.Sprintf("failed to register collector, %v", err))
	}

	httpMux := http.NewServeMux()
	httpMux.Handle(path, promhttp.Handler())

	server := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: httpMux,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Error(fmt.Sprintf("failed to serving prometheus client, %v", err))
		}
	}()

	return server
}

func SetZapLogger(in *zap.Logger) {
	if in != nil {
		logger = in
	}
}
