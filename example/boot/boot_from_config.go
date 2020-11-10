// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"log"
	"time"
)

func main() {
	fac := rk_query.NewEventFactory()
	entry := rk_prom.NewPromEntryWithConfig("example/boot/boot.yaml", fac, rk_logger.StdoutLogger)
	entry.Bootstrap(fac.CreateEvent())

	// with custom namespace and subsystem
	metricsSet := rk_prom.NewMetricsSet("new_namespace", "new_service")

	tryCounters(metricsSet)
	tryGauges(metricsSet)
	trySummary(metricsSet)
	tryHistogram(metricsSet)

	entry.Wait(1 * time.Second)
}

func tryCounters(metricsSet *rk_prom.MetricsSet) {
	err := metricsSet.RegisterCounter("counter", "key_1")
	if err != nil {
		log.Fatal(err)
	}

	metricsSet.GetCounterWithValues("counter", "value_1").Inc()
	metricsSet.GetCounterWithLabels("counter", prometheus.Labels{"key_1": "value_1"}).Inc()
}

func tryGauges(metricsSet *rk_prom.MetricsSet) {
	err := metricsSet.RegisterGauge("gauge", "key_1")
	if err != nil {
		log.Fatal(err)
	}

	metricsSet.GetGaugeWithValues("gauge", "value_1").Inc()
	metricsSet.GetGaugeWithLabels("gauge", prometheus.Labels{"key_1": "value_1"}).Inc()
}

func trySummary(metricsSet *rk_prom.MetricsSet) {
	err := metricsSet.RegisterSummary("summary", rk_prom.SummaryObjectives, "key_1")
	if err != nil {
		log.Fatal(err)
	}

	metricsSet.GetSummaryWithValues("summary", "value_1").Observe(1.0)
	metricsSet.GetSummaryWithLabels("summary", prometheus.Labels{"key_1": "value_1"}).Observe(1.0)
}

func tryHistogram(metricsSet *rk_prom.MetricsSet) {
	err := metricsSet.RegisterHistogram("histogram", []float64{}, "key_1")
	if err != nil {
		log.Fatal(err)
	}

	metricsSet.GetHistogramWithValues("histogram", "value_1").Observe(1.0)
	metricsSet.GetHistogramWithLabels("histogram", prometheus.Labels{"key_1": "value_1"}).Observe(1.0)
}
