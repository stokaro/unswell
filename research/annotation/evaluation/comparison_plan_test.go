package evaluation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation/training"
)

func comparisonInputs() (ComparisonPlan, training.Predictions, training.Predictions) {
	plan := ComparisonPlan{Version: ComparisonVersion, ID: "frozen", ProtocolSHA256: strings.Repeat("a", 64),
		CandidateSHA256: strings.Repeat("b", 64), ComparatorSHA256: strings.Repeat("c", 64)}
	a := training.Predictions{SHA256: plan.CandidateSHA256, Plan: training.PredictionPlan{
		ProtocolSHA256: plan.ProtocolSHA256, CorpusSHA256: strings.Repeat("d", 64), Partition: "final_test", Context: "prepared_piece"}}
	b := a
	b.SHA256 = plan.ComparatorSHA256
	return plan, a, b
}

func TestComparisonRejectsUnmatchedScopes(t *testing.T) {
	for _, row := range []struct {
		name   string
		change func(*training.Predictions)
	}{
		{"selection", func(p *training.Predictions) { p.SHA256 = strings.Repeat("0", 64) }},
		{"protocol", func(p *training.Predictions) { p.Plan.ProtocolSHA256 = strings.Repeat("0", 64) }},
		{"corpus", func(p *training.Predictions) { p.Plan.CorpusSHA256 = strings.Repeat("0", 64) }},
		{"partition", func(p *training.Predictions) { p.Plan.Partition = "development" }},
		{"context", func(p *training.Predictions) { p.Plan.Context = "source_document" }},
		{"kind", func(p *training.Predictions) { p.Model.Identity.Kind = "sentence" }},
		{"profile", func(p *training.Predictions) { p.Model.Identity.ProfileSHA256 = strings.Repeat("0", 64) }},
		{"NLP", func(p *training.Predictions) { p.Model.Identity.NLP.Version = "different" }},
		{"policy", func(p *training.Predictions) { p.Model.Identity.ExtractionPolicyHash = "different" }},
		{"preparation", func(p *training.Predictions) { p.Model.Identity.PreparationHash = "different" }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			plan, a, b := comparisonInputs()
			_, err := comparisonPlanIdentity(t.Context(), plan, a, b)
			c.Assert(err, qt.IsNil)
			row.change(&b)
			hash, err := comparisonPlanIdentity(t.Context(), plan, a, b)
			c.Assert(err, qt.IsNotNil)
			c.Assert(hash, qt.Equals, "")
		})
	}
}

func TestComparisonPlanStrictInput(t *testing.T) {
	c := qt.New(t)
	plan, a, b := comparisonInputs()
	data, err := json.Marshal(plan)
	c.Assert(err, qt.IsNil)
	loaded, err := LoadComparisonPlan(t.Context(), data)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, plan)
	b.Model.Identity.ColumnsSHA256 = strings.Repeat("f", 64)
	b.Model.Identity.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS}
	b.Plan.Response = "isotonic"
	_, err = comparisonPlanIdentity(t.Context(), plan, a, b)
	c.Assert(err, qt.IsNil)
	for _, bad := range []string{`{}`, `null`, string(data) + `{}`, strings.Replace(string(data), `"id":`, `"id":"x","id":`, 1),
		strings.Replace(string(data), `"id":`, `"unknown":`, 1), strings.ReplaceAll(string(data), strings.Repeat("a", 64), "ABC"),
		strings.ReplaceAll(string(data), "frozen", strings.Repeat("x", 129))} {
		_, err := LoadComparisonPlan(t.Context(), []byte(bad))
		c.Assert(err, qt.IsNotNil)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = LoadComparisonPlan(ctx, data)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func FuzzComparisonPlan(f *testing.F) {
	plan, _, _ := comparisonInputs()
	data, err := json.Marshal(plan)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(data)
	f.Add([]byte(`{"version":null}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = LoadComparisonPlan(t.Context(), data)
	})
}
