package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-prom"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {
	logger, _ := zap.NewDevelopment()
	rk_prom.SetZapLogger(logger)

	// start prom on local
	server := rk_prom.StartProm("1608", "/metrics")

	defer func() {
		err := server.Shutdown(context.Background())
		log.Fatal(err)
	}()

	// with custom namespace and subsystem
	metricsSet := rk_prom.NewMetricsSet("new_namespace", "new_service")

	tryCounters(metricsSet)
	tryGauges(metricsSet)
	trySummary(metricsSet)
	tryHistogram(metricsSet)

	time.Sleep(20 * time.Second)
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

func tryPublisher() {
	pub, _ := rk_prom.NewPushGatewayPublisher(2 * time.Second, "localhost:8888", "test_job")
	pub.Start()

	defer pub.Shutdown()
	time.Sleep(2 * time.Second)
}