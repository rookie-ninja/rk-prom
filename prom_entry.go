// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkprom

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	// Why 1608? It is the year of first telescope was invented
	defaultPort = uint64(1608)
	defaultPath = "/metrics"
)

const (
	PromEntryNameDefault = "PromDefault"
	PromEntryType        = "PromEntry"
	PromEntryDescription = "Internal RK entry which implements prometheus client."
)

// Register prometheus initializer function to global RK Context which could be access via
// rk_context.GetEntry(name) whose name is rk-prom by default
func init() {
	rkentry.RegisterEntryRegFunc(RegisterPromEntriesWithConfig)
}

// Boot config which is for prom entry.
//
// 1: Path: PromEntry path, /metrics is default value.
// 2: Enabled: Enable prom entry.
// 3: Pusher.Enabled: Enable pushgateway pusher.
// 4: Pusher.IntervalMS: Interval of pushing metrics to remote pushgateway in milliseconds.
// 5: Pusher.JobName: Job name would be attached as label while pushing to remote pushgateway.
// 6: Pusher.RemoteAddress: Pushgateway address, could be form of http://x.x.x.x or x.x.x.x
// 7: Pusher.BasicAuth: Basic auth used to interact with remote pushgateway.
// 8: Pusher.Cert.Ref: Reference of rkentry.CertEntry.
// 9: Cert.Ref: Reference of rkentry.CertEntry.
type BootConfigProm struct {
	Prom struct {
		Path    string `yaml:"path" json:"path"`
		Port    uint64 `yaml:"port" json:"port"`
		Enabled bool   `yaml:"enabled" json:"enabled"`
		Pusher  struct {
			Enabled       bool   `yaml:"enabled" json:"enabled"`
			IntervalMs    int64  `yaml:"intervalMs" json:"intervalMs"`
			JobName       string `yaml:"jobName" json:"jobName"`
			RemoteAddress string `yaml:"remoteAddress" json:"remoteAddress"`
			BasicAuth     string `yaml:"basicAuth" json:"basicAuth"`
			Cert          struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"cert" json:"cert"`
		} `yaml:"pusher" json:"pusher"`
		Cert struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
		Logger struct {
			ZapLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"zapLogger" json:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"prom" json:"prom"`
}

// Prometheus entry which implements rkentry.Entry.
//
// 1: Pusher            Periodic pushGateway pusher
// 2: ZapLoggerEntry    rkentry.ZapLoggerEntry
// 3: EventLoggerEntry  rkentry.EventLoggerEntry
// 4: Port              Exposed port by prom entry
// 5: Path              Exposed path by prom entry
// 6: Registry          Prometheus registry
// 7: Registerer        Prometheus registerer
// 8: Gatherer          Prometheus gatherer
// 9: CertEntry         rkentry.CertEntry
type PromEntry struct {
	Pusher           *PushGatewayPusher        `json:"pushGatewayPusher" yaml:"pushGatewayPusher"`
	EntryName        string                    `json:"entryName" yaml:"entryName"`
	EntryType        string                    `json:"entryType" yaml:"entryType"`
	EntryDescription string                    `json:"entryDescription" yaml:"entryDescription"`
	ZapLoggerEntry   *rkentry.ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
	EventLoggerEntry *rkentry.EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
	CertEntry        *rkentry.CertEntry        `json:"certEntry" yaml:"certEntry"`
	Port             uint64                    `json:"port" yaml:"port"`
	Path             string                    `json:"path" yaml:"path"`
	Server           *http.Server              `json:"-" yaml:"-"`
	Registry         *prometheus.Registry      `json:"-" yaml:"-"`
	Registerer       prometheus.Registerer     `json:"-" yaml:"-"`
	Gatherer         prometheus.Gatherer       `json:"-" yaml:"-"`
}

// Prom entry option used while initializing prom entry via code
type PromEntryOption func(*PromEntry)

// Provide entry name
func WithName(name string) PromEntryOption {
	return func(entry *PromEntry) {
		entry.EntryName = name
	}
}

// Port of prom entry
func WithPort(port uint64) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Port = port
	}
}

// Path of prom entry
func WithPath(path string) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Path = path
	}
}

// Logger of prom entry
func WithZapLoggerEntry(zapLoggerEntry *rkentry.ZapLoggerEntry) PromEntryOption {
	return func(entry *PromEntry) {
		entry.ZapLoggerEntry = zapLoggerEntry
	}
}

// Event factory of prom entry
func WithEventLoggerEntry(eventLoggerEntry *rkentry.EventLoggerEntry) PromEntryOption {
	return func(entry *PromEntry) {
		entry.EventLoggerEntry = eventLoggerEntry
	}
}

// PushGateway of prom entry
func WithPusher(pusher *PushGatewayPusher) PromEntryOption {
	return func(entry *PromEntry) {
		entry.Pusher = pusher
	}
}

// Provide a new prometheus registry
func WithPromRegistry(registry *prometheus.Registry) PromEntryOption {
	return func(entry *PromEntry) {
		if registry != nil {
			entry.Registry = registry
		}
	}
}

// Provide cert entry
func WithCertEntry(certEntry *rkentry.CertEntry) PromEntryOption {
	return func(entry *PromEntry) {
		entry.CertEntry = certEntry
	}
}

// Create a new prom entry
// although it returns a map of prom entries, only one prom entry would be assigned to map
// the reason is for compatibility with rk_ctx.RegisterEntryInitializer
// path could be either relative or absolute directory
func RegisterPromEntriesWithConfig(configFilePath string) map[string]rkentry.Entry {
	config := &BootConfigProm{}

	rkcommon.UnmarshalBootConfig(configFilePath, config)

	res := make(map[string]rkentry.Entry)
	if config.Prom.Enabled {
		zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(config.Prom.Logger.ZapLogger.Ref)
		if zapLoggerEntry == nil {
			zapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
		}

		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(config.Prom.Logger.EventLogger.Ref)
		if eventLoggerEntry == nil {
			eventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
		}

		var pusher *PushGatewayPusher
		if config.Prom.Pusher.Enabled {
			certEntry := rkentry.GlobalAppCtx.GetCertEntry(config.Prom.Pusher.Cert.Ref)
			var certStore *rkentry.CertStore

			if certEntry != nil {
				certStore = certEntry.Store
			}

			pusher, _ = NewPushGatewayPusher(
				WithIntervalMSPusher(time.Duration(config.Prom.Pusher.IntervalMs)*time.Millisecond),
				WithRemoteAddressPusher(config.Prom.Pusher.RemoteAddress),
				WithJobNamePusher(config.Prom.Pusher.JobName),
				WithBasicAuthPusher(config.Prom.Pusher.BasicAuth),
				WithCertStorePusher(certStore),
				WithZapLoggerEntryPusher(zapLoggerEntry),
				WithEventLoggerEntryPusher(eventLoggerEntry))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(config.Prom.Cert.Ref)

		entry := RegisterPromEntry(
			WithPort(config.Prom.Port),
			WithPath(config.Prom.Path),
			WithCertEntry(certEntry),
			WithZapLoggerEntry(zapLoggerEntry),
			WithEventLoggerEntry(eventLoggerEntry),
			WithPusher(pusher))

		if entry.Pusher != nil {
			entry.Pusher.SetGatherer(entry.Gatherer)
		}

		res[entry.GetName()] = entry
	}

	return res
}

// Create a prom entry with options and add prom entry to rk_ctx.GlobalAppCtx
func RegisterPromEntry(opts ...PromEntryOption) *PromEntry {
	entry := &PromEntry{
		Port:             defaultPort,
		Path:             defaultPath,
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EntryName:        PromEntryNameDefault,
		EntryType:        PromEntryType,
		EntryDescription: PromEntryDescription,
		Registerer:       prometheus.DefaultRegisterer,
		Gatherer:         prometheus.DefaultGatherer,
	}

	for i := range opts {
		opts[i](entry)
	}

	// Trim space by default
	entry.Path = strings.TrimSpace(entry.Path)

	if len(entry.Path) < 1 {
		// Invalid path, use default one
		entry.Path = defaultPath
	}

	if !strings.HasPrefix(entry.Path, "/") {
		entry.Path = "/" + entry.Path
	}

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	if entry.Registry != nil {
		entry.Registerer = entry.Registry
		entry.Gatherer = entry.Registry
	}

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// Start prometheus client
func (entry *PromEntry) Bootstrap(context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start("bootstrap")
	defer entry.EventLoggerEntry.GetEventHelper().Finish(event)

	fields := make([]zap.Field, 0)

	fields = append(fields,
		zap.String("promPath", entry.Path),
		zap.Uint64("promPort", entry.Port))

	httpMux := http.NewServeMux()

	// if registry was provided, then use the one
	if entry.Registry != nil {
		// register process collector and go collector
		entry.Registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		entry.Registry.MustRegister(prometheus.NewGoCollector())
		httpMux.Handle(entry.Path, promhttp.HandlerFor(entry.Registry, promhttp.HandlerOpts{}))
	} else {
		httpMux.Handle(entry.Path, promhttp.Handler())
	}

	entry.Server = &http.Server{
		Addr:    "0.0.0.0:" + strconv.FormatUint(entry.Port, 10),
		Handler: httpMux,
	}

	if entry.CertEntry != nil && entry.CertEntry.Store != nil {
		if cert, err := tls.X509KeyPair(entry.CertEntry.Store.ServerCert, entry.CertEntry.Store.ServerKey); err != nil {
			rkcommon.ShutdownWithError(err)
		} else {
			entry.Server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		}
	}

	// start prom client
	entry.ZapLoggerEntry.GetLogger().Info("starting prom-client", fields...)
	entry.EventLoggerEntry.GetEventHelper().Finish(event)

	go func(*PromEntry) {
		if entry.CertEntry != nil && entry.CertEntry.Store != nil {
			if err := entry.Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				entry.ZapLoggerEntry.GetLogger().Error("error while serving prom-listener with tls", fields...)
				entry.EventLoggerEntry.GetEventHelper().FinishWithError(event, err)
				rkcommon.ShutdownWithError(err)
			}
		} else {
			if err := entry.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				entry.ZapLoggerEntry.GetLogger().Error("error while serving prom-listener", fields...)
				entry.EventLoggerEntry.GetEventHelper().FinishWithError(event, err)
				rkcommon.ShutdownWithError(err)
			}
		}
	}(entry)

	// start pusher
	if entry.Pusher != nil {
		fields = append(fields,
			zap.Bool("pusher", true),
			zap.String("remoteAddress", entry.Pusher.RemoteAddress),
			zap.String("jobName", entry.Pusher.JobName),
			zap.Int64("intervalMs", entry.Pusher.IntervalMs.Milliseconds()))
		entry.Pusher.Start()
	}

	event.AddPayloads(fields...)
}

// Shutdown prometheus client
func (entry *PromEntry) Interrupt(context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start("interrupt")

	fields := []zap.Field{
		zap.String("promPath", entry.Path),
		zap.Uint64("promPort", entry.Port),
	}

	if entry.Pusher != nil {
		fields = append(fields,
			zap.Bool("pusher", true),
			zap.String("remoteAddress", entry.Pusher.RemoteAddress),
			zap.String("jobName", entry.Pusher.JobName),
			zap.Int64("intervalMs", entry.Pusher.IntervalMs.Milliseconds()))

		entry.Pusher.Stop()
	}

	event.AddPayloads(fields...)

	if entry.Server != nil {
		entry.ZapLoggerEntry.GetLogger().Info("stopping prom-client", fields...)
		if err := entry.Server.Shutdown(context.Background()); err != nil {
			fields = append(fields, zap.Error(err))
			entry.ZapLoggerEntry.GetLogger().Warn("error occurs while stopping rk-prom-client", fields...)
		}
	}

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Return name of prom entry
func (entry *PromEntry) GetName() string {
	return entry.EntryName
}

// Return type of prom entry
func (entry *PromEntry) GetType() string {
	return entry.EntryType
}

// Stringfy prom entry
func (entry *PromEntry) String() string {
	m := map[string]interface{}{
		"entryName": entry.EntryName,
		"entryType": entry.EntryType,
		"path":      entry.Path,
		"port":      entry.Port,
	}

	if entry.Pusher != nil {
		m["pusherRemoteAddr"] = entry.Pusher.RemoteAddress
		m["pusherIntervalMs"] = entry.Pusher.IntervalMs
		m["pusherJobName"] = entry.Pusher.JobName
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}

// Get description of entry
func (entry *PromEntry) GetDescription() string {
	return entry.EntryDescription
}

// Marshal entry
func (entry *PromEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":         entry.EntryName,
		"entryType":         entry.EntryType,
		"entryDescription":  entry.EntryDescription,
		"pushGateWayPusher": entry.Pusher,
		"eventLoggerEntry":  entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":    entry.ZapLoggerEntry.GetName(),
		"port":              entry.Port,
		"path":              entry.Path,
	}

	return json.Marshal(&m)
}

// Unmarshal entry
func (entry *PromEntry) UnmarshalJSON(b []byte) error {
	return nil
}

// Register collectors
func (entry *PromEntry) RegisterCollectors(collectors ...prometheus.Collector) error {
	var err error
	for i := range collectors {
		if innerErr := entry.Registerer.Register(collectors[i]); innerErr != nil {
			err = innerErr
		}
	}

	return err
}
