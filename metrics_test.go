package metrics_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/metrics"
	"github.com/FollowTheProcess/metrics/unit"
	"github.com/FollowTheProcess/test"
	"github.com/kinbiko/jsonassert"
)

func TestMetricsLog(t *testing.T) {
	testdata := test.Data(t)

	tests := []struct {
		name  string
		env   map[string]string     // Env vars to set before each test
		logFn func(*metrics.Logger) // Function to apply to the logger for the test
		want  string                // Name of file containing expected JSON
	}{
		{
			name:  "no metrics means empty json",
			logFn: func(logger *metrics.Logger) {},
			want:  "empty.json",
		},
		{
			name: "count",
			logFn: func(logger *metrics.Logger) {
				logger.Count("something", 5, metrics.StandardResolution)
			},
			want: "count.json",
		},
		{
			name: "count with trace id",
			env:  map[string]string{"_X_AMZN_TRACE_ID": "something_looks_like_id_with_Sampled=1"},
			logFn: func(logger *metrics.Logger) {
				logger.Count("something", 5, metrics.StandardResolution)
			},
			want: "traceid.json",
		},
		{
			name: "count high resolution",
			logFn: func(logger *metrics.Logger) {
				logger.Count("something", 27, metrics.HighResolution)
			},
			want: "count-high-res.json",
		},
		{
			name: "dimension and a count",
			logFn: func(logger *metrics.Logger) {
				logger.Dimension("TestDimension", "value").Count("something", 7, metrics.StandardResolution)
			},
			want: "dimensions.json",
		},
		{
			name: "generic metric",
			logFn: func(logger *metrics.Logger) {
				logger.Add("Foo", 27, unit.Percent, metrics.StandardResolution)
			},
			want: "generic.json",
		},
		{
			name: "all the lambda env vars",
			env: map[string]string{
				"AWS_LAMBDA_FUNCTION_NAME":        "FuncyFunc",
				"AWS_EXECUTION_ENV":               "Go1.21",
				"AWS_LAMBDA_FUNCTION_MEMORY_SIZE": "256",
				"AWS_LAMBDA_FUNCTION_VERSION":     "12",
				"AWS_LAMBDA_LOG_STREAM_NAME":      "LoggyLog",
			},
			logFn: func(logger *metrics.Logger) {
				logger.Add("Bar", 267, unit.BytesPerSecond, metrics.HighResolution)
			},
			want: "env.json",
		},
		{
			name: "namespace",
			logFn: func(logger *metrics.Logger) {
				logger.Namespace("MyNameSpace").Add("Foo", 27, unit.Percent, metrics.StandardResolution)
			},
			want: "namespace.json",
		},
		{
			name: "log group name",
			logFn: func(logger *metrics.Logger) {
				logger.Namespace("MyNameSpace").Add("Foo", 27, unit.Percent, metrics.StandardResolution)
				fn := metrics.LogGroupName("MyLogGroup")
				fn(logger)
			},
			want: "log-group-name.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars if specified
			if len(tt.env) != 0 {
				for key, value := range tt.env {
					t.Setenv(key, value)
				}
			}

			buf := &bytes.Buffer{}
			logger := metrics.New(metrics.Stdout(buf))
			tt.logFn(logger)

			test.Ok(t, logger.Flush(), "logger.Flush() returned an error")

			want, err := os.ReadFile(filepath.Join(testdata, tt.want))
			test.Ok(t, err, "read tt.want")

			ja := jsonassert.New(t)
			ja.Assertf(buf.String(), string(want))
		})
	}
}
