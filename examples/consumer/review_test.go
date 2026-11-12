package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/review"
)

// A caller owns the model. The library prepares the prompt, reads the answer,
// and binds each flag to source bytes; it never contacts a model itself.
func TestReviewBindsACallersModelAnswerToTheSource(t *testing.T) {
	c := qt.New(t)
	text := "It is important to note that the client retries after a failure.\n"
	reviewer, err := review.New(review.Options{})
	c.Assert(err, qt.IsNil)
	prepared, err := reviewer.Prepare(t.Context(), document.Source{Name: "guide.md", Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(prepared.Items, qt.HasLen, 1)
	c.Assert(prepared.Prompt(), qt.Contains, text)

	// The answer of the caller's own model to prepared.Prompt().
	answer := `{"flags": [{"item": "i001", "categories": ["empty_framing"], ` +
		`"quote": "It is important to note that", "reason": "The opening only announces the statement."}]}`
	flags, err := review.ParseFlags(answer)
	c.Assert(err, qt.IsNil)
	findings, rejected, err := prepared.Bind(flags)
	c.Assert(err, qt.IsNil)
	c.Assert(rejected, qt.HasLen, 0)
	c.Assert(findings, qt.HasLen, 1)
	c.Assert(text[findings[0].Span.Start:findings[0].Span.End], qt.Equals, "It is important to note that")

	// A second model may judge the bound flags without seeing the first reasons.
	c.Assert(prepared.ConfirmationPrompt(findings), qt.Contains, "[f001] defect:")
	verdicts, err := review.ParseVerdicts(`{"reviews": [{"id": "f001", "justified": "yes", "reason": "Adds nothing."}]}`, len(findings))
	c.Assert(err, qt.IsNil)
	c.Assert(verdicts[0].Justified, qt.Equals, review.Justified)
}

// A trim is a proposal of the caller's model that the library accepts only as
// a pure deletion of words.
func TestTrimAcceptsOnlyDeletionsFromACallersModel(t *testing.T) {
	c := qt.New(t)
	text := "It is important to note that the client retries after a connection failure.\n"
	reviewer, err := review.New(review.Options{})
	c.Assert(err, qt.IsNil)
	prepared, err := reviewer.Prepare(t.Context(), document.Source{Name: "guide.md", Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(prepared.TrimPrompt(), qt.Contains, "i001: It is important to note")

	// The answer of the caller's own model to prepared.TrimPrompt().
	proposals, err := review.ParseTrims(`{"trims": [{"item": "i001", "text": "The client retries after a connection failure."}]}`)
	c.Assert(err, qt.IsNil)
	trims, rejected, err := prepared.BindTrims(proposals)
	c.Assert(err, qt.IsNil)
	c.Assert(rejected, qt.HasLen, 0)
	c.Assert(string(prepared.ApplyTrims(trims)), qt.Equals, "The client retries after a connection failure.\n")

	// A proposal that adds a word is not a trim.
	proposals, err = review.ParseTrims(`{"trims": [{"item": "i001", "text": "The client always retries after a connection failure."}]}`)
	c.Assert(err, qt.IsNil)
	trims, rejected, err = prepared.BindTrims(proposals)
	c.Assert(err, qt.IsNil)
	c.Assert(trims, qt.HasLen, 0)
	c.Assert(rejected[0].Problem, qt.Equals, "adds, replaces, or reorders words")
}
