package evaluation_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

func frozenComparisonPlan() evaluation.ComparisonPlan {
	return evaluation.ComparisonPlan{Version: evaluation.ComparisonVersion, ID: "frozen", ProtocolSHA256: strings.Repeat("a", 64),
		CandidateSHA256: strings.Repeat("b", 64), ComparatorSHA256: strings.Repeat("c", 64)}
}

func TestComparisonPlanStrictInput(t *testing.T) {
	c := qt.New(t)
	plan := frozenComparisonPlan()
	data, err := json.Marshal(plan)
	c.Assert(err, qt.IsNil)
	loaded, err := evaluation.LoadComparisonPlan(t.Context(), data)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, plan)
	for _, bad := range []string{`{}`, `null`, string(data) + `{}`, strings.Replace(string(data), `"id":`, `"id":"x","id":`, 1),
		strings.Replace(string(data), `"id":`, `"unknown":`, 1), strings.ReplaceAll(string(data), strings.Repeat("a", 64), "ABC"),
		strings.ReplaceAll(string(data), "frozen", strings.Repeat("x", 129))} {
		_, err := evaluation.LoadComparisonPlan(t.Context(), []byte(bad))
		c.Assert(err, qt.IsNotNil)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = evaluation.LoadComparisonPlan(ctx, data)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func FuzzComparisonPlan(f *testing.F) {
	plan := frozenComparisonPlan()
	data, err := json.Marshal(plan)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(data)
	f.Add([]byte(`{"version":null}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = evaluation.LoadComparisonPlan(t.Context(), data)
	})
}
