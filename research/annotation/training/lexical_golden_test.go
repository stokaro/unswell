package training_test

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

func TestLexicalUnigramVocabularyGoldenAndUnknownWords(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	input.ReplaceSource(5, "Quasar xylophone.")
	candidates, round := input.Compile(t)
	options := lexicalFitting()
	options.Calibration = "none"
	vocabulary := training.LexicalOptions{Counts: feature.LexicalOptions{WordMin: 1, WordMax: 1}, MaxFeatures: 128, MinTargets: 1}
	fitted, err := training.RunLexical(t.Context(), candidates, round, input.Files, options, vocabulary)
	c.Assert(err, qt.IsNil)
	counts := make(map[string]int)
	for _, term := range fitted.Lexical.Terms {
		counts[term.Key] = term.Targets
	}
	c.Assert(counts, qt.DeepEquals, map[string]int{
		`w:["retry"]`: 2, `w:["failed"]`: 1, `w:["requests"]`: 1, `w:["it"]`: 1,
		`w:["is"]`: 1, `w:["important"]`: 1, `w:["to"]`: 1, `w:["note"]`: 1,
		`w:["that"]`: 1, `w:["clients"]`: 1, `w:["may"]`: 1,
	})
	weight := 0.0
	for _, value := range fitted.Logistic.Weights {
		weight += math.Abs(value)
	}
	c.Assert(weight > 0, qt.IsTrue)
	p := fitted.Logistic
	classifier, err := model.NewLogistic(model.Parameters{Means: p.Means, Scales: p.Scales, Weights: p.Weights, Intercept: p.Intercept})
	c.Assert(err, qt.IsNil)
	// The held-out target has only unknown words: its known-term counts are zero.
	want, err := classifier.Evaluate(t.Context(), make([]float64, len(fitted.Lexical.Terms)))
	c.Assert(err, qt.IsNil)
	got, err := training.Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(got.Rows, qt.HasLen, 1)
	c.Assert(got.Rows[0].Status, qt.Equals, "available")
	c.Assert(*got.Rows[0].LinearScore, qt.Equals, want.LinearScore)
	c.Assert(*got.Rows[0].Response, qt.Equals, want.Response)
}
