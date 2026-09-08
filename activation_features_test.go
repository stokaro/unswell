package unswell_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/rule"
)

type observationRule struct {
	id   string
	full bool
	run  func(rule.View, rule.Emitter) error
}

func (r observationRule) Descriptor() rule.Descriptor {
	return rule.Descriptor{ID: r.id, Version: "1", Group: "test", Scope: "paragraph", BlockObservations: r.full,
		Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences},
		Defaults: rule.Settings{Enabled: true, Severity: "warning", Gate: "none"}}
}

func (r observationRule) Evaluate(_ context.Context, view rule.View, emit rule.Emitter) error {
	if r.run == nil {
		return nil
	}
	return r.run(view, emit)
}

func emitObserved(view rule.View, emit rule.Emitter, blockID, activation int) error {
	block := view.Document.Blocks[blockID]
	sentence := block.Sentences[0]
	return emit.Emit(rule.Evidence{Kind: "heuristic", Activation: activation, Occurrences: []rule.Occurrence{
		{BlockID: block.ID, SentenceID: sentence.ID, Spans: sentence.Tokens[0].Spans},
	}})
}

func TestActivationCollectionPreservesZeroUnknownAndMaximum(t *testing.T) {
	c := qt.New(t)
	observed := observationRule{id: "test.observed", full: true, run: func(v rule.View, emit rule.Emitter) error {
		for _, block := range v.Document.Blocks {
			if err := v.Observe(feature.BlockObservation{BlockID: block.ID, Status: "evaluated"}); err != nil {
				return err
			}
		}
		for _, amount := range []int{250, 750, 750} {
			if err := emitObserved(v, emit, 0, amount); err != nil {
				return err
			}
		}
		return nil
	}}
	unknown := observationRule{id: "test.unknown"}
	engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{observed, unknown},
		Features: []string{"activation/test.unknown", "activation/test.observed", "prose-words"}})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("The cache expires.\n\nThe client retries.\n")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Features.ActivationContract, qt.Equals, feature.ActivationContract)
	c.Assert(result.Features.Sources[0].RulesetHash, qt.Equals, result.Manifest.RulesetHash)
	units := result.Features.Sources[0].Units
	c.Assert(units, qt.HasLen, 2)
	c.Assert(*units[0].Values[0].Number, qt.Equals, 0.75)
	c.Assert(*units[1].Values[0].Number, qt.Equals, float64(0))
	c.Assert(units[0].Values[1].Number, qt.IsNil)
	c.Assert(units[0].Values[1].Reason, qt.Equals, "applicability_unknown")
	c.Assert(*units[0].Values[2].Number, qt.Equals, float64(3))
}

func TestActivationCollectionDoesNotEnableRules(t *testing.T) {
	c := qt.New(t)
	calls := 0
	implementation := observationRule{id: "test.disabled", run: func(rule.View, rule.Emitter) error { calls++; return nil }}
	engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{implementation}, Features: []string{"activation/test.disabled"},
		Config: []byte("version: 1\nrules:\n  test.disabled: {enabled: false}\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("The cache expires.")})
	c.Assert(err, qt.IsNil)
	c.Assert(calls, qt.Equals, 0)
	value := result.Features.Sources[0].Units[0].Values[0]
	c.Assert(value.Number, qt.IsNil)
	c.Assert(value.Reason, qt.Equals, "disabled")
	_, err = unswell.New(unswell.Options{Features: []string{"activation/not.registered"}})
	c.Assert(err, qt.ErrorMatches, `unknown requested feature "activation/not.registered"`)
}

func TestActivationCollectionRejectsIgnoredObserverErrors(t *testing.T) {
	for _, full := range []bool{false, true} {
		c := qt.New(t)
		broken := observationRule{id: "test.broken", full: full, run: func(v rule.View, _ rule.Emitter) error {
			if !full {
				_ = v.Observe(feature.BlockObservation{BlockID: 99, Status: "evaluated"})
			}
			return nil
		}}
		later := observationRule{id: "test.later"}
		engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{broken, later}, NoGate: true,
			Features: []string{"activation/test.broken", "activation/test.later"}})
		c.Assert(err, qt.IsNil)
		result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
			Bytes: []byte("The cache expires.")})
		c.Assert(err, qt.IsNotNil)
		c.Assert(result.Manifest.Complete, qt.IsFalse)
		c.Assert(result.Gate.Passed, qt.IsFalse)
		values := result.Features.Sources[0].Units[0].Values
		c.Assert(values[0].Number, qt.IsNil)
		c.Assert(values[0].Reason, qt.Equals, "evaluation_failed")
		c.Assert(values[1].Number, qt.IsNil)
		c.Assert(values[1].Reason, qt.Equals, "not_evaluated")
	}
}

func TestActivationCollectionDiscardsFailedEvaluationValues(t *testing.T) {
	c := qt.New(t)
	broken := observationRule{id: "test.broken", run: func(v rule.View, emit rule.Emitter) error {
		if err := emitObserved(v, emit, 0, 1000); err != nil {
			return err
		}
		return fmt.Errorf("deliberate failure after evidence")
	}}
	engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{broken}, Features: []string{"activation/test.broken"}})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("The cache expires.")})
	c.Assert(err, qt.IsNotNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Features.Sources[0].Units[0].Values[0].Number, qt.IsNil)
	c.Assert(result.Features.Sources[0].Units[0].Values[0].Reason, qt.Equals, "evaluation_failed")
}

func TestReadabilityActivationApplicabilityComesFromMeasuredInput(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{"activation/readability.grade-metric"}, Config: []byte(
		"version: 1\nextends: [builtin:custom]\nrules:\n  readability.grade-metric: {enabled: true}\n")})
	c.Assert(err, qt.IsNil)
	text := "# Heading\n\nThe cache expires.\n\n" + strings.Repeat("The client opens a link. The server sends a reply. ", 6) + "\n\n" +
		strings.Repeat("The implementation requires comprehensive configuration, systematic verification, "+
			"and consistent documentation of operational prerequisites before production deployment. ", 4)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	units := result.Features.Sources[0].Units
	c.Assert(units, qt.HasLen, 4)
	c.Assert(units[0].Values[0].Reason, qt.Equals, "inapplicable/unsupported_unit")
	c.Assert(units[1].Values[0].Reason, qt.Equals, "inapplicable/insufficient_words")
	c.Assert(*units[2].Values[0].Number, qt.Equals, float64(0))
	c.Assert(*units[3].Values[0].Number > 0, qt.IsTrue)
	c.Assert(result.Findings, qt.HasLen, 1)
}

func TestActivationCollectionKeepsSuppressedEvidenceAndOwnsResults(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{"activation/filler.announced-importance"}})
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(
		"<!-- unswell-disable-next-block filler.announced-importance -- Required external wording. -->\n" +
			"It is important to note that the client retries.\n")}
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(want.Gate.Passed, qt.IsTrue)
	c.Assert(want.Findings, qt.HasLen, 1)
	c.Assert(want.Findings[0].Suppressed, qt.IsTrue)
	blockID := want.Findings[0].Evidence.Occurrences[0].BlockID
	c.Assert(*want.Features.Sources[0].Units[blockID].Values[0].Number, qt.Equals, float64(1))
	for range 4 {
		t.Run("owned activation", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			got, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.DeepEquals, want)
			*got.Features.Sources[0].Units[blockID].Values[0].Number = -1
		})
	}
}

func TestDisabledActivationDoesNotRequireUnexecutedPOS(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Features: []string{"activation/readability.grade-metric"},
		NLP: limitedPolicyNLP{Provider: provider}, Config: []byte("version: 1\nextends: [builtin:custom]\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("The cache expires.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Features.Sources[0].Capabilities, qt.Not(qt.Contains), nlp.POS)
	c.Assert(result.Features.Sources[0].Units[0].Values[0].Reason, qt.Equals, "disabled")
}

func TestActivationCollectionBoundsOutputBeforeAllocatingValues(t *testing.T) {
	c := qt.New(t)
	var implementations []rule.Rule
	var ids []string
	for i := range 10 {
		id := fmt.Sprintf("test.rule%d", i)
		implementations = append(implementations, observationRule{id: id})
		ids = append(ids, "activation/"+id)
	}
	engine, err := unswell.New(unswell.Options{Rules: implementations, Features: ids,
		Config: []byte("version: 1\nanalysis:\n  max_candidates: 20\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("One.\n\nTwo.\n\nThree.\n")})
	c.Assert(err, qt.ErrorMatches, "collected feature values exceed max_candidates")
	c.Assert(result.Manifest.Complete, qt.IsFalse)
	c.Assert(result.Features.Sources, qt.HasLen, 0)
}

func TestActivationCollectionRejectsIgnoredCancellation(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	implementation := observationRule{id: "test.canceled", run: func(v rule.View, _ rule.Emitter) error {
		if err := v.Observe(feature.BlockObservation{Status: "evaluated"}); err != nil {
			return err
		}
		cancel()
		return nil
	}}
	engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{implementation}, Features: []string{"activation/test.canceled"}})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(ctx, document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("The cache expires.")})
	c.Assert(err, qt.IsNotNil)
	c.Assert(result.Manifest.Complete, qt.IsFalse)
	c.Assert(result.Features.Sources[0].Units[0].Values[0].Number, qt.IsNil)
	c.Assert(result.Features.Sources[0].Units[0].Values[0].Reason, qt.Equals, "evaluation_failed")
}

func TestReadabilityActivationRetainsDataMinimumReasons(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{"activation/readability.grade-metric"}, AllowEmpty: true,
		Config: []byte("version: 1\nextends: [builtin:custom]\nrules:\n  readability.grade-metric: {enabled: true}\n")})
	c.Assert(err, qt.IsNil)
	for _, row := range []struct{ text, reason string }{
		{".", "inapplicable/no_prose_words"},
		{strings.Repeat("cache ", 50) + "expires.", "inapplicable/insufficient_sentences"},
	} {
		result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.txt", Format: document.Plain, Bytes: []byte(row.text)})
		c.Assert(err, qt.IsNil)
		c.Assert(result.Features.Sources[0].Units, qt.HasLen, 1)
		value := result.Features.Sources[0].Units[0].Values[0]
		c.Assert(value.Number, qt.IsNil)
		c.Assert(value.Reason, qt.Equals, row.reason)
	}
}
