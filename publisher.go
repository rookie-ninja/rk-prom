// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_prom

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	rk_logger "github.com/rookie-ninja/rk-logger"
	"go.uber.org/zap"
	"sync"
	"time"
)

type PushGatewayPusher struct {
	logger    *zap.Logger
	pusher    *push.Pusher
	interval  time.Duration
	url       string
	jobName   string
	isRunning bool
	lock      *sync.Mutex
}

func NewPushGatewayPublisher(interval time.Duration, url, jobName string, logger *zap.Logger) (*PushGatewayPusher, error) {
	if interval < 1 {
		return nil, errors.New("negative interval")
	}

	if len(url) < 1 {
		return nil, errors.New("empty url")
	}

	if len(jobName) < 1 {
		return nil, errors.New("empty job name")
	}

	if logger == nil {
		logger = rk_logger.StdoutLogger
	}

	return &PushGatewayPusher{
		logger:    logger,
		interval:  interval,
		pusher:    push.New(url, jobName),
		url:       url,
		jobName:   jobName,
		lock:      &sync.Mutex{},
		isRunning: false,
	}, nil
}

func (pub *PushGatewayPusher) Start() {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.isRunning = true

	pub.logger.Info("starting pushGateway publisher,",
		zap.String("remote_addr", pub.url),
		zap.String("job_name", pub.jobName))

	go pub.publish()
}

func (pub *PushGatewayPusher) publish() {
	for pub.isRunning {
		err := pub.pusher.Push()

		if err != nil {
			pub.logger.Debug("could not push metrics to PushGateway", zap.Error(err))
		}

		time.Sleep(pub.interval)
	}
}

func (pub *PushGatewayPusher) IsRunning() bool {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	return pub.isRunning
}

func (pub *PushGatewayPusher) Shutdown() {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	if !pub.isRunning {
		return
	}

	pub.isRunning = false
}

// Simply call pusher.Gatherer()
// We add prefix "Add" before the function name since the original one is a little bit confusing.
// Thread safe
func (pub *PushGatewayPusher) AddGatherer(g prometheus.Gatherer) *PushGatewayPusher {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.pusher.Gatherer(g)
	return pub
}

// Simply call pusher.Collector()
// We add prefix "Add" before the function name since the original one is a little bit confusing.
// Thread safe
func (pub *PushGatewayPusher) AddCollector(c prometheus.Collector) *PushGatewayPusher {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.pusher.Collector(c)
	return pub
}

// Simply call pusher.Grouping()
// We add prefix "Add" before the function name since the original one is a little bit confusing.
// Thread safe
func (pub *PushGatewayPusher) AddGrouping(name, value string) *PushGatewayPusher {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.pusher.Grouping(name, value)
	return pub
}

// Simply call pusher.BasicAuth()
// We add prefix "Set" before the function name since the original one is a little bit confusing.
// Thread safe
func (pub *PushGatewayPusher) SetBasicAuth(username, password string) *PushGatewayPusher {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.pusher.BasicAuth(username, password)
	return pub
}

// Simply call pusher.Format()
// We add prefix "Set" before the function name since the original one is a little bit confusing.
// Thread safe
func (pub *PushGatewayPusher) SetFormat(format expfmt.Format) *PushGatewayPusher {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.pusher.Format(format)
	return pub
}

// Simply call pusher.Client()
// We add prefix "Set" before the function name since the original one is a little bit confusing.
// Thread safe
func (pub *PushGatewayPusher) SetClient(c push.HTTPDoer) *PushGatewayPusher {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	pub.pusher.Client(c)
	return pub
}

// Push collects/gathers all metrics from all Collectors and Gatherers added to
// this Pusher. Then, it pushes them to the PushGateway configured while
// creating this Pusher, using the configured job name and any added grouping
// labels as grouping key. All previously pushed metrics with the same job and
// other grouping labels will be replaced with the metrics pushed by this
// call. (It uses HTTP method “PUT” to push to the PushGateway.)
//
// Push returns the first error encountered by any method call (including this
// one) in the lifetime of the Pusher.
// Thread safe
func (pub *PushGatewayPusher) PUT() error {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	return pub.pusher.Push()
}

// Add works like push, but only previously pushed metrics with the same name
// (and the same job and other grouping labels) will be replaced. (It uses HTTP
// method “POST” to push to the PushGateway.)
// Thread safe
func (pub *PushGatewayPusher) POST() error {
	pub.lock.Lock()
	defer pub.lock.Unlock()

	return pub.pusher.Add()
}
