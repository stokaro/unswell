package ruleset_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/rule"
	"github.com/stokaro/unswell/ruleset"
)

func TestMetadataAndStrictStructure(t *testing.T) {
	t.Parallel()
	valid := string(definition("type: phrase\nvalues: [seamless]", ""))
	for _, mutate := range []func(string) string{
		func(s string) string { return strings.Replace(s, "version: 1", "version: 2", 1) },
		func(s string) string { return strings.Replace(s, "license: MIT", "license: null", 1) },
		func(s string) string { return strings.Replace(s, "company.wording", "other.wording", 1) },
		func(s string) string { return strings.Replace(s, "scope: sentence", "scope: token", 1) },
		func(s string) string { return s + "extra: true\n" },
		func(s string) string { return s + "---\nversion: 1\n" },
		func(s string) string {
			return strings.Replace(s, "license: MIT", "license: &license MIT\nprovenance: *license", 1)
		},
		func(s string) string {
			return strings.Replace(s, "match:", "requires: [tokens, sentences, dependencies]\n    match:", 1)
		},
	} {
		_, err := ruleset.Load([]byte(mutate(valid)))
		qt.New(t).Assert(err, qt.IsNotNil)
	}
}

func TestBuiltinShadowingIsAnError(t *testing.T) {
	t.Parallel()
	data := string(definition("type: phrase\nvalues: [seamless]", ""))
	data = strings.ReplaceAll(data, "company.wording", "scaffold.chat-preamble")
	data = strings.ReplaceAll(data, "namespace: company", "namespace: scaffold")
	_, err := unswell.New(unswell.Options{RuleSets: [][]byte{[]byte(data)}})
	qt.New(t).Assert(err, qt.ErrorMatches, `duplicate rule ID "scaffold.chat-preamble"`)
}

func TestBudgetAndCancellationDoNotPass(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	set, err := ruleset.Load(definition("type: sequence\ntokens: [{value: a}, {gap: {max: 32}}, {value: missing}]", ""))
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Rules: set.Rules(), Config: []byte("version: 1\nanalysis:\n  max_candidates: 3\n")})
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "input.txt", Format: document.Plain, Bytes: []byte("a client returns a response.")}
	result, err := engine.Analyze(context.Background(), source)
	c.Assert(err, qt.IsNotNil)
	c.Assert(result.Status, qt.Equals, "incomplete")
	c.Assert(result.Gate.Passed, qt.IsFalse)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err = engine.Analyze(ctx, source)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result.Gate.Passed, qt.IsFalse)
}

type tokensOnly struct{ inner *english.Provider }

func (p tokensOnly) Identity() nlp.Identity {
	identity := p.inner.Identity()
	identity.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences}
	return identity
}

func (p tokensOnly) Analyze(ctx context.Context, text document.MappedText, caps []nlp.Capability) ([]document.Sentence, error) {
	return p.inner.Analyze(ctx, text, caps)
}

func TestMissingPOSFailsAtConstruction(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	set, err := ruleset.Load(definition("type: sequence\ntokens: [{pos: JJ}]", ""))
	c.Assert(err, qt.IsNil)
	_, err = unswell.New(unswell.Options{Rules: set.Rules(), NLP: tokensOnly{inner: provider}})
	c.Assert(err, qt.ErrorMatches, "rule company.wording requires unavailable capability pos")
}

func TestCompiledRulesAndReportsOwnTheirMetadata(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	data := definition("type: phrase\nvalues: [seamless]", "")
	set, err := ruleset.Load(data)
	c.Assert(err, qt.IsNil)
	registry := set.Rules()
	descriptor := registry[0].Descriptor()
	descriptor.Origin.Hash = "changed"
	descriptor.Contexts[0] = "changed"
	descriptor.Examples[0].Text = "changed"
	for i := range data {
		data[i] = 0
	}
	engine, err := unswell.New(unswell.Options{Rules: registry})
	c.Assert(err, qt.IsNil)
	registry[0] = nil
	source := document.Source{Name: "input.txt", Format: document.Plain, Bytes: []byte("This seamless client works.")}
	before, err := engine.Analyze(context.Background(), source)
	c.Assert(err, qt.IsNil)
	encoded, err := json.Marshal(before)
	c.Assert(err, qt.IsNil)
	before.Manifest.Rules[0].Origin.Hash = "changed"
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			result, analyzeErr := engine.Analyze(context.Background(), source)
			if analyzeErr != nil {
				t.Error(analyzeErr)
				return
			}
			actual, marshalErr := json.Marshal(result)
			if marshalErr != nil || string(actual) != string(encoded) {
				t.Error("concurrent result changed or failed serialization", marshalErr)
			}
		})
	}
	workers.Wait()
}

func TestScopedDensityAndNoDilutionByProtectedCode(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	data := definition("type: density\nper: words\nmin: 20\nmatch: {type: token-set, values: [seamless]}", "")
	plain := check(t, data, "This seamless client works.", document.Markdown)
	withCode := check(t, data, "This seamless `seamless code words here` client works.\n\n```text\nMany extra words in code.\n```\n",
		document.Markdown)
	c.Assert(withCode.Findings, qt.HasLen, 1)
	c.Assert(withCode.Findings[0].Evidence.Metrics, qt.DeepEquals, plain.Findings[0].Evidence.Metrics)
	withClean := check(t, data, "This seamless client works.\n\nThe service rejects requests without a valid session.",
		document.Markdown)
	c.Assert(withClean.Findings[0].Evidence.Metrics, qt.DeepEquals, plain.Findings[0].Evidence.Metrics)
	for _, scope := range []string{"paragraph", "document"} {
		count := string(definition("type: count\nmin: 2\nmatch: {type: token-set, values: [seamless]}", ""))
		count = strings.Replace(count, "scope: sentence", "scope: "+scope, 1)
		result := check(t, []byte(count), "A seamless client works. A seamless server responds.", document.Plain)
		c.Assert(result.Findings, qt.HasLen, 1)
		c.Assert(result.Findings[0].Related, qt.HasLen, 1)
	}
}

func TestContextSelectionAndFormattedExamples(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	data := string(definition("type: phrase\nvalues: [seamless]", "    contexts: [comment]\n"))
	data = strings.Replace(data, `fail: ["Let's dive into the configuration."]`,
		`fail: [{format: go, text: "package sample\n// This seamless client works.\n"}]`, 1)
	set, err := ruleset.Load([]byte(data))
	c.Assert(err, qt.IsNil)
	c.Assert(set.Rules()[0].Descriptor().Examples[0].Format, qt.Equals, document.Go)
	source := "package sample\n// This seamless client works.\nconst message = \"This seamless client works.\"\n"
	result := check(t, []byte(data), source, document.Go)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Start.Line, qt.Equals, 2)
}

var _ rule.Rule = (*markerRule)(nil)

// markerRule is used to verify declarative registration composes with Go rules.
type markerRule struct{}

func (*markerRule) Descriptor() rule.Descriptor {
	return rule.Descriptor{ID: "test.marker", Version: "1", Group: "custom", Scope: "sentence",
		Defaults: rule.Settings{Enabled: true, Severity: "warning", Gate: "none"}}
}

func (*markerRule) Evaluate(context.Context, rule.View, rule.Emitter) error { return nil }

func TestOptionsComposeDeclarativeAndGoRules(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{&markerRule{}},
		RuleSets: [][]byte{definition("type: phrase\nvalues: [seamless]", "")}})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(context.Background(), document.Source{Name: "input.txt", Format: document.Plain,
		Bytes: []byte("This seamless client works.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Manifest.Rules, qt.HasLen, 2)
	c.Assert(result.Findings, qt.HasLen, 1)
}
