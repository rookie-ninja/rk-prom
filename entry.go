// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_prom

import (
	"context"
	"encoding/json"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/context"
	rk_entry "github.com/rookie-ninja/rk-common/entry"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var (
	// Why 1608? It is the year of first telescope was invented
	defaultPort = uint64(1608)
	defaultPath = "/metrics"
)

const PromEntryNameDefault = "rk-prom"

func init() {
	rk_ctx.RegisterEntryInitializer(NewPromEntries)
}

type bootConfig struct {
	Prom struct {
		Path    string `yaml:"path"`
		Port    uint64 `yaml:"port"`
		Enabled bool   `yaml:"enabled"`
		Pusher  struct {
			Enabled  bool   `yaml:"enabled"`
			Interval int64  `yaml:"interval"`
			Job      string `yaml:"job"`
			URL      string `yaml:"url"`
		} `yaml:"pusher"`
	} `yaml:"prom"`
}

type PromEntry struct {
	pusher         *PushGatewayPusher
	name           string
	entryType      string
	logger         *zap.Logger
	factory        *rk_query.EventFactory
	port           uint64
	path           string
	enablePusher   bool
	pusherInterval int64
	pusherJob      string
	pusherURL      string
	server         *http.Server
}

type PromEntryOption func(*PromEntry)

func WithPort(port uint64) PromEntryOption {
	return func(entry *PromEntry) {
		entry.port = port
	}
}

func WithPath(path string) PromEntryOption {
	return func(entry *PromEntry) {
		entry.path = path
	}
}

func WithLogger(logger *zap.Logger) PromEntryOption {
	return func(entry *PromEntry) {
		entry.logger = logger
	}
}

func WithEventFactory(fac *rk_query.EventFactory) PromEntryOption {
	return func(entry *PromEntry) {
		entry.factory = fac
	}
}

func WithEnablePusher(enable bool) PromEntryOption {
	return func(entry *PromEntry) {
		entry.enablePusher = enable
	}
}

func WithPusherInterval(interval int64) PromEntryOption {
	return func(entry *PromEntry) {
		entry.pusherInterval = interval
	}
}

func WithPusherJob(job string) PromEntryOption {
	return func(entry *PromEntry) {
		entry.pusherJob = job
	}
}

func WithPusherURL(url string) PromEntryOption {
	return func(entry *PromEntry) {
		entry.pusherURL = url
	}
}

func NewPromEntries(path string, factory *rk_query.EventFactory, logger *zap.Logger) map[string]rk_entry.Entry {
	bytes := readFile(path)
	config := &bootConfig{}
	if err := yaml.Unmarshal(bytes, config); err != nil {
		shutdownWithError(err)
		return nil
	}

	return getPromServerEntries(config, factory, logger)
}

func getPromServerEntries(config *bootConfig, factory *rk_query.EventFactory, logger *zap.Logger) map[string]rk_entry.Entry {
	res := make(map[string]rk_entry.Entry)
	if config.Prom.Enabled {
		entry := NewPromEntry(
			WithPort(config.Prom.Port),
			WithPath(config.Prom.Path),
			WithLogger(logger),
			WithEventFactory(factory),
			WithEnablePusher(config.Prom.Pusher.Enabled),
			WithPusherInterval(config.Prom.Pusher.Interval),
			WithPusherJob(config.Prom.Pusher.Job),
			WithPusherURL(config.Prom.Pusher.URL))
		res[entry.GetName()] = entry
	}

	return res
}

func NewPromEntry(opts ...PromEntryOption) *PromEntry {
	entry := &PromEntry{
		port:      defaultPort,
		path:      defaultPath,
		name:      PromEntryNameDefault,
		entryType: "prom",
	}

	for i := range opts {
		opts[i](entry)
	}

	rk_ctx.GlobalAppCtx.AddEntry(entry.GetName(), entry)

	return entry
}

func (entry *PromEntry) Bootstrap(event rk_query.Event) {
	fields := make([]zap.Field, 0)

	// Trim space by default
	entry.path = strings.TrimSpace(entry.path)

	if len(entry.path) < 1 {
		// Invalid path, use default one
		entry.logger.Info("invalid path, using default path",
			zap.String("path", defaultPath))
		entry.path = defaultPath
	}

	if !strings.HasPrefix(entry.path, "/") {
		entry.path = "/" + entry.path
	}

	fields = append(fields,
		zap.String("prom_path", entry.path),
		zap.Uint64("prom_port", entry.port))

	// Register by default
	err := prometheus.Register(ProcessCollector)
	if err != nil {
		fields := append(fields, zap.Error(err))
		entry.logger.Error("failed to register collector", fields...)
		shutdownWithError(err)
	}

	httpMux := http.NewServeMux()
	httpMux.Handle(entry.path, promhttp.Handler())

	entry.server = &http.Server{
		Addr:    "0.0.0.0:" + strconv.FormatUint(entry.port, 10),
		Handler: httpMux,
	}

	// start pusher
	if entry.enablePusher {
		fields = append(fields,
			zap.Bool("pusher", true),
			zap.String("pusher_url", entry.pusherURL),
			zap.String("pusher_job", entry.pusherJob),
			zap.Int64("pusher_interval", entry.pusherInterval))

		entry.pusher, err = NewPushGatewayPublisher(
			time.Duration(entry.pusherInterval)*time.Second,
			entry.pusherURL,
			entry.pusherJob,
			entry.logger)
		if err != nil {
			shutdownWithError(err)
		}
		entry.pusher.Start()
	}

	event.AddFields(fields...)

	// start prom client
	rk_ctx.GlobalAppCtx.GetDefaultLogger().Info("starting prom-server", fields...)
	go func(*PromEntry) {
		if err := entry.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			rk_ctx.GlobalAppCtx.GetDefaultLogger().Error("error while serving prom-listener", fields...)
			shutdownWithError(err)
		}
	}(entry)
}

func (entry *PromEntry) Shutdown(event rk_query.Event) {
	fields := []zap.Field{
		zap.Uint64("prom_port", entry.port),
		zap.String("prom_name", entry.name),
	}

	if entry.enablePusher {
		fields = append(fields,
			zap.Bool("pusher", true),
			zap.String("pusher_url", entry.pusherURL),
			zap.String("pusher_job", entry.pusherJob),
			zap.Int64("pusher_interval", entry.pusherInterval))

		if entry.pusher != nil {
			entry.pusher.Shutdown()
		}
	}

	event.AddFields(fields...)

	if entry.server != nil {
		rk_ctx.GlobalAppCtx.GetDefaultLogger().Info("stopping prom-server", fields...)
		if err := entry.server.Shutdown(context.Background()); err != nil {
			fields = append(fields, zap.Error(err))
			rk_ctx.GlobalAppCtx.GetDefaultLogger().Warn("error occurs while stopping prom-server", fields...)
		}
	}
}

func (entry *PromEntry) GetName() string {
	return entry.name
}

func (entry *PromEntry) GetType() string {
	return entry.entryType
}

func (entry *PromEntry) Wait(draining time.Duration) {
	sig := <-rk_ctx.GlobalAppCtx.GetShutdownSig()

	helper := rk_query.NewEventHelper(rk_ctx.GlobalAppCtx.GetEventFactory())
	event := helper.Start("rk_app_stop")

	rk_ctx.GlobalAppCtx.GetDefaultLogger().Info("draining", zap.Duration("draining_duration", draining))
	time.Sleep(draining)

	event.AddFields(
		zap.Duration("app_lifetime_nano", time.Since(rk_ctx.GlobalAppCtx.GetStartTime())),
		zap.Time("app_start_time", rk_ctx.GlobalAppCtx.GetStartTime()))

	event.AddPair("signal", sig.String())

	entry.Shutdown(event)

	helper.Finish(event)
}

func (entry *PromEntry) String() string {
	m := map[string]string{
		"name": entry.GetName(),
		"type": entry.GetType(),
		"port": strconv.FormatUint(entry.GetPort(), 10),
		"path": entry.GetPath(),
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}

func (entry *PromEntry) GetPort() uint64 {
	return entry.port
}

func (entry *PromEntry) GetPath() string {
	return entry.path
}

func (entry *PromEntry) GetPusherURL() string {
	return entry.pusherURL
}

// Register collectors in default registry
func RegisterCollectors(c ...prometheus.Collector) {
	prometheus.MustRegister(c...)
}

func shutdownWithError(err error) {
	debug.PrintStack()
	glog.Error(err)
	os.Exit(1)
}

func readFile(filePath string) []byte {
	if !path.IsAbs(filePath) {
		wd, err := os.Getwd()

		if err != nil {
			shutdownWithError(err)
		}
		filePath = path.Join(wd, filePath)
	}

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		shutdownWithError(err)
	}

	return bytes
}
