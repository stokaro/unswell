package nlp_test

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

func dependencySentence(text string, heads []int) (document.MappedText, document.Sentence) {
	mapped := document.MappedText{Text: text, Map: make([]document.Span, len(text))}
	for i := range text {
		_, size := utf8.DecodeRuneInString(text[i:])
		end := i + size
		for j := i; j < end; j++ {
			mapped.Map[j] = document.Span{Start: i + 10, End: end + 10}
		}
	}
	sentence := document.Sentence{Text: text, Spans: mapped.Spans(0, len(text)), Dependencies: &document.DependencyTree{}}
	sentence.Span = document.Bounds(sentence.Spans)
	start := 0
	for i, word := range strings.Fields(text) {
		start += strings.Index(text[start:], word)
		end := start + len(word)
		sentence.Tokens = append(sentence.Tokens, document.Token{
			Text: word, Start: start, End: end, Spans: mapped.Spans(start, end), Word: true,
		})
		sentence.Dependencies.Arcs = append(sentence.Dependencies.Arcs, document.DependencyArc{Head: heads[i], Relation: "dep"})
		start = end
	}
	return mapped, sentence
}

func TestDependencyTrees(t *testing.T) {
	for _, row := range []struct {
		name  string
		heads []int
		err   string
	}{
		{"chain", []int{1, 2, 3, -1}, ""},
		{"nonprojective", []int{2, 3, -1, 2}, ""},
		{"two roots", []int{-1, 0, -1, 2}, ".*2 roots.*"},
		{"no root", []int{1, 2, 3, 0}, ".*0 roots.*"},
		{"disconnected cycle", []int{-1, 2, 3, 1}, ".*cycle.*"},
		{"self head", []int{0, 0, 1, 2}, ".*invalid head.*"},
		{"negative head", []int{-2, 0, 1, 2}, ".*invalid head.*"},
		{"outside sentence", []int{4, 0, 1, -1}, ".*invalid head.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			mapped, sentence := dependencySentence("Café clients may retry", row.heads)
			err := nlp.ValidateDependencyTree(t.Context(), mapped, sentence)
			if row.err == "" {
				c.Assert(err, qt.IsNil)
			} else {
				c.Assert(err, qt.ErrorMatches, row.err)
			}
		})
	}
}

func TestDependencyRepresentationFailures(t *testing.T) {
	for _, row := range []struct {
		name   string
		change func(*document.MappedText, *document.Sentence)
		err    string
	}{
		{"missing", func(_ *document.MappedText, s *document.Sentence) { s.Dependencies = nil }, ".*missing"},
		{"incomplete", func(_ *document.MappedText, s *document.Sentence) { s.Dependencies.Arcs = nil }, ".*one arc.*"},
		{"no tokens", func(_ *document.MappedText, s *document.Sentence) { s.Tokens = nil }, ".*one arc.*"},
		{"map size", func(m *document.MappedText, _ *document.Sentence) { m.Map = nil }, ".*source map"},
		{"negative start", func(_ *document.MappedText, s *document.Sentence) { s.Tokens[0].Start = -1 }, ".*byte range"},
		{"overlap", func(_ *document.MappedText, s *document.Sentence) { s.Tokens[1].Start = 0 }, ".*byte range"},
		{"outside input", func(_ *document.MappedText, s *document.Sentence) { s.Tokens[0].End = 999 }, ".*byte range"},
		{"protected", func(_ *document.MappedText, s *document.Sentence) { s.Tokens[0].Protected = true }, ".*protected token.*"},
		{"token text", func(_ *document.MappedText, s *document.Sentence) { s.Tokens[0].Text = "other" }, ".*text differs.*"},
		{"token map", func(_ *document.MappedText, s *document.Sentence) { s.Tokens[0].Spans = nil }, ".*source spans"},
		{"sentence text", func(_ *document.MappedText, s *document.Sentence) { s.Text = "other" }, ".*differs.*"},
		{"sentence map", func(_ *document.MappedText, s *document.Sentence) { s.Spans = nil }, ".*source spans"},
		{"sentence bounds", func(_ *document.MappedText, s *document.Sentence) { s.Span.End++ }, ".*source spans"},
		{"empty relation", func(_ *document.MappedText, s *document.Sentence) { s.Dependencies.Arcs[0].Relation = "" }, ".*relation label"},
		{"space in relation", func(_ *document.MappedText, s *document.Sentence) {
			s.Dependencies.Arcs[0].Relation = "nsubj pass"
		}, ".*relation label"},
		{"long relation", func(_ *document.MappedText, s *document.Sentence) {
			s.Dependencies.Arcs[0].Relation = strings.Repeat("a", 129)
		}, ".*relation label"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			mapped, sentence := dependencySentence("Clients may retry", []int{2, 2, -1})
			row.change(&mapped, &sentence)
			c.Assert(nlp.ValidateDependencyTree(t.Context(), mapped, sentence), qt.ErrorMatches, row.err)
		})
	}
}

func TestDependencyProtectedBoundaryAndCancellation(t *testing.T) {
	c := qt.New(t)
	mapped, sentence := dependencySentence("Clients \x00 retry", []int{2, 2, -1})
	c.Assert(nlp.ValidateDependencyTree(t.Context(), mapped, sentence), qt.ErrorMatches, ".*protected boundary.*")
	// Omitting the boundary token cannot create a dependency across excluded code.
	sentence.Tokens = append(sentence.Tokens[:1], sentence.Tokens[2])
	sentence.Dependencies.Arcs = []document.DependencyArc{{Head: 1, Relation: "nsubj"}, {Head: -1, Relation: "ROOT"}}
	c.Assert(nlp.ValidateDependencyTree(t.Context(), mapped, sentence), qt.ErrorMatches, ".*protected boundary.*")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(nlp.ValidateDependencyTree(ctx, mapped, sentence), qt.ErrorIs, context.Canceled)
}

func TestDependencyCoverage(t *testing.T) {
	c := qt.New(t)
	mapped, sentence := dependencySentence("Clients may retry", []int{2, 2, -1})
	c.Assert(nlp.ValidateDependencies(t.Context(), mapped, nil), qt.ErrorMatches, ".*omit extracted prose")
	c.Assert(nlp.ValidateDependencies(t.Context(), mapped, []document.Sentence{sentence, sentence}), qt.ErrorMatches, ".*overlap.*")
	c.Assert(nlp.ValidateDependencies(t.Context(), mapped, []document.Sentence{sentence}), qt.IsNil)
	sentence.Tokens = append(sentence.Tokens[:1], sentence.Tokens[2])
	sentence.Dependencies.Arcs = []document.DependencyArc{{Head: 1, Relation: "nsubj"}, {Head: -1, Relation: "ROOT"}}
	c.Assert(nlp.ValidateDependencyTree(t.Context(), mapped, sentence), qt.ErrorMatches, ".*tokens omit extracted prose")
}
