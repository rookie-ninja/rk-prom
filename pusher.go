// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkprom

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rookie-ninja/rk-entry/entry"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"sync"
	"time"
)

// a pusher which contains bellow instances
// thread safe
//
// 1: logger:          zap logger for logging periodic job information
// 2: pusher:          prometheus pusher which will push metrics to remote pushGateway
// 3: intervalMS:      periodic job interval in milliseconds
// 4: remoteAddress:   remote pushGateway URL. You can use just host:port or ip:port as url,
//                     in which case “http://” is added automatically. Alternatively, include the
//                     schema in the URL. However, do not include the “/metrics/jobs/…” part.
// 5: jobName:         job name of periodic job
// 6: isRunning:       a boolean flag for validating status of periodic job
// 7: lock:            a mutex lock for thread safety
// 8: credential:      basic auth credential
type PushGatewayPusher struct {
	ZapLoggerEntry   *rkentry.ZapLoggerEntry
	EventLoggerEntry *rkentry.EventLoggerEntry
	CertStore        *rkentry.CertStore
	Pusher           *push.Pusher
	IntervalMS       time.Duration
	RemoteAddress    string
	JobName          string
	Running          *atomic.Bool
	lock             *sync.Mutex
	Credential       string
}

type PushGatewayPusherOption func(*PushGatewayPusher)

func WithIntervalMSPusher(intervalMS time.Duration) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.IntervalMS = intervalMS
	}
}

func WithRemoteAddressPusher(remoteAddress string) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.RemoteAddress = remoteAddress
	}
}

func WithJobNamePusher(jobName string) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.JobName = jobName
	}
}

func WithBasicAuthPusher(cred string) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.Credential = cred
	}
}

func WithZapLoggerEntryPusher(zapLoggerEntry *rkentry.ZapLoggerEntry) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.ZapLoggerEntry = zapLoggerEntry
	}
}

func WithEventLoggerEntryPusher(eventLoggerEntry *rkentry.EventLoggerEntry) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.EventLoggerEntry = eventLoggerEntry
	}
}

func WithCertStorePusher(certStore *rkentry.CertStore) PushGatewayPusherOption {
	return func(pusher *PushGatewayPusher) {
		pusher.CertStore = certStore
	}
}

// create a new pushGateway periodic job instances with intervalMS, remote URL and job name
// 1: intervalMS: should be a positive integer
// 2: url:        should be a non empty and valid url
// 3: jabName:    should be a non empty string
// 4: cred:       credential of basic auth format as user:pass
// 5: logger:     a logger with stdout output would be assigned if nil
func NewPushGatewayPusher(opts ...PushGatewayPusherOption) (*PushGatewayPusher, error) {
	pg := &PushGatewayPusher{
		ZapLoggerEntry:   rkentry.GlobalAppCtx.GetZapLoggerEntryDefault(),
		EventLoggerEntry: rkentry.GlobalAppCtx.GetEventLoggerEntryDefault(),
		IntervalMS:       1 * time.Second,
		lock:             &sync.Mutex{},
		Running:          atomic.NewBool(false),
	}

	for i := range opts {
		opts[i](pg)
	}

	if pg.IntervalMS < 1 {
		return nil, errors.New("invalid intervalMS")
	}

	if len(pg.RemoteAddress) < 1 {
		return nil, errors.New("empty remoteAddress")
	}

	// certificate was provided, we need to use https for remote address
	if pg.CertStore != nil {
		if !strings.HasPrefix(pg.RemoteAddress, "https://") {
			pg.RemoteAddress = "https://" + pg.RemoteAddress
		}
	}

	if len(pg.JobName) < 1 {
		return nil, errors.New("empty job name")
	}

	if pg.ZapLoggerEntry == nil {
		pg.ZapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if pg.EventLoggerEntry == nil {
		pg.EventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	pg.Pusher = push.New(pg.RemoteAddress, pg.JobName)

	// assign credential of basic auth
	if len(pg.Credential) > 0 && strings.Contains(pg.Credential, ":") {
		pg.Credential = strings.TrimSpace(pg.Credential)
		tokens := strings.Split(pg.Credential, ":")
		if len(tokens) == 2 {
			pg.Pusher = pg.Pusher.BasicAuth(tokens[0], tokens[1])
		}
	}

	httpClient := &http.Client{
		Timeout: rkentry.DefaultTimeout,
	}

	// deal with tls
	if pg.CertStore != nil {
		certPool := x509.NewCertPool()

		certPool.AppendCertsFromPEM(pg.CertStore.ServerCert)

		conf := &tls.Config{RootCAs: certPool}

		cert, err := tls.X509KeyPair(pg.CertStore.ClientCert, pg.CertStore.ClientKey)

		if err == nil {
			conf.Certificates = []tls.Certificate{cert}
		}

		httpClient.Transport = &http.Transport{TLSClientConfig: conf}
	}

	pg.Pusher.Client(httpClient)

	return pg, nil
}

// start a periodic job
func (pub *PushGatewayPusher) Start() {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	// periodic job already started
	// caution, do not call pub.isRunning() function directory, since it will cause dead lock
	if pub.Running.Load() {
		pub.ZapLoggerEntry.GetLogger().Info("pushGateway publisher already started",
			zap.String("remote_address", pub.RemoteAddress),
			zap.String("job_name", pub.JobName))
		return
	}

	pub.Running.CAS(false, true)

	pub.ZapLoggerEntry.GetLogger().Info("starting pushGateway publisher",
		zap.String("remote_address", pub.RemoteAddress),
		zap.String("job_name", pub.JobName))

	go pub.push()
}

// internal use only
func (pub *PushGatewayPusher) push() {
	for pub.Running.Load() {
		event := pub.EventLoggerEntry.GetEventHelper().Start("publish")
		event.AddFields(
			zap.String("job_name", pub.JobName),
			zap.String("remote_addr", pub.RemoteAddress),
			zap.Duration("interval_ms", pub.IntervalMS))

		err := pub.Pusher.Push()

		if err != nil {
			pub.ZapLoggerEntry.GetLogger().Warn("failed to push metrics to PushGateway",
				zap.String("remote_address", pub.RemoteAddress),
				zap.String("job_name", pub.JobName),
				zap.Error(err))
			pub.EventLoggerEntry.GetEventHelper().FinishWithError(event, err)
		} else {
			pub.EventLoggerEntry.GetEventHelper().Finish(event)
		}

		time.Sleep(pub.IntervalMS)
	}
}

// validate whether periodic job is running or not
func (pub *PushGatewayPusher) IsRunning() bool {
	return pub.Running.Load()
}

// stop periodic job
func (pub *PushGatewayPusher) Stop() {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.Running.CAS(true, false)
}

// Simply call pusher.Gatherer()
// We add prefix "Add" before the function name since the original one is a little bit confusing.
// Thread safe
func (pub *PushGatewayPusher) GetPusher() *push.Pusher {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	return pub.Pusher
}

func (pub *PushGatewayPusher) String() string {
	bytes, err := json.Marshal(pub)
	if err != nil {
		// failed to marshal, just return empty string
		return ""
	}

	return string(bytes)
}

func (pub *PushGatewayPusher) SetGatherer(gatherer prometheus.Gatherer) {
	if pub.Pusher != nil {
		pub.Pusher.Gatherer(gatherer)
	}
}
