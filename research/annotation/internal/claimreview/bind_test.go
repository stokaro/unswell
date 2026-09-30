package claimreview_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation/internal/claimreview"
)

func parsed(t *testing.T, text string, format document.Format) document.Document {
	c := qt.New(t)
	t.Helper()
	doc, err := extract.Parse(context.Background(), document.Source{Name: "example.md", Format: format,
		Bytes: []byte(text)}, extract.Options{})
	c.Assert(err, qt.IsNil)
	return doc
}

func reference(t *testing.T, doc document.Document, quote string, occurrence int) claimreview.Reference {
	c := qt.New(t)
	t.Helper()
	start, from := -1, 0
	for range occurrence {
		index := strings.Index(string(doc.Source[from:]), quote)
		c.Assert(index >= 0, qt.IsTrue)
		start = from + index
		from = start + len(quote)
	}
	span := document.Span{Start: start, End: start + len(quote)}
	blockID := -1
	for _, block := range doc.Blocks {
		if block.Span.Start <= span.Start && span.End <= block.Span.End {
			blockID = block.ID
			break
		}
	}
	return claimreview.Reference{Block: blockID, Span: span, Quote: quote}
}

func specification(target claimreview.Reference, key string) claimreview.Specification {
	return claimreview.Specification{Origin: claimreview.Origin{Run: "constructed", Candidate: "candidate-1", Key: key},
		Category: key, Diagnostic: "Review this wording.", Reason: "A separately declared wording defect.",
		Suggestion: "Revise the target while retaining its conditions.", Targets: []claimreview.Reference{target},
		Support: []claimreview.Reference{}}
}

func TestBindOwnsOriginalClaims(t *testing.T) {
	c := qt.New(t)
	doc := parsed(t, "The café reports an error.\r\n\r\nThe café reports an error.", document.Markdown)
	first := reference(t, doc, "The café reports an error.", 1)
	second := reference(t, doc, "The café reports an error.", 2)
	spec := specification(second, "repetition")
	spec.Support = []claimreview.Reference{first}
	inventory, err := claimreview.Bind(context.Background(), doc, []claimreview.Specification{spec})
	c.Assert(err, qt.IsNil)
	before := inventory.Claims()
	c.Assert(before[0].Targets[0].Span.Start > before[0].Support[0].Span.Start, qt.IsTrue)
	spec.Targets[0].Quote = "changed"
	doc.Source[0] = 'X'
	before[0].Support[0].Quote = "changed"
	c.Assert(inventory.Claims()[0].Targets[0].Quote, qt.Equals, second.Quote)
	c.Assert(inventory.Claims()[0].Support[0].Quote, qt.Equals, first.Quote)
	c.Assert(inventory.Claims()[0].ID, qt.Equals, before[0].ID)
}

func TestBindRejectsInventedSourceAndRoles(t *testing.T) {
	c := qt.New(t)
	doc := parsed(t, "A café reports an error.\n\nRun `privateCall()` only locally.\n\n```go\nprivateCall()\n```\n", document.Markdown)
	valid := specification(reference(t, doc, "A café reports an error.", 1), "wording")
	tests := []struct {
		name string
		edit func(*claimreview.Specification)
	}{
		{"invented quote", func(s *claimreview.Specification) { s.Targets[0].Quote = "invented" }},
		{"invented owner", func(s *claimreview.Specification) { s.Targets[0].Block = 999 }},
		{"split Unicode", func(s *claimreview.Specification) {
			s.Targets[0].Span.End = strings.Index(string(doc.Source), "é") + 1
		}},
		{"context is also target", func(s *claimreview.Specification) { s.Support = append(s.Support, s.Targets[0]) }},
		{"protected inline code", func(s *claimreview.Specification) { s.Targets[0] = reference(t, doc, "privateCall()", 1) }},
		{"protected fence", func(s *claimreview.Specification) { s.Targets[0] = reference(t, doc, "privateCall()", 2) }},
		{"missing target", func(s *claimreview.Specification) { s.Targets = nil }},
		{"empty rationale", func(s *claimreview.Specification) { s.Reason = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			spec := valid
			spec.Targets = append([]claimreview.Reference(nil), valid.Targets...)
			test.edit(&spec)
			_, err := claimreview.Bind(context.Background(), doc, []claimreview.Specification{spec})
			c.Assert(err, qt.IsNotNil)
		})
	}
	doc.Hash = strings.Repeat("0", 64)
	_, err := claimreview.Bind(context.Background(), doc, []claimreview.Specification{valid})
	c.Assert(err, qt.ErrorMatches, "invalid source identity or size")
}

func TestClaimIDsDistinguishIndependentCriticisms(t *testing.T) {
	c := qt.New(t)
	doc := parsed(t, "Four settings decides where requests go.", document.Plain)
	target := reference(t, doc, string(doc.Source), 1)
	specs := []claimreview.Specification{specification(target, "grammar"), specification(target, "rhetoric")}
	inventory, err := claimreview.Bind(context.Background(), doc, specs)
	c.Assert(err, qt.IsNil)
	claims := inventory.Claims()
	c.Assert(claims[0].ID, qt.Not(qt.Equals), claims[1].ID)
	other, err := claimreview.Bind(context.Background(), doc, specs)
	c.Assert(err, qt.IsNil)
	c.Assert(other.Claims(), qt.DeepEquals, claims)
	specs[1].Origin = specs[0].Origin
	_, err = claimreview.Bind(context.Background(), doc, specs)
	c.Assert(err, qt.ErrorMatches, "duplicate original claim provenance")
}

func TestBindCancellation(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := claimreview.Bind(ctx, document.Document{}, nil)
	c.Assert(err, qt.Equals, context.Canceled)
}
