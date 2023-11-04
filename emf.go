// Package emf is a placeholder for something cool.
package emf

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/FollowTheProcess/emf/unit"
)

// StorageResolution represents the metrics resolution.
type StorageResolution int

const (
	HighResolution     StorageResolution = 1  // 1 second resolution, for high precision metrics
	StandardResolution StorageResolution = 60 // Minute resolution, suitable for most metrics
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
}

// Option is a functional option to configure a Logger.
type Option func(logger *Logger)

// Stdout sets the output for a Logger.
func Stdout(stdout io.Writer) Option {
	return func(logger *Logger) {
		logger.stdout = stdout
		logger.encoder = json.NewEncoder(stdout)
	}
}

// LogGroupName sets the CloudWatch log group name for a Logger.
func LogGroupName(name string) Option {
	return func(logger *Logger) {
		logger.logGroupName = name
	}
}

// New returns an EMF Metrics logger suitable for use in Lambda functions.
func New(opts ...Option) Logger {
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

	return logger
}

// Count records a count metric.
func (l *Logger) Count(name string, count int) *Logger {
	l.put(name, count, unit.Count, StandardResolution)
	return l
}

// Log outputs the current state of the metrics Logger in EMF JSON.
func (l *Logger) Log() error {
	if len(l.metrics.Metrics) == 0 {
		// Bail early if we have nothing to do
		return nil
	}

	l.values["_aws"] = Metadata{
		Timestamp:    time.Now().UnixMilli(),
		Metrics:      []MetricDirective{l.metrics},
		LogGroupName: l.logGroupName,
	}

	if err := l.encoder.Encode(l.values); err != nil {
		return fmt.Errorf("Could not encode metrics to JSON: %w", err)
	}
	return nil
}

// put inserts a metric into the Logger, to be flushed later.
func (l *Logger) put(name string, value any, unit unit.Unit, resolution StorageResolution) {
	// Store the metric metadata
	metric := MetricDefinition{
		Name:       name,
		Unit:       unit,
		Resolution: resolution,
	}
	l.metrics.Metrics = append(l.metrics.Metrics, metric)

	// Add the metric values to the root node
	l.values[name] = value
}
