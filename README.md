# Overview
A simple prometheus initializer.
What rk-prom trying to do is described as bellow:
- Start prometheus client by calling StartProm()
- Start a daemon thread which will periodically push local prometheus metrics to PushGateway
- Simple wrapper of Counter, Gauge, Summary, Histogram like POJO with GetXXX(), RegisterXXX(), UnRegisterXXX()
- Go & Process collector variables which is originally implemented by prometheus client package.

## Installation
`go get -u github.com/rookie-ninja/rk-prom`

## Development Status: Active
In **Prod** version. 

## Quick start
```go
// With Port and Path
server := rk_prom.StartProm("1608", "/metrics")

// Without Port and Path
server := rk_prom.StartProm("", "")
// Default port and path would be assigned
var (
	// Why 1608? It is the year of first telescope was invented
	DefaultPort = "1608"
	DefaultPath = "/metrics"
)
```

## Example
- Start Prometheus Service
```go
// With Port and Path
server := rk_prom.StartProm("1984", "/metrics")

// Default port and path would be assigned
var (
	// Why 1608? It is the year of first telescope was invented
	DefaultPort = "1608"
	DefaultPath = "/metrics"
)
```

- Working with Counter (namespace and subsystem)
```go
metricsSet := rk_prom.NewMetricsSet("my_namespace", "my_service")

metricsSet.RegisterCounter("counter", "key_1")

metricsSet.GetCounterWithValues("counter", "value_1").Inc()
metricsSet.GetCounterWithLabels("counter", prometheus.Labels{"key_1":"value_1"}).Inc()
```

- Working with Gauge (namespace and subsystem)
```go
metricsSet := rk_prom.NewMetricsSet("my_namespace", "my_service")
metricsSet.RegisterGauge("gauge", "key_1")

metricsSet.GetGaugeWithValues("gauge", "value_1").Inc()
metricsSet.GetGaugeWithLabels("gauge", prometheus.Labels{"key_1":"value_1"}).Inc()
```

- Working with Summary (custom namespace and subsystem)
```go
metricsSet := rk_prom.NewMetricsSet("my_namespace", "my_service")
metricsSet.RegisterSummary("summary", rk_prom.SummaryObjectives, "key_1")

metricsSet.GetSummaryWithValues("summary", "value_1").Observe(1.0)
metricsSet.GetSummaryWithLabels("summary", prometheus.Labels{"key_1":"value_1"}).Observe(1.0)
```

- Working with Histogram (custom namespace and subsystem)
```go
metricsSet := rk_prom.NewMetricsSet("new_namespace", "new_service")
metricsSet.RegisterHistogram("histogram", []float64{}, "key_1")

metricsSet.GetHistogramWithValues("histogram", "value_1").Observe(1.0)
metricsSet.GetHistogramWithLabels("histogram", prometheus.Labels{"key_1":"value_1"}).Observe(1.0)
```

- Working with PushGateway publisher
```go
pub := rk_prom.NewPushGatewayPublisher(2 * time.Second, "localhost:8888", "test_job")
pub.Start()
defer pub.Shutdown()

time.Sleep(2 * time.Second)
```

## Contributing
We encourage and support an active, healthy community of contributors — including you!
Details are in the [contribution guide](/CONTRIBUTING.md) and the [code of conduct](/CODE_OF_CONDUCT.md). The pulse-line maintainers keep an eye on issues and pull requests. So don't hesitate to hold us to a high standard.
