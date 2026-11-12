package review_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/review"
)

const guide = "# Guide\n\n" +
	"It is important to note that the client retries after a failure. The retry\ninterval is 30 seconds.\n\n" +
	"Run `make build` before the first start of the service.\n\n" +
	"```sh\nmake build && make install-the-service-on-every-host\n```\n"

func prepare(c *qt.C, t *testing.T, name, text string) *review.Document {
	c.Helper()
	reviewer, err := review.New(review.Options{})
	c.Assert(err, qt.IsNil)
	prepared, err := reviewer.Prepare(t.Context(), document.Source{Name: name, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	return prepared
}

func itemContaining(c *qt.C, items []review.Item, text string) review.Item {
	c.Helper()
	for _, item := range items {
		if strings.Contains(item.Text, text) {
			return item
		}
	}
	c.Fatalf("no item contains %q in %+v", text, items)
	return review.Item{}
}

func TestPrepareListsProseItemsInSourceOrder(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "guide.md", guide)
	c.Assert(prepared.SourceHash, qt.HasLen, 64)
	first := itemContaining(c, prepared.Items, "It is important to note")
	c.Assert(first.Kind, qt.Equals, "sentence")
	c.Assert(first.Line, qt.Equals, 3)
	second := itemContaining(c, prepared.Items, "interval is 30 seconds")
	c.Assert(second.ID, qt.Not(qt.Equals), first.ID)
	for index, item := range prepared.Items {
		c.Assert(item.ID, qt.Not(qt.Equals), "")
		// Code is not prose, and a short heading is not worth a review.
		c.Assert(item.Text, qt.Not(qt.Contains), "install-the-service")
		c.Assert(item.Text, qt.Not(qt.Equals), "Guide")
		if index > 0 {
			c.Assert(item.ID > prepared.Items[index-1].ID, qt.IsTrue)
		}
	}
}

func TestBindReturnsOriginalBytes(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "guide.md", guide)
	first := itemContaining(c, prepared.Items, "It is important to note")
	second := itemContaining(c, prepared.Items, "interval is 30 seconds")
	accepted, rejected, err := prepared.Bind([]review.Flag{
		{Item: first.ID, Categories: []string{"empty_framing"},
			Quote: "It is important to note that", Reason: " The opening only announces the statement. "},
		// The quote was copied with a space where the source wraps the line.
		{Item: second.ID, Categories: []string{"wordiness"}, Quote: "The retry interval", Reason: "A wrapped quote."},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(rejected, qt.HasLen, 0)
	c.Assert(accepted, qt.HasLen, 2)
	finding := accepted[0]
	c.Assert(guide[finding.Span.Start:finding.Span.End], qt.Equals, "It is important to note that")
	c.Assert(finding.Quote, qt.Equals, "It is important to note that")
	c.Assert(finding.Reason, qt.Equals, "The opening only announces the statement.")
	c.Assert(finding.Path, qt.Equals, "guide.md")
	c.Assert(finding.Start.Line, qt.Equals, 3)
	c.Assert(finding.Start.Column, qt.Equals, 1)
	wrapped := accepted[1]
	c.Assert(guide[wrapped.Span.Start:wrapped.Span.End], qt.Equals, "The retry\ninterval")
	c.Assert(wrapped.End.Line, qt.Equals, 4)
}

func TestBindRejectsFlagsItCannotLocate(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "guide.md", guide)
	first := itemContaining(c, prepared.Items, "It is important to note")
	command := itemContaining(c, prepared.Items, "before the first start")
	reason := "A reason."
	accepted, rejected, err := prepared.Bind([]review.Flag{
		{Item: "i999", Categories: []string{"wordiness"}, Quote: "It is", Reason: reason},
		{Item: first.ID, Categories: []string{}, Quote: "It is", Reason: reason},
		{Item: first.ID, Categories: []string{"wordiness"}, Quote: "text that is absent", Reason: reason},
		{Item: first.ID, Categories: []string{"wordiness"}, Quote: "It is", Reason: " "},
		{Item: first.ID, Categories: []string{"authorship"}, Quote: "It is", Reason: reason},
		{Item: first.ID, Categories: []string{"wordiness"}, Quote: " \n", Reason: reason},
		// A quote must not cross protected inline code.
		{Item: command.ID, Categories: []string{"wordiness"}, Quote: "Run make build before", Reason: reason},
		{Item: first.ID, Categories: []string{"wordiness"}, Quote: "It is", Reason: reason},
		{Item: first.ID, Categories: []string{"wordiness"}, Quote: "the client", Reason: reason},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(accepted, qt.HasLen, 1)
	problems := make([]string, 0, len(rejected))
	for _, rejection := range rejected {
		problems = append(problems, rejection.Problem)
	}
	c.Assert(problems, qt.DeepEquals, []string{"unknown item", "no category",
		"quote does not occur in one segment of the item", "no reason", "unknown category authorship",
		"quote does not occur in one segment of the item", "quote does not occur in one segment of the item",
		"duplicate item"})
	c.Assert(rejected[0].Flag.Item, qt.Equals, "i999")
}

func TestBindKeepsMultibyteQuotesWhole(t *testing.T) {
	c := qt.New(t)
	text := "The naïve résumé parser is truly a world-class and remarkably café-ready tool for everyone.\n"
	prepared := prepare(c, t, "notes.txt", text)
	c.Assert(prepared.Items, qt.HasLen, 1)
	accepted, rejected, err := prepared.Bind([]review.Flag{{Item: prepared.Items[0].ID,
		Categories: []string{"unjustified_intensifiers"}, Quote: "remarkably café-ready", Reason: "Praise without information."}})
	c.Assert(err, qt.IsNil)
	c.Assert(rejected, qt.HasLen, 0)
	c.Assert(accepted, qt.HasLen, 1)
	c.Assert(text[accepted[0].Span.Start:accepted[0].Span.End], qt.Equals, "remarkably café-ready")
}

func TestBindRefusesMoreFlagsThanTheLimit(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "guide.md", guide)
	_, _, err := prepared.Bind(make([]review.Flag, review.MaxFlags+1))
	c.Assert(err, qt.ErrorMatches, "flags must contain at most 256 entries")
}

func TestPrepareReturnsNoItemsWithoutProse(t *testing.T) {
	for _, text := range []string{"", "```sh\nmake build\n```\n", "# Title\n"} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			prepared := prepare(c, t, "notes.md", text)
			c.Assert(prepared.Items, qt.HasLen, 0)
			c.Assert(prepared.SourceHash, qt.HasLen, 64)
			accepted, rejected, err := prepared.Bind([]review.Flag{{Item: "i001", Categories: []string{"wordiness"},
				Quote: "make", Reason: "A reason."}})
			c.Assert(err, qt.IsNil)
			c.Assert(accepted, qt.HasLen, 0)
			c.Assert(rejected, qt.HasLen, 1)
			c.Assert(rejected[0].Problem, qt.Equals, "unknown item")
		})
	}
}

func TestPrepareRejectsAnUnknownFormat(t *testing.T) {
	c := qt.New(t)
	reviewer, err := review.New(review.Options{})
	c.Assert(err, qt.IsNil)
	_, err = reviewer.Prepare(t.Context(), document.Source{Name: "a.bin", Format: "no-such-format", Bytes: []byte("Text.")})
	c.Assert(err, qt.IsNotNil)
}

func TestNewRejectsAnInvalidPolicy(t *testing.T) {
	c := qt.New(t)
	_, err := review.New(review.Options{Config: []byte("version: 999\n")})
	c.Assert(err, qt.IsNotNil)
}

func TestEveryCategoryHasAStatement(t *testing.T) {
	c := qt.New(t)
	c.Assert(review.Categories(), qt.HasLen, 7)
	for _, category := range review.Categories() {
		statement, ok := review.Statement(category)
		c.Assert(ok, qt.IsTrue, qt.Commentf("category %s", category))
		c.Assert(statement, qt.Not(qt.Equals), "")
		c.Assert(review.Instructions, qt.Contains, "- "+category+":")
	}
	_, ok := review.Statement("authorship")
	c.Assert(ok, qt.IsFalse)
}
