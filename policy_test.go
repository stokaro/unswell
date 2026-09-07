package unswell_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func TestFilePolicyAndExactTermExemptions(t *testing.T) {
	c := qt.New(t)
	policy := []byte(`version: 1
extends: [builtin:custom]
rules:
  policy.banned-phrases:
    enabled: true
    parameters: {phrases: [robust]}
vocabulary:
  terms: [robust estimator]
  term_exemptions: [policy.banned-phrases]
overrides:
  - files: [docs/reference/**]
    rules: {policy.banned-phrases: {enabled: false}}
`)
	engine, err := unswell.New(unswell.Options{Config: policy, Jobs: 4, IncludeSource: true})
	c.Assert(err, qt.IsNil)
	text := "The robust estimator uses robust methods."
	sources := []document.Source{
		{Name: "docs/guide.md", Format: document.Markdown, Bytes: []byte(text)},
		{Name: "docs/reference/api.md", Format: document.Markdown, Bytes: []byte(text)},
	}
	result, err := engine.AnalyzeAll(context.Background(), sources)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Path, qt.Equals, "docs/guide.md")
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, strings.LastIndex(text, "robust"))
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Documents[0].Source, qt.Equals, text)
	for _, doc := range result.Documents {
		resolved, err := engine.PolicyForFile(doc.Name)
		c.Assert(err, qt.IsNil)
		c.Assert(doc.ConfigHash, qt.Equals, resolved.Hash)
	}
	serial, err := unswell.New(unswell.Options{Config: policy, Jobs: 1, IncludeSource: true})
	c.Assert(err, qt.IsNil)
	again, err := serial.AnalyzeAll(context.Background(), sources)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, result)
	for i := range 4 {
		t.Run(fmt.Sprintf("concurrent-%d", i), func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			actual, err := engine.AnalyzeAll(t.Context(), sources)
			c.Assert(err, qt.IsNil)
			c.Assert(actual, qt.DeepEquals, result)
			policy, err := engine.PolicyForFile("docs/guide.md")
			c.Assert(err, qt.IsNil)
			settings := policy.Rules["policy.banned-phrases"]
			settings.Enabled = false
			policy.Rules["policy.banned-phrases"] = settings
			policy.Vocabulary.ResolvedTerms[0] = "changed"
		})
	}
}

type limitedPolicyNLP struct{ nlp.Provider }

func (p limitedPolicyNLP) Identity() nlp.Identity {
	identity := p.Provider.Identity()
	identity.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences}
	return identity
}

func TestFileOverrideRequiresAvailableNLP(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	_, err = unswell.New(unswell.Options{NLP: limitedPolicyNLP{Provider: provider}, Config: []byte(`version: 1
extends: [builtin:custom]
overrides:
  - files: [docs/**]
    rules: {hype.modifier-cluster: {enabled: true}}
`)})
	c.Assert(err, qt.ErrorMatches, "rule hype.modifier-cluster requires unavailable capability pos")
}

func TestTermExemptionsRecomputeAggregateCounts(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(`version: 1
extends: [builtin:custom]
rules:
  density.connective-overuse:
    enabled: true
    parameters: {phrases: [moreover, furthermore], min_words: 0, onset: 1, saturation: 4}
vocabulary:
  terms: [moreover]
  term_exemptions: [density.connective-overuse]
`)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(context.Background(), document.Source{Name: "api.txt", Format: document.Plain,
		Bytes: []byte("Moreover, the server starts. Furthermore, it reads the configuration. Furthermore, it accepts connections.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Evidence.Metrics[0].Value, qt.Equals, float64(2))
	c.Assert(result.Findings[0].Evidence.Occurrences, qt.HasLen, 2)
	c.Assert(result.Findings[0].Evidence.Activation, qt.Equals, 333)
}

func TestAllFilePoliciesValidatedBeforeAnalysis(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(`version: 1
overrides:
  - files: ['*.md']
    analysis: {max_file_bytes: 4}
`)})
	c.Assert(err, qt.IsNil)
	result, err := engine.AnalyzeAll(context.Background(), []document.Source{
		{Name: "a.txt", Format: document.Plain, Bytes: []byte("The server starts.")},
		{Name: "z.md", Format: document.Markdown, Bytes: []byte("The server starts.")},
	})
	c.Assert(err, qt.ErrorMatches, ".*max_file_bytes")
	c.Assert(result.Documents, qt.HasLen, 0)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	_, err = unswell.New(unswell.Options{Config: []byte("version: 1"), ConfigBundle: &config.Bundle{}})
	c.Assert(err, qt.ErrorMatches, "config bytes and ConfigBundle cannot be combined")
}
