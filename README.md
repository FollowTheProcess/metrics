# Metrics

[![License](https://img.shields.io/github/license/FollowTheProcess/metrics)](https://github.com/FollowTheProcess/metrics)
[![Go Reference](https://pkg.go.dev/badge/github.com/FollowTheProcess/metrics.svg)](https://pkg.go.dev/github.com/FollowTheProcess/metrics)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/metrics)](https://goreportcard.com/report/github.com/FollowTheProcess/metrics)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/metrics?logo=github&sort=semver)](https://github.com/FollowTheProcess/metrics)
[![CI](https://github.com/FollowTheProcess/metrics/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/metrics/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/metrics/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/metrics)

Go library for publishing AWS CloudWatch Embedded Metrics Format (EMF) logs 📈

## Project Description

AWS CloudWatch allows publishing of metrics using a special format known as [Embedded Metrics Format] (EMF), which involves printing specially formatted JSON messages to stdout
from inside a Lambda function. CloudWatch then scans for these messages in the logs, parses them and publishes the metrics behind the scenes.

The primary advantage being this can be done without having to make a HTTP call to the CloudWatch API through AWS client libraries, which means it's much faster and your metrics
can be gathered up asynchronously by CloudWatch in the background.

### EMF JSON

An EMF JSON blob looks like this:

```json
{
  "_aws": {
    "Timestamp": 1574109732004,
    "CloudWatchMetrics": [
      {
        "Namespace": "lambda-function-metrics",
        "Dimensions": [["functionVersion"]],
        "Metrics": [
          {
            "Name": "time",
            "Unit": "Milliseconds",
            "StorageResolution": 60
          }
        ]
      }
    ]
  },
  "functionVersion": "$LATEST",
  "time": 100,
  "requestId": "989ffbf8-9ace-4817-a57c-e4dd734019ee"
}
```

## Installation

```shell
go get go.followtheprocess.codes/metrics@latest
```

## Quickstart

The library exposes a `Logger` type that accumulates metrics during your application and can be flushed to stdout on command, although typical usage will `defer` the call
to `Flush`:

```go
import (
    "go.followtheprocess.codes/metrics"
    "go.followtheprocess.codes/metrics/unit"
)
// Assuming you're inside a lambda handler

// Get a new Logger
m := metrics.New()

// Something happened 10 times!
m.Count("something", 10, metrics.StandardResolution)

// Add new dimensions
m.Dimension("MyDimension", "value")

// Generic metrics, and configurable resolution
m.Add("Foo", 27, unit.Percent, metrics.HighResolution)

// Set namespaces
m.Namespace("MyNameSpace").Add("Foo", 27, unit.Percent, metrics.StandardResolution)

// Just before you return, flush the Logger so CloudWatch can scan your metrics
defer m.Flush()

// Beware, `Flush()` returns an error which you should handle...
defer func() {
    err = m.Flush()
}()
```

### Credits

This package was created with [copier] and the [FollowTheProcess/go_copier] project template.

[copier]: https://copier.readthedocs.io/en/stable/
[FollowTheProcess/go_copier]: https://github.com/FollowTheProcess/go_copier
[Embedded Metrics Format]: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html
