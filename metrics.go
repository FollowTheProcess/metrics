// Package metrics provides a simple, idiomatic API for Go Lambda Functions to record custom metrics in the CloudWatch
// Embedded Metrics Format (EMF).
//
// It is fully compliant with the [EMF Specification]
//
// [EMF Specification]: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Embedded_Metric_Format_Specification.html
package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/FollowTheProcess/metrics/unit"
)

// StorageResolution represents the metrics resolution.
type StorageResolution int

const (
	HighResolution     StorageResolution = 1  // 1 second resolution, for high precision metrics
	StandardResolution StorageResolution = 60 // 1 minute resolution, suitable for most metrics
)

// Metadata encodes the EMF Metadata object.
type Metadata struct {
	// Optional name of the CloudWatch log group.
	LogGroupName string `json:"LogGroupName,omitempty"`

	// List of metric directives.
	Metrics []MetricDirective `json:"CloudWatchMetrics"`

	// UNIX (milliseconds) timestamp for the metric.
	Timestamp int64 `json:"Timestamp"`
}

// MetricDirective encodes the EMF MetricDirective object.
type MetricDirective struct {
	// The CloudWatch namespace for the metric.
	Namespace string `json:"Namespace"`

	// List of EMF dimension keys.
	Dimensions []Dimension `json:"Dimensions"`

	// The actual metric definitions.
	Metrics []MetricDefinition `json:"Metrics"`
}

// Dimension encodes a single EMF metric dimension.
type Dimension []string

// MetricDefinition encodes a single EMF metric definition.
type MetricDefinition struct {
	// The name of the metric
	Name string `json:"Name"`

	// The unit of measurement, optional. If omitted, None is assumed
	Unit unit.Unit `json:"Unit,omitempty"`

	// Resolution for the metric, optional. If omitted, standard resolution is assumed.
	Resolution StorageResolution `json:"StorageResolution,omitempty"`
}

// Logger is the mechanism to write EMF metrics.
//
// A Logger is safe to use concurrently across goroutines.
type Logger struct {
	// Where to write EMF metrics to.
	stdout io.Writer

	// The actual metric values to encode into the root node of the EMF JSON, they must
	// be an unstructured map as we don't know ahead of time what metrics the user will add.
	values map[string]any

	// The JSON encoder.
	encoder *json.Encoder

	// The name of the CloudWatch Log Group.
	logGroupName string

	// The actual metrics.
	metrics MetricDirective

	// Synchronisation
	mu sync.Mutex
}

// Option is a functional option to configure a Logger.
type Option func(logger *Logger)

// Stdout sets the output for a Logger.
func Stdout(stdout io.Writer) Option {
	return func(logger *Logger) {
		logger.mu.Lock()
		defer logger.mu.Unlock()
		logger.stdout = stdout
		logger.encoder = json.NewEncoder(stdout)
	}
}

// Indent sets JSON indenting for the output, primarily
// used for testing.
func Indent(indent bool) Option {
	return func(logger *Logger) {
		logger.mu.Lock()
		defer logger.mu.Unlock()
		logger.encoder.SetIndent("", "  ")
	}
}

// LogGroupName sets the CloudWatch log group name for a Logger.
func LogGroupName(name string) Option {
	return func(logger *Logger) {
		logger.mu.Lock()
		defer logger.mu.Unlock()
		logger.logGroupName = name
	}
}

// New returns an EMF Metrics logger suitable for use in Lambda functions.
func New(opts ...Option) *Logger {
	logger := Logger{
		// Default to os.Stdout
		stdout:  os.Stdout,
		encoder: json.NewEncoder(os.Stdout),
	}

	for _, opt := range opts {
		opt(&logger)
	}

	values := make(map[string]any)

	// Default env vars/config exposed in all lambda functions
	// https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars.html#configuration-envvars-runtime
	name := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	values["executionEnvironment"] = os.Getenv("AWS_EXECUTION_ENV")
	values["memorySize"] = os.Getenv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE")
	values["functionVersion"] = os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")
	values["logStreamId"] = os.Getenv("AWS_LAMBDA_LOG_STREAM_NAME")
	values["functionName"] = name

	traceID := os.Getenv("_X_AMZN_TRACE_ID")
	if strings.Contains(traceID, "Sampled=1") {
		values["traceId"] = traceID
	}

	dimensions := []Dimension{{"ServiceName", "ServiceType"}}
	values["ServiceType"] = "AWS::Lambda::Function"
	values["ServiceName"] = name

	logger.values = values
	logger.metrics.Dimensions = dimensions
	logger.metrics.Namespace = "aws-embedded-metrics"

	return &logger
}

// Count records a simple count metric.
//
//	m := metrics.New()
//	m.Count("UserLogIn", 1, metrics.StandardResolution)
func (l *Logger) Count(name string, count int, res StorageResolution) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store(name, count, unit.Count, res)
	return l
}

// Add records a generic user defined metric.
//
//	m := metrics.New()
//	m.Add("FileSize", 256, unit.Megabytes, metrics.StandardResolution)
func (l *Logger) Add(name string, value any, unit unit.Unit, res StorageResolution) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.store(name, value, unit, res)
	return l
}

// Dimension adds a metrics dimension.
//
//	m := metrics.New()
//	m.Dimension("MyDimension", "Yes")
func (l *Logger) Dimension(key, value string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.metrics.Dimensions = append(l.metrics.Dimensions, Dimension{key})
	l.values[key] = value
	return l
}

// Namespace sets the namespace for the Logger.
//
//	m := metrics.New()
//	m.Namespace("MyCustomNamespace")
func (l *Logger) Namespace(namespace string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.metrics.Namespace = namespace
	return l
}

// Flush outputs the collected metrics to stdout so they can be discovered by CloudWatch.
//
// Typical usage in a lambda handler would be to populate metrics throughout and then
// defer a call to Flush before returning to the main entry point.
//
//	m := metrics.New()
//	m.Count("something", 5) // Something happened 5 times, very important business metric!
//	... // More logic
//	defer m.Flush() // Although you should handle the error Flush returns
func (l *Logger) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.metrics.Metrics) == 0 {
		// Bail early if we have nothing to do
		return nil
	}

	if os.Getenv("METRICS_OMIT_TIMESTAMP") != "" {
		l.values["_aws"] = Metadata{
			Metrics:      []MetricDirective{l.metrics},
			LogGroupName: l.logGroupName,
		}
	} else {
		l.values["_aws"] = Metadata{
			Timestamp:    time.Now().UTC().UnixMilli(),
			Metrics:      []MetricDirective{l.metrics},
			LogGroupName: l.logGroupName,
		}
	}

	if err := l.encoder.Encode(l.values); err != nil {
		return fmt.Errorf("could not encode metrics to JSON: %w", err)
	}
	return nil
}

// store inserts a metric into the Logger, to be flushed later.
func (l *Logger) store(name string, value any, unit unit.Unit, res StorageResolution) {
	// Store the metric metadata
	metric := MetricDefinition{
		Name:       name,
		Unit:       unit,
		Resolution: res,
	}
	l.metrics.Metrics = append(l.metrics.Metrics, metric)

	// Add the metric values to the root
	l.values[name] = value
}
