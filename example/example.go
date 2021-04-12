// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-prom"
	"log"
)

func main() {
	startFromConfig()
	startFromCode()
}

func startFromCode() {
	// create prom entry
	entry := rkprom.RegisterPromEntry(
		rkprom.WithPort(1608),
		rkprom.WithPath("metrics"),
		rkprom.WithPromRegistry(prometheus.NewRegistry()))

	// start server
	entry.Bootstrap(context.TODO())

	metricsSet := rkprom.NewMetricsSet("new_namespace", "new_service", entry.Registerer)

	tryCounters(metricsSet)
	tryGauges(metricsSet)
	trySummary(metricsSet)
	tryHistogram(metricsSet)

	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// stop server
	entry.Interrupt(context.TODO())
}

func startFromConfig() {
	rkentry.RegisterBasicEntriesFromConfig("example/boot.yaml")

	maps := rkprom.RegisterPromEntriesWithConfig("example/boot.yaml")

	entry := maps[rkprom.PromEntryNameDefault]
	entry.Bootstrap(context.TODO())

	// with custom namespace and subsystem
	metricsSet := rkprom.NewMetricsSet("new_namespace", "new_service", entry.(*rkprom.PromEntry).Registerer)

	tryCounters(metricsSet)
	tryGauges(metricsSet)
	trySummary(metricsSet)
	tryHistogram(metricsSet)

	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// stop server
	entry.Interrupt(context.TODO())
}

func tryCounters(metricsSet *rkprom.MetricsSet) {
	err := metricsSet.RegisterCounter("counter", "key_1", "instance")
	if err != nil {
		log.Fatal(err)
	}

	metricsSet.GetCounterWithValues("counter", "value_1", "localhost").Inc()
	metricsSet.GetCounterWithLabels("counter", prometheus.Labels{"key_1": "value_1", "instance": "localhost"}).Inc()
}

func tryGauges(metricsSet *rkprom.MetricsSet) {
	err := metricsSet.RegisterGauge("gauge", "key_1")
	if err != nil {
		log.Fatal(err)
	}

	metricsSet.GetGaugeWithValues("gauge", "value_1").Inc()
	metricsSet.GetGaugeWithLabels("gauge", prometheus.Labels{"key_1": "value_1"}).Inc()
}

func trySummary(metricsSet *rkprom.MetricsSet) {
	err := metricsSet.RegisterSummary("summary", rkprom.SummaryObjectives, "key_1")
	if err != nil {
		log.Fatal(err)
	}

	metricsSet.GetSummaryWithValues("summary", "value_1").Observe(1.0)
	metricsSet.GetSummaryWithLabels("summary", prometheus.Labels{"key_1": "value_1"}).Observe(1.0)
}

func tryHistogram(metricsSet *rkprom.MetricsSet) {
	err := metricsSet.RegisterHistogram("histogram", []float64{}, "key_1")
	if err != nil {
		log.Fatal(err)
	}

	metricsSet.GetHistogramWithValues("histogram", "value_1").Observe(1.0)
	metricsSet.GetHistogramWithLabels("histogram", prometheus.Labels{"key_1": "value_1"}).Observe(1.0)
}
