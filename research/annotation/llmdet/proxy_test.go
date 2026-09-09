package llmdet_test

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/llmdet"
)

func TestProxyMatchesPinnedFunctionControls(t *testing.T) {
	c := qt.New(t)
	data, err := os.ReadFile("testdata/proxy-controls.json")
	c.Assert(err, qt.IsNil)
	var cases []struct {
		ID             string           `json:"id"`
		Tokens         []int            `json:"tokens"`
		Spec           llmdet.ProxySpec `json:"spec"`
		ReferenceScore float64          `json:"reference_score"`
	}
	c.Assert(json.Unmarshal(data, &cases), qt.IsNil)
	c.Assert(cases, qt.HasLen, 15)
	for _, test := range cases {
		t.Run(test.ID, func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			data, err := json.Marshal(test.Spec)
			c.Assert(err, qt.IsNil)
			proxy, err := llmdet.LoadProxy(t.Context(), data)
			c.Assert(err, qt.IsNil)
			result, err := proxy.Measure(t.Context(), test.Tokens)
			c.Assert(err, qt.IsNil)
			c.Assert(math.Abs(result.ReferenceScore-test.ReferenceScore) <= 1e-12, qt.IsTrue)
		})
	}
}

func TestProxyReportsCoverageAndSkippedLikelihoods(t *testing.T) {
	c := qt.New(t)
	spec := llmdet.ProxySpec{Version: llmdet.ProxyVersion, VocabSize: 16, Rows: []llmdet.ProbabilityRow{
		{Context: []int{1, 2, 3}, Continuations: []int{4}, Probabilities: []float64{0}},
		{Context: []int{4}, Continuations: []int{5}, Probabilities: []float64{0.5}},
	}}
	proxy, err := llmdet.NewProxy(t.Context(), spec)
	c.Assert(err, qt.IsNil)
	result, err := proxy.Measure(t.Context(), []int{1, 2, 3, 4, 5, 6})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Possible, qt.Equals, 3)
	c.Assert(result.Matched, qt.Equals, 2)
	c.Assert(result.Evaluated, qt.Equals, 1)
	c.Assert(result.Skipped, qt.Equals, 1)
	c.Assert(result.Orders, qt.Equals, [3]int{1, 0, 1})
	c.Assert(result.Coverage, qt.Equals, 2.0/3)
	c.Assert(result.ValueCoverage, qt.Equals, 1.0/3)
	c.Assert(result.ReferenceScore, qt.Equals, 1.0/3)
	c.Assert(result.Value, qt.IsNotNil)
	c.Assert(result.Reason, qt.Equals, "")
	*result.Value = 100
	spec.Rows[0].Context[0] = 15
	spec.Rows[0].Probabilities[0] = 1
	spec.Rows[1].Continuations[0] = 15
	result, err = proxy.Measure(t.Context(), []int{1, 2, 3, 4})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Value, qt.IsNil)
	c.Assert(result.Reason, qt.Equals, "no_evaluated_probability")
	c.Assert(result.Matched, qt.Equals, 1)
}

func TestProxyNeverTreatsMissingEvidenceAsAZeroFeature(t *testing.T) {
	c := qt.New(t)
	proxy, err := llmdet.NewProxy(t.Context(), llmdet.ProxySpec{Version: llmdet.ProxyVersion, VocabSize: 16})
	c.Assert(err, qt.IsNil)
	for _, test := range []struct {
		tokens []int
		reason string
	}{
		{nil, "insufficient_tokens"}, {[]int{1, 2, 3}, "insufficient_tokens"},
		{[]int{1, 2, 3, 4}, "no_matching_context"},
	} {
		result, err := proxy.Measure(t.Context(), test.tokens)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Value, qt.IsNil)
		c.Assert(result.ReferenceScore, qt.Equals, 0.0)
		c.Assert(result.Reason, qt.Equals, test.reason)
	}
	_, err = proxy.Measure(t.Context(), []int{16})
	c.Assert(err, qt.ErrorMatches, "token ID .* outside vocabulary")
	_, err = proxy.Measure(t.Context(), make([]int, llmdet.MaxTokens+1))
	c.Assert(err, qt.ErrorMatches, ".*excessive token count")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = proxy.Measure(ctx, nil)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestProxyRejectsInvalidRows(t *testing.T) {
	for _, test := range []struct {
		name string
		rows []llmdet.ProbabilityRow
	}{
		{"empty context", []llmdet.ProbabilityRow{{}}},
		{"long context", []llmdet.ProbabilityRow{{Context: []int{1, 2, 3, 4}}}},
		{"unknown ID", []llmdet.ProbabilityRow{{Context: []int{-1}}}},
		{"length mismatch", []llmdet.ProbabilityRow{{Context: []int{1}, Continuations: []int{2}}}},
		{"duplicate token", []llmdet.ProbabilityRow{{Context: []int{1}, Continuations: []int{2, 2}, Probabilities: []float64{0.1, 0.1}}}},
		{"negative", []llmdet.ProbabilityRow{{Context: []int{1}, Continuations: []int{2}, Probabilities: []float64{-0.1}}}},
		{"NaN", []llmdet.ProbabilityRow{{Context: []int{1}, Continuations: []int{2}, Probabilities: []float64{math.NaN()}}}},
		{"Inf", []llmdet.ProbabilityRow{{Context: []int{1}, Continuations: []int{2}, Probabilities: []float64{math.Inf(1)}}}},
		{"mass", []llmdet.ProbabilityRow{{Context: []int{1}, Continuations: []int{2, 3}, Probabilities: []float64{0.6, 0.6}}}},
		{"duplicate context", []llmdet.ProbabilityRow{{Context: []int{1}}, {Context: []int{1}}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := llmdet.NewProxy(t.Context(), llmdet.ProxySpec{Version: llmdet.ProxyVersion, VocabSize: 16, Rows: test.rows})
			c.Assert(err, qt.IsNotNil)
		})
	}
}
