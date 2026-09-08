package builtin_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

type phraseBenchmarkSink struct{}

func (phraseBenchmarkSink) Emit(rule.Evidence) error {
	return fmt.Errorf("unexpected phrase match")
}

func (phraseBenchmarkSink) Observe(feature.BlockObservation) error { return nil }

// BenchmarkPhraseApplicability isolates matching over 100,000 prepared tokens
// and 128 distinct 16-token phrases. It does not measure extraction or NLP.
func BenchmarkPhraseApplicability(b *testing.B) {
	for _, implementation := range builtin.Rules() {
		if implementation.Descriptor().ID != "policy.banned-phrases" {
			continue
		}
		for _, collect := range []bool{false, true} {
			b.Run(fmt.Sprintf("collection=%t", collect), func(b *testing.B) {
				benchmarkPhrase(b, implementation, collect)
			})
		}
	}
}

func benchmarkPhrase(b *testing.B, implementation rule.Rule, collect bool) {
	b.Helper()
	view := phraseBenchmarkView()
	if collect {
		view.Observer = phraseBenchmarkSink{}
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := implementation.Evaluate(b.Context(), view, phraseBenchmarkSink{}); err != nil {
			b.Fatal(err)
		}
	}
}

func phraseBenchmarkView() rule.View {
	view := rule.View{Document: &document.Document{}, Parameters: rule.Parameters{Positions: []string{"any"}}}
	for i := range 128 {
		view.Parameters.Phrases = append(view.Parameters.Phrases, strings.Repeat("missing ", 15)+fmt.Sprintf("term%d", i))
	}
	for i := range 2500 {
		sentence := document.Sentence{ID: i, BlockID: i, Words: 40}
		for range 40 {
			sentence.Tokens = append(sentence.Tokens, document.Token{Text: "present", Normal: "present", Word: true})
		}
		view.Document.Blocks = append(view.Document.Blocks,
			document.Block{ID: i, Kind: "paragraph", Words: 40, Sentences: []document.Sentence{sentence}})
	}
	return view
}
