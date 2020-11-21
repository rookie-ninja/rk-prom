// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"os"
)

var (
	// options
	ProcessCollectorOpts = prometheus.ProcessCollectorOpts{
		PidFn:        func() (int, error) { return os.Getpid(), nil },
		Namespace:    "rk",
		ReportErrors: false,
	}
	// collectors
	ProcessCollector = prometheus.NewProcessCollector(ProcessCollectorOpts)
)
