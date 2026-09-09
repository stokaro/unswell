package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestRepetitionGroupActivationsPreserveClustersAndIndependentProse(t *testing.T) {
	const sentence = "The service accepts a request from the client and sends the response with its credentials."
	for _, test := range []struct {
		name, id, text string
		values         []float64
		findings       int
	}{
		{"exact singleton", "repetition.exact-sentence", sentence, []float64{0}, 0},
		{"exact repeated across blocks", "repetition.exact-sentence", sentence + "\n\n" + sentence, []float64{1, 1}, 1},
		{"exact with independent prose", "repetition.exact-sentence", sentence + "\n\n" + sentence +
			"\n\nEach connection has a timeout that the client checks before sending another request to the server.", []float64{1, 1, 0}, 1},
		{"opener cluster", "repetition.paragraph-openers", "The client opens a connection to the database.\n\n" +
			"The client opens a file from the disk.\n\nThe client opens a channel to the service.", []float64{0.333, 0.333, 0.333}, 1},
		{"opener below threshold", "repetition.paragraph-openers", "The client opens a connection to the database.\n\n" +
			"The client opens a file from the disk.", []float64{0, 0}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			options := unswell.Options{Features: []string{"activation/" + test.id}, Config: []byte(
				"version: 1\nextends: [builtin:custom]\nrules:\n  " + test.id + ": {enabled: true}\n")}
			engine, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(test.text)}
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, test.findings)
			assertPhraseMeasurements(t, result, make([]string, len(test.values)), test.values)
			options.Features = nil
			ordinary, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			want, err := ordinary.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			result.Features = nil
			c.Assert(result, qt.DeepEquals, want)
		})
	}
}

func TestRepetitionGroupActivationsRetainSuppressedEvidence(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{"activation/repetition.exact-sentence"}, Config: []byte(
		"version: 1\nextends: [builtin:custom]\nrules:\n  repetition.exact-sentence: {enabled: true, parameters: {min_words: 1}}\n")})
	c.Assert(err, qt.IsNil)
	block := "<!-- unswell-disable-next-block repetition.exact-sentence -- Required repeated definition. -->\n\nThe client starts.\n\n"
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(block + block)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	assertPhraseMeasurements(t, result, []string{"", ""}, []float64{1, 1})
}

func TestRepetitionGroupActivationsRetainCancellationAndConcurrentOwnership(t *testing.T) {
	checkConcurrentActivations(t, repetitionGroupActivationIDs(),
		"The client opens a connection to the database. The client opens a file from the disk. The client opens a channel to the service.", 1)
}
