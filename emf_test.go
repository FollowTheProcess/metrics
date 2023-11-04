package emf_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/FollowTheProcess/emf"
	"github.com/FollowTheProcess/emf/unit"
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
				Metrics: []emf.MetricDirective{
					{
						Namespace: "lambda-function-metrics",
						Dimensions: []emf.Dimension{
							{"functionVersion"},
						},
						Metrics: []emf.MetricDefinition{
							{
								Name:       "time",
								Unit:       unit.Milliseconds,
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

func TestCount(t *testing.T) {
	buf := &bytes.Buffer{}
	metrics := emf.New(emf.Stdout(buf), emf.LogGroupName("test"))

	metrics.Count("something", 5)
	test.Ok(t, metrics.Log(), "metrics.Log() returned an error")

	count := filepath.Join(test.Data(t), "count.json")
	want, err := os.ReadFile(count)
	test.Ok(t, err, "read count.json")

	ja := jsonassert.New(t)
	ja.Assertf(buf.String(), string(want))
}
