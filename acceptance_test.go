package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func TestGroupCapsAndInclusiveGate(t *testing.T) {
	c := qt.New(t)
	policy := `version: 1
rules:
  filler.announced-importance:
    score: {weight: 80, cap: 80}
  filler.wordy-phrase:
    score: {weight: 80, cap: 80}
gate:
  sentence_score: {fail_at: 40, min_words: 1}
  paragraph_score: {fail_at: 100, min_words: 1}
`
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "draft.txt", Format: document.Plain,
		Bytes: []byte("It is important to note that the client opens a connection in order to send a request.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Findings, qt.HasLen, 3)
	for _, assessment := range result.Assessments {
		c.Assert(assessment.SlopScore, qt.Equals, float64(40))
		c.Assert(assessment.Contributions, qt.HasLen, 2)
	}
	c.Assert(result.Gate.Reasons, qt.HasLen, 1)
	c.Assert(result.Gate.Reasons[0].Code, qt.Equals, "gate.sentence-score")
}

func TestTechnicalBoundaries(t *testing.T) {
	base := "The client opens a connection to the server and sends the request with its credentials."
	contrast := "The client not only reads the file but also checks its size."
	cases := []struct {
		name, id, text string
		format         document.Format
		matches        bool
	}{
		{"negation", "repetition.near-sentence", base + " " + strings.Replace(base, "opens", "never opens", 1), document.Plain, false},
		{
			"numbers",
			"repetition.near-sentence",
			base + " Wait 10 seconds. " + strings.Replace(base, "a connection", "2 connections", 1),
			document.Plain,
			false,
		},
		{
			"code boundary",
			"filler.announced-importance",
			"It is`x`important to note that the client retries.",
			document.Markdown,
			false,
		},
		{
			"URL boundary",
			"filler.announced-importance",
			"It is https://example.com important to note that the client retries.",
			document.Plain,
			false,
		},
		{"one contrast", "syntax.not-only-density", contrast, document.Plain, false},
		{
			"window boundary",
			"syntax.not-only-density",
			strings.Repeat("The client reads a file. ", 7) + contrast + " " + contrast,
			document.Plain,
			true,
		},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			var selected rule.Rule
			for _, candidate := range builtin.Rules() {
				if candidate.Descriptor().ID == row.id {
					selected = candidate
				}
			}
			c.Assert(selected, qt.IsNotNil)
			engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{selected}})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: "fixture", Format: row.format, Bytes: []byte(row.text)})
			c.Assert(err, qt.IsNil)
			c.Assert(len(result.Findings) > 0, qt.Equals, row.matches)
		})
	}
}
