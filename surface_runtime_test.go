package unswell_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func TestSurfaceLimitsAndCapabilities(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !surfaceRuleID(d.ID) {
			continue
		}
		engine := singleRuleEngine(t, d.ID, "", "analysis: {max_candidates: 1}\n")
		result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
			Bytes: []byte(d.Examples[0].Text)})
		c.Assert(err, qt.ErrorMatches, ".*max_candidates.*", qt.Commentf("%s", d.ID))
		c.Assert(result.Gate.Passed, qt.IsFalse)
		if d.ID == "syntax.parenthetical-load" || d.ID == "format.em-dash-density" {
			continue
		}
		config := []byte("version: 1\nextends: [builtin:custom]\nrules:\n  " + d.ID + ": {enabled: true}\n")
		_, err = unswell.New(unswell.Options{Config: config, NLP: limitedPolicyNLP{Provider: provider}})
		c.Assert(err, qt.ErrorMatches, "rule "+d.ID+" requires unavailable capability pos")
	}
}

func TestSurfaceParameterValidation(t *testing.T) {
	for _, row := range []struct{ id, parameters string }{
		{"syntax.nominalization-chain", "verbs: [perform evaluation]"},
		{"syntax.nominalization-chain", "nouns: [evaluation!]"},
		{"syntax.parenthetical-load", "allowed_depth: 17"},
		{"syntax.parenthetical-load", "allowed_depth: 4, saturation_depth: 4"},
		{"readability.grade-metric", "min_sentences: 0"},
		{"readability.long-paragraph", "sentence_words: 0"},
		{"readability.long-paragraph", "min_long_sentences: 0"},
		{"format.list-fragmentation", "max_item_words: 0"},
		{"format.list-fragmentation", "max_list_items: 101"},
		{"format.em-dash-density", "allowed_occurrences: -1"},
		{"syntax.noun-stack", "verbs: [perform evaluation]"},
		{"syntax.parenthetical-load", "min_insertion_words: 101"},
		{"syntax.parenthetical-load", "min_insertion_words: -1"},
	} {
		t.Run(row.id+"/"+row.parameters, func(t *testing.T) {
			c := qt.New(t)
			config := "version: 1\nrules:\n  " + row.id + ": {enabled: true, parameters: {" + row.parameters + "}}\n"
			_, err := unswell.New(unswell.Options{Config: []byte(config)})
			c.Assert(err, qt.IsNotNil)
		})
	}
}

type malformedChunkNLP struct{ nlp.Provider }

func (p malformedChunkNLP) Analyze(ctx context.Context, text document.MappedText, required []nlp.Capability) ([]document.Sentence, error) {
	sentences, err := p.Provider.Analyze(ctx, text, required)
	if err == nil && len(sentences) > 0 {
		sentences[0].Chunks = []document.Chunk{{Kind: "NP", FirstToken: -1, EndToken: 500}}
	}
	return sentences, err
}

func TestSurfaceRejectsMalformedChunksAndExcessiveNesting(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{NLP: malformedChunkNLP{Provider: provider},
		Config: []byte("version: 1\nextends: [builtin:custom]\nrules:\n  syntax.noun-stack: {enabled: true}\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "sample.txt", Format: document.Plain, Bytes: []byte(nounProse)})
	c.Assert(err, qt.ErrorMatches, ".*invalid NP chunk token range.*")
	c.Assert(result.Gate.Passed, qt.IsFalse)
	engine = singleRuleEngine(t, "syntax.parenthetical-load", "{min_words: 1}", "")
	text := "The client " + strings.Repeat("(", 257) + "may retry" + strings.Repeat(")", 257) + " after failure."
	result, err = engine.Analyze(t.Context(), document.Source{Name: "sample.txt", Format: document.Plain, Bytes: []byte(text)})
	c.Assert(err, qt.ErrorMatches, ".*nesting exceeds 256 levels.*")
	c.Assert(result.Gate.Passed, qt.IsFalse)
}

func TestSurfaceConcurrentReuseAndCancellation(t *testing.T) {
	c := qt.New(t)
	engine := singleRuleEngine(t, "format.list-fragmentation", "", "")
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(fragmentedLists)}
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := engine.Analyze(ctx, source)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	for range 4 {
		t.Run("shared engine", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			got, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.DeepEquals, want)
			c.Assert(string(source.Bytes), qt.Equals, fragmentedLists)
		})
	}
}

func TestSurfaceDictionariesAreImmutable(t *testing.T) {
	c := qt.New(t)
	engine := singleRuleEngine(t, "syntax.nominalization-chain", "", "")
	policy, err := engine.PolicyForFile("guide.md")
	c.Assert(err, qt.IsNil)
	policy.Rules["syntax.nominalization-chain"].Parameters.Verbs[0] = "corrupted"
	policy.Rules["syntax.nominalization-chain"].Parameters.Nouns[2] = "corrupted"
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(nominalProse)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	engine = singleRuleEngine(t, "syntax.noun-stack", "", "")
	policy, err = engine.PolicyForFile("guide.md")
	c.Assert(err, qt.IsNil)
	// A leaked "request" would end the run after "service" and silence the stack.
	policy.Rules["syntax.noun-stack"].Parameters.Verbs[0] = "request"
	result, err = engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(nounProse)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
}
