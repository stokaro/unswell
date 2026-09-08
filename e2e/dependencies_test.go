package e2e_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

type dependencyTraces struct {
	Schema    string            `json:"schema"`
	Provider  string            `json:"provider"`
	Revision  string            `json:"revision"`
	Scheme    string            `json:"scheme"`
	ModelHash string            `json:"model_archive_sha256"`
	Sources   []dependencyTrace `json:"sources"`
}

type dependencyTrace struct {
	ID        string              `json:"id"`
	Format    document.Format     `json:"format"`
	Text      string              `json:"text"`
	Sentences []document.Sentence `json:"sentences"`
}

// These are frozen model predictions, not human grammatical or quality labels.
// The probe rule checks how actual parser evidence reaches the public engine.
func TestDependencyTraceReplay(t *testing.T) {
	var traces dependencyTraces
	decodeFile(t, "testdata/dependency-traces.json", &traces)
	c := qt.New(t)
	c.Assert(traces.Schema, qt.Equals, "unswell-dependency-replay-v1")
	c.Assert(traces.Sources, qt.HasLen, 10)
	wants := map[string]string{
		"technical-1.txt": "request", "technical-3.txt": "server", "markup.md": "request",
		"protected.md": "request", "comment.go": "request",
	}
	for _, source := range traces.Sources {
		t.Run(source.ID, func(t *testing.T) {
			c := qt.New(t)
			provider := replayProvider(t, traces, source)
			engine, err := unswell.New(unswell.Options{NLP: provider, Rules: []rule.Rule{dependencyRelationProbe{}}})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: source.ID, Format: source.Format, Bytes: []byte(source.Text)})
			c.Assert(err, qt.IsNil)
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Gate.Passed, qt.IsTrue)
			c.Assert(result.Manifest.NLP.DependencyScheme, qt.Equals, traces.Scheme)
			want := wants[source.ID]
			if want == "" {
				c.Assert(result.Findings, qt.HasLen, 0)
			} else {
				c.Assert(result.Findings, qt.HasLen, 1)
				span := result.Findings[0].Primary.Span
				c.Assert(source.Text[span.Start:span.End], qt.Equals, want)
			}
			verifyLocations(t, result, map[string][]byte{source.ID: []byte(source.Text)})
		})
	}
}

type dependencyReplay struct {
	identity nlp.Identity
	blocks   []document.Block
	source   dependencyTrace
}

func replayProvider(t *testing.T, traces dependencyTraces, source dependencyTrace) dependencyReplay {
	t.Helper()
	c := qt.New(t)
	input := document.Source{Name: source.ID, Format: source.Format, Bytes: []byte(source.Text)}
	doc, err := extract.Parse(t.Context(), input, extract.Options{})
	c.Assert(err, qt.IsNil)
	return dependencyReplay{blocks: doc.Blocks, source: source, identity: nlp.Identity{
		Name: traces.Provider, Version: traces.Revision, Model: "en_core_web_sm-3.8.0", ModelHash: traces.ModelHash,
		License: "MIT; WordNet-3.0", DependencyScheme: traces.Scheme,
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS, nlp.Dependencies},
	}}
}

func (p dependencyReplay) Identity() nlp.Identity { return p.identity }

func (p dependencyReplay) Analyze(ctx context.Context, mapped document.MappedText, _ []nlp.Capability) ([]document.Sentence, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, block := range p.blocks {
		if block.Text != mapped.Text || !slices.Equal(block.Map, mapped.Map) {
			continue
		}
		var result []document.Sentence
		for _, sentence := range p.source.Sentences {
			if sentence.BlockID == block.ID {
				result = append(result, sentence)
			}
		}
		return result, nil
	}
	return nil, fmt.Errorf("fixture has no prediction for the extracted block")
}

type dependencyRelationProbe struct{}

func (dependencyRelationProbe) Descriptor() rule.Descriptor {
	return rule.Descriptor{ID: "test.dependency-relation", Version: "1", Group: "test", Scope: "sentence",
		Requires: []nlp.Capability{nlp.Dependencies}, DependencyScheme: "spacy-en-3.8.0",
		Defaults: rule.Settings{Enabled: true, Severity: "warning", Gate: "none"}}
}

func (dependencyRelationProbe) Evaluate(ctx context.Context, view rule.View, emit rule.Emitter) error {
	for _, block := range view.Document.Blocks {
		for _, sentence := range block.Sentences {
			if err := emitDependencySubjects(sentence, emit); err != nil {
				return err
			}
		}
	}
	return ctx.Err()
}

func emitDependencySubjects(sentence document.Sentence, emit rule.Emitter) error {
	for i, arc := range sentence.Dependencies.Arcs {
		if arc.Relation != "nsubjpass" {
			continue
		}
		if err := emit.Emit(rule.Evidence{Kind: "statistical", Message: "Model assigned the nsubjpass relation.",
			Occurrences: []rule.Occurrence{{BlockID: sentence.BlockID, SentenceID: sentence.ID, Spans: sentence.Tokens[i].Spans}}}); err != nil {
			return err
		}
	}
	return nil
}
