package unswell_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func TestExecutableCatalog(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		descriptor := implementation.Descriptor()
		t.Run(descriptor.ID, func(t *testing.T) {
			c := qt.New(t)
			for _, example := range descriptor.Examples {
				engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{implementation}, Config: []byte(example.Config)})
				c.Assert(err, qt.IsNil)
				result, err := engine.Analyze(
					t.Context(),
					document.Source{Name: "fixture.txt", Format: document.Plain, Bytes: []byte(example.Text)},
				)
				c.Assert(err, qt.IsNil)
				c.Check(len(result.Findings) > 0, qt.Equals, example.Match, qt.Commentf("%q", example.Text))
				for _, finding := range result.Findings {
					c.Check(finding.RuleID, qt.Equals, descriptor.ID)
					c.Check(finding.Primary.Span.Valid(len(example.Text)), qt.IsTrue)
				}
			}
		})
	}
}

func TestGateAndProbability(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(
		t.Context(),
		document.Source{Name: "draft.md", Format: document.Markdown, Bytes: []byte("Certainly! The client can retry the request.")},
	)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Status, qt.Equals, "complete")
	c.Assert(result.Findings[0].RuleID, qt.Equals, "scaffold.chat-preamble")
	for _, assessment := range result.Assessments {
		c.Assert(assessment.SlopProbability, qt.IsNil)
		c.Assert(assessment.ProbabilityStatus, qt.Equals, "calibration_unavailable")
	}
	_, err = engine.AnalyzeAll(t.Context(), nil)
	c.Assert(err, qt.ErrorMatches, "scan contains no applicable English prose")
}

func TestSourceAndScoringInvariants(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	texts := []string{
		"It is important to note that the client can retry the request.",
		"It is **important** to note that the client can retry the request.",
		"It is important to note that the client can retry the request.\n\n```\n" + strings.Repeat("code ", 200) + "\n```",
		"It is important to note that the client can retry the request.\n\nA clean paragraph describes the configuration.",
	}
	for _, text := range texts {
		result, err := engine.Analyze(t.Context(), document.Source{Name: "draft.md", Format: document.Markdown, Bytes: []byte(text)})
		c.Assert(err, qt.IsNil)
		c.Assert(result.Findings[0].RuleID, qt.Equals, "filler.announced-importance")
		c.Assert(result.Assessments[0].SlopScore, qt.Equals, float64(15))
	}
	result, err := engine.Analyze(
		t.Context(),
		document.Source{
			Name:   "draft.md",
			Format: document.Markdown,
			Bytes:  []byte("It is `important` to note that the client can retry the request."),
		},
	)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 0)
}

func TestSeverityAndInactiveRuleDoNotChangeScore(t *testing.T) {
	cases := []string{
		"version: 1\nrules:\n  filler.announced-importance:\n    severity: error\n",
		"version: 1\nrules:\n  scaffold.ai-self-reference:\n    enabled: false\n",
	}
	for _, configuration := range cases {
		c := qt.New(t)
		engine, err := unswell.New(unswell.Options{Config: []byte(configuration)})
		c.Assert(err, qt.IsNil)
		result, err := engine.Analyze(
			t.Context(),
			document.Source{
				Name:   "a.txt",
				Format: document.Plain,
				Bytes:  []byte("It is important to note that the client can retry the request."),
			},
		)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Assessments[0].SlopScore, qt.Equals, float64(15))
	}
}

func TestDeterministicWorkersAndConcurrentEngine(t *testing.T) {
	c := qt.New(t)
	sources := []document.Source{
		{Name: "z.txt", Format: document.Plain, Bytes: []byte("Certainly! The client opens connections.")},
		{Name: "a.md", Format: document.Markdown, Bytes: []byte("Let's dive into the configuration.")},
	}
	single, err := unswell.New(unswell.Options{Jobs: 1})
	c.Assert(err, qt.IsNil)
	want, err := single.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	parallel, err := unswell.New(unswell.Options{Jobs: 4})
	c.Assert(err, qt.IsNil)
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			result, analyzeErr := parallel.AnalyzeAll(t.Context(), []document.Source{sources[1], sources[0]})
			c.Check(analyzeErr, qt.IsNil)
			c.Check(result, qt.DeepEquals, want)
		})
	}
	workers.Wait()
}

func TestStrictConfigAndCancellation(t *testing.T) {
	c := qt.New(t)
	invalid := []string{
		"version: 1\nunknown: true\n",
		"version: 1\nrules:\n  typo.rule:\n    enabled: false\n",
		"version: 1\nrules:\n  filler.wordy-phrase:\n    parameters:\n      similarty: 0.5\n",
		"version: 1\nrules:\n  filler.wordy-phrase:\n    enabled: true\n    enabled: false\n",
		"version: 1\ncalibration:\n  model: builtin:technical-v1\n",
		"version: 1\nrules:\n  syntax.not-only-density:\n    parameters:\n      window_sentences: 0\n",
	}
	for _, configuration := range invalid {
		_, err := unswell.New(unswell.Options{Config: []byte(configuration)})
		c.Check(err, qt.IsNotNil, qt.Commentf("%s", configuration))
	}
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := engine.Analyze(ctx, document.Source{Name: "a.txt", Format: document.Plain, Bytes: []byte("Clean prose.")})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result.Status, qt.Equals, "incomplete")
	c.Assert(result.Gate.Passed, qt.IsFalse)
}

func TestIndependentResultOwnership(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "a.txt", Format: document.Plain, Bytes: []byte("Certainly! A clean sentence follows.")}
	first, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	want, err := json.Marshal(first)
	c.Assert(err, qt.IsNil)
	first.Manifest.Rules[0].Defaults.Parameters.Phrases = append(first.Manifest.Rules[0].Defaults.Parameters.Phrases, "mutation")
	first.Manifest.Rules[0].Summary = "mutated"
	second, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	got, err := json.Marshal(second)
	c.Assert(err, qt.IsNil)
	c.Assert(string(got), qt.Equals, string(want))
}
