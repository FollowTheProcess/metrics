package emf_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/emf"
	"github.com/FollowTheProcess/test"
	"github.com/kinbiko/jsonassert"
)

func TestMetricsJSON(t *testing.T) {
	testdata := test.Data(t)

	tests := []struct {
		name    string
		want    string // Name of the file in testdata containing the expected JSON
		metrics emf.Metadata
	}{
		{
			name: "valid spec test case",
			metrics: emf.Metadata{
				Metrics: []emf.Metric{
					{
						Namespace: "lambda-function-metrics",
						Dimensions: []emf.Dimension{
							{"functionVersion"},
						},
						Metrics: []emf.MetricDefinition{
							{
								Name:       "time",
								Unit:       emf.Milliseconds,
								Resolution: emf.StandardResolution,
							},
						},
					},
				},
				Timestamp: 1574109732004,
			},
			want: "metadata-only.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.metrics)
			test.Ok(t, err, "json.Marshal(tt.metrics)")

			file := filepath.Join(testdata, tt.want)
			want, err := os.ReadFile(file)
			test.Ok(t, err, "read tt.want")

			ja := jsonassert.New(t)
			ja.Assertf(string(got), string(want))
		})
	}
}
