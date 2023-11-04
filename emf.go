// Package emf is a placeholder for something cool.
package emf

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
	Metrics []Metric `json:"CloudWatchMetrics"`

	// UNIX (milliseconds) timestamp for the metric.
	Timestamp int64 `json:"Timestamp"`
}

// Metric encodes the EMF MetricDirective object.
type Metric struct {
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
	Unit Unit `json:"Unit,omitempty"`

	// Resolution for the metric, optional. If omitted, standard resolution is assumed.
	Resolution StorageResolution `json:"StorageResolution,omitempty"`
}
