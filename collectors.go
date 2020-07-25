package rk_prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"os"
)

var (
	// options
	ProcessCollectorOpts = prometheus.ProcessCollectorOpts{
		PidFn:        func() (int, error) { return os.Getpid(), nil },
		Namespace:    "pl",
		ReportErrors: false,
	}
	// collectors
	ProcessCollector = prometheus.NewProcessCollector(ProcessCollectorOpts)
)
