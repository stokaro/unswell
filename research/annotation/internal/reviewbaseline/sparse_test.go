package reviewbaseline_test

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/reviewbaseline"
)

func executeSparse(t *testing.T, input map[string]any) map[string]any {
	t.Helper()
	c := qt.New(t)
	data, err := json.Marshal(input)
	c.Assert(err, qt.IsNil)
	var output bytes.Buffer
	c.Assert(reviewbaseline.RunSparse(t.Context(), bytes.NewReader(data), &output), qt.IsNil)
	var result map[string]any
	c.Assert(json.Unmarshal(output.Bytes(), &result), qt.IsNil)
	return result
}

func TestSparseComparisonPreservesDenseBaselineAndLabelIsolation(t *testing.T) {
	c := qt.New(t)
	input := fixture(t, false)
	dense := execute(t, input)["models"].([]any)[3].(map[string]any)
	first := executeSparse(t, input)["models"].([]any)[0].(map[string]any)
	c.Assert(first["kind"], qt.Equals, "SW128")
	c.Assert(dense["features"], qt.DeepEquals, first["features"])
	for i, value := range first["predictions"].([]any) {
		prediction := value.(map[string]any)
		reference := dense["predictions"].([]any)[i].(map[string]any)
		// The earlier command stops at gradient 1e-6. Its many correlated
		// columns amplify the remaining score error; numerical objective parity
		// uses the tighter dense reference in sparsemodel's separate tests.
		c.Assert(math.Abs(prediction["raw_score"].(float64)-reference["raw_score"].(float64)) < 1e-3,
			qt.IsTrue, qt.Commentf("sparse=%g dense=%g", prediction["raw_score"], reference["raw_score"]))
		c.Assert(prediction["selected"], qt.Equals, reference["selected"])
	}
	for _, page := range input["pages"].([]map[string]any) {
		if page["fold"] == 0 {
			for _, unit := range page["units"].([]map[string]any) {
				unit["label"] = 1 - unit["label"].(int)
			}
		}
	}
	second := executeSparse(t, input)["models"].([]any)[0].(map[string]any)
	for _, key := range []string{"features", "parameters", "training_sha256", "operating_point"} {
		c.Assert(second[key], qt.DeepEquals, first[key], qt.Commentf("evaluation labels affected %s", key))
	}
}
