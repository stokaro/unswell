package ruleset_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/rule"
	"github.com/stokaro/unswell/ruleset"
)

func FuzzLoad(f *testing.F) {
	f.Add(definition("type: phrase\nvalues: [seamless]", ""))
	f.Add(definition("type: sequence\ntokens: [{pos: JJ}, {gap: {max: 3}}, {value: client}]", ""))
	f.Add(definition("type: regex\npattern: '[a-z]+'", ""))
	f.Add([]byte("---\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		set, err := ruleset.Load(data)
		if err != nil {
			return
		}
		c := qt.New(t)
		c.Assert(len(set.Rules()) > 0 && len(set.Rules()) <= 100, qt.IsTrue)
		c.Assert(set.Origin().Hash, qt.HasLen, 64)
		for _, implementation := range set.Rules() {
			encoded, marshalErr := json.Marshal(implementation.Descriptor())
			c.Assert(marshalErr, qt.IsNil)
			c.Assert(json.Valid(encoded), qt.IsTrue)
		}
	})
}

func TestDefinitionResourceLimits(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, matcher string }{
		{"regex bytes", "type: regex\npattern: '" + strings.Repeat("a", 4097) + "'"},
		{"value bytes", "type: phrase\nvalues: ['" + strings.Repeat("a", 1001) + "']"},
		{"deep matcher", strings.Repeat("{type: not, match: ", 18) + "{type: phrase, values: [word]}" + strings.Repeat("}", 18)},
		{"sequence length", "{type: sequence, tokens: [" + strings.Repeat("{value: a},", 33) + "]}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := ruleset.Load(definition(tc.matcher, ""))
			c.Assert(err, qt.IsNotNil)
		})
	}
	c := qt.New(t)
	_, err := ruleset.Load([]byte(strings.Repeat(" ", 1<<20+1)))
	c.Assert(err, qt.ErrorMatches, "ruleset must contain 1 byte to 1 MiB")
}

type failingEmitter struct{ err error }

// Emit injects the caller's failure at the rule boundary.
func (e failingEmitter) Emit(rule.Evidence) error { return e.err }

func proseDocument(t *testing.T, text string) *document.Document {
	t.Helper()
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	mapped := document.MappedText{Text: text}
	for i := range len(text) {
		mapped.Map = append(mapped.Map, document.Span{Start: i, End: i + 1})
	}
	sentences, err := provider.Analyze(t.Context(), mapped, []nlp.Capability{nlp.Tokens, nlp.Sentences})
	c.Assert(err, qt.IsNil)
	return &document.Document{Blocks: []document.Block{{Kind: "paragraph", MappedText: mapped, Sentences: sentences}}}
}

func TestEmitterFailurePropagates(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	set, err := ruleset.Load(definition("type: phrase\nvalues: [seamless]", ""))
	c.Assert(err, qt.IsNil)
	want := errors.New("writer rejected evidence")
	doc := proseDocument(t, "A seamless client works.")
	err = set.Rules()[0].Evaluate(t.Context(), rule.View{Document: doc, MaxCandidates: 100}, failingEmitter{err: want})
	c.Assert(err, qt.ErrorIs, want)
}

func TestWorkersPreserveCustomRuleResults(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	data := definition("type: count\nmin: 1\nmatch: {type: token-set, values: [seamless]}", "")
	options := unswell.Options{Rules: []rule.Rule{}, RuleSets: [][]byte{data}, Jobs: 1}
	sequential, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	options.Jobs = 4
	parallel, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	sources := []document.Source{
		{Name: "z.txt", Format: document.Plain, Bytes: []byte("The seamless client responds.")},
		{Name: "a.txt", Format: document.Plain, Bytes: []byte("The seamless server responds.")},
	}
	want, err := sequential.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	got, err := parallel.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	c.Assert(got, qt.DeepEquals, want)
}

func TestShortCircuitPreservesTheWorkBudget(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, match, extra string
		findings           int
	}{
		{"any", "type: any\nitems:\n  - {type: feature, name: prose.words, min: 1}\n" +
			"  - {type: phrase, values: [absent]}", "", 1},
		{"all", "type: all\nitems:\n  - {type: feature, name: prose.words, max: 0}\n" +
			"  - {type: phrase, values: [absent]}", "", 0},
		{"exception", "type: phrase\nvalues: [absent]",
			"    except: [{type: feature, name: prose.words, min: 1}]\n", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{},
				RuleSets: [][]byte{definition(tc.match, tc.extra)},
				Config:   []byte("version: 1\nanalysis:\n  max_candidates: 3\n")})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: "input.txt", Format: document.Plain,
				Bytes: []byte("The client reads the configuration file.")})
			c.Assert(err, qt.IsNil)
			c.Assert(result.Status, qt.Equals, "complete")
			c.Assert(result.Findings, qt.HasLen, tc.findings)
		})
	}
}

func TestChunkFeaturesRemainHeuristic(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	result := check(t, definition("type: feature\nname: chunks.np\nmin: 1", ""),
		"The client opens a connection.", document.Plain)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Evidence.Kind, qt.Equals, "heuristic")
	c.Assert(result.Findings[0].Evidence.Metrics[0].Name, qt.Equals, "chunks.np")
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, 0)
}

func BenchmarkSequenceScan(b *testing.B) {
	data := definition("type: sequence\ntokens: [{value: the}, {gap: {max: 4}}, {value: responds}]", "")
	engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{}, RuleSets: [][]byte{data}})
	if err != nil {
		b.Fatal(err)
	}
	source := document.Source{Name: "input.txt", Format: document.Plain,
		Bytes: []byte(strings.Repeat("The client responds to a request. ", 100))}
	b.ReportAllocs()
	b.SetBytes(int64(len(source.Bytes)))
	b.ResetTimer()
	for range b.N {
		if _, err := engine.Analyze(context.Background(), source); err != nil {
			b.Fatal(err)
		}
	}
}
