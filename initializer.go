// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

var (
	logger = zap.NewNop()
	// Why 1608? It is the year of first telescope was invented
	defaultPort = uint64(1608)
	defaultPath = "/metrics"
)

// Register collectors in default registry
func RegisterCollectors(c ...prometheus.Collector) {
	prometheus.MustRegister(c...)
}

func StartProm(port uint64, path string) (*http.Server, error) {
	// Trim space by default
	path = strings.TrimSpace(path)

	if len(path) < 1 {
		// Invalid path, use default one
		logger.Warn("invalid path, using default path",
			zap.String("path", defaultPath))
		path = defaultPath
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Register by default
	err := prometheus.Register(ProcessCollector)
	if err != nil {
		logger.Error("failed to register collector",
			zap.Error(err))
		return nil, err
	}

	httpMux := http.NewServeMux()
	httpMux.Handle(path, promhttp.Handler())

	server := &http.Server{
		Addr:    "0.0.0.0:" + strconv.FormatUint(port, 10),
		Handler: httpMux,
	}

	go func() {
		logger.Info("starting prometheus client,",
			zap.Uint64("port", port),
			zap.String("path", path))

		server.ListenAndServe()
	}()

	return server, err
}

func SetZapLogger(in *zap.Logger) {
	if in != nil {
		logger = in
	}
}
