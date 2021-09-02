# rk-prom
[![build](https://github.com/rookie-ninja/rk-prom/actions/workflows/ci.yml/badge.svg)](https://github.com/rookie-ninja/rk-prom/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/rookie-ninja/rk-prom/branch/master/graph/badge.svg?token=7SMMXOLMJQ)](https://codecov.io/gh/rookie-ninja/rk-prom)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A simple prometheus initializer.
What rk-prom trying to do is described as bellow:
- Start prometheus client by calling StartProm()
- Start prometheus client by providing yaml config
- Start a daemon thread which will periodically push local prometheus metrics to PushGateway
- Simple wrapper of Counter, Gauge, Summary, Histogram like POJO with GetXXX(), RegisterXXX(), UnRegisterXXX()
- Go & Process collector variables which is originally implemented by prometheus client package.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installation](#installation)
- [Development Status: Active](#development-status-active)
- [Quick start](#quick-start)
- [Example](#example)
- [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation
`go get -u github.com/rookie-ninja/rk-prom`

## Development Status: Active
In **Prod** version. 

## Quick start
Start prom client with StartProm()

```go
// With Port and Path
server := rk_prom.StartProm("1608", "/metrics")

// Without Port and Path
// Default port and path would be assigned
// DefaultPort = "1608"
// DefaultPath = "/metrics"
// server := rk_prom.StartProm("", "")
```

Start with Bootstrap() with config file
```yaml
---
prom:
  enabled: true
  port: 1608
  path: metrics
  pusher:
    enabled: false
    intervalMS: 1
    jobName: "rk-job"
    remoteAddress: "localhost:9091"
    basicAuth: "user:pass"
```

```go
package main

import (
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"time"
)

func main() {
    // create event data
    fac := rk_query.NewEventFactory()

    // create prom entry
    entry := rk_prom.NewPromEntry(
        rk_prom.WithPort(1608),
        rk_prom.WithPath("metrics"))

    // start server
    entry.Bootstrap(fac.CreateEvent())
    entry.WaitForShutdownSig(1 * time.Second)
}
```

| Name | Description | Option | Default Value |
| ------ | ------ | ------ | ------ |
| prom.enabled | Enable prometheus | bool | false |
| prom.port | Prometheus port | integer | 1608 |
| prom.path | Prometheus path | string | metrics |
| prom.pusher.enabled | Enable push gateway pusher | bool | false |
| prom.pusher.intervalMS | Push interval to remote push gateway | integer | 0 |
| prom.pusher.jobName | Pusher job name | string | empty string |
| prom.pusher.remoteAddress | Pusher url | string | empty string |
| prom.pusher.basicAuth | basic auth as user:password | string | empty string |


Start with Bootstrap() with code
```go
package main

import (
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"time"
)

func main() {
    fac := rk_query.NewEventFactory()
    maps := rk_prom.NewPromEntryWithConfig("example/boot/boot.yaml", fac, rk_logger.StdoutLogger)
    entry := maps[rk_prom.PromEntryNameDefault]
    entry.Bootstrap(fac.CreateEvent())
    entry.WaitForShutdownSig(1 * time.Second)
}
```

## Example
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
