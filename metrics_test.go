package metrics_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/metrics"
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
			name: "count",
			logFn: func(logger *metrics.Logger) {
				logger.Count("something", 5, metrics.StandardResolution)
			},
			want: "count.json",
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
