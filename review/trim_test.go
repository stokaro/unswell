package review_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/review"
)

const manual = "# Retry policy\n\n" +
	"It is important to note that the client retries a request after a connection failure. " +
	"The delay doubles after every attempt, up to 30 seconds.\n\n" +
	"The client does not retry a request that the server rejected.\n\n" +
	"Run `make build` before the first start of the service.\n\n" +
	"A really careful operator may raise the \"soft\" limit. A **bold** remark is never trimmed in place.\n"

func trim(c *qt.C, prepared *review.Document, contains, text string) review.Trim {
	c.Helper()
	return review.Trim{Item: itemContaining(c, prepared.Items, contains).ID, Text: text}
}

func TestTrimPromptListsOnlySentencesInOnePiece(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "manual.md", manual)
	prompt := prepared.TrimPrompt()
	c.Assert(prompt, qt.Contains, "===== DOCUMENT =====\n"+manual)
	_, listed, found := strings.Cut(prompt, "===== SENTENCES =====")
	c.Assert(found, qt.IsTrue)
	c.Assert(listed, qt.Contains, "It is important to note that the client retries")
	c.Assert(listed, qt.Contains, "The client does not retry a request that the server rejected.")
	// A sentence with inline code lies in two pieces of source and is not offered.
	c.Assert(listed, qt.Not(qt.Contains), "before the first start")
	for _, item := range prepared.Trimmable() {
		c.Assert(item.Kind, qt.Equals, "sentence")
	}
}

func TestBindTrimsAcceptsPureDeletions(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "manual.md", manual)
	accepted, rejected, err := prepared.BindTrims([]review.Trim{
		trim(c, prepared, "A really careful", "A careful operator may raise the \"soft\" limit."),
		trim(c, prepared, "It is important to note", "The client retries a request after a connection failure."),
	})
	c.Assert(err, qt.IsNil)
	c.Assert(rejected, qt.HasLen, 0)
	c.Assert(accepted, qt.HasLen, 2)
	// Accepted trims come back in source order.
	first := accepted[0]
	c.Assert(first.Original, qt.Equals, "It is important to note that the client retries a request after a connection failure.")
	c.Assert(first.Trimmed, qt.Equals, "The client retries a request after a connection failure.")
	c.Assert(first.Removed, qt.DeepEquals, []string{"it", "is", "important", "to", "note", "that"})
	c.Assert(manual[first.Span.Start:first.Span.End], qt.Equals, first.Original)
	c.Assert(first.Start.Line, qt.Equals, 3)
	c.Assert(first.Path, qt.Equals, "manual.md")
	c.Assert(accepted[1].Removed, qt.DeepEquals, []string{"really"})

	revised := string(prepared.ApplyTrims(accepted))
	c.Assert(revised, qt.Equals, "# Retry policy\n\n"+
		"The client retries a request after a connection failure. "+
		"The delay doubles after every attempt, up to 30 seconds.\n\n"+
		"The client does not retry a request that the server rejected.\n\n"+
		"Run `make build` before the first start of the service.\n\n"+
		"A careful operator may raise the \"soft\" limit. A **bold** remark is never trimmed in place.\n")
}

func TestBindTrimsRejectsEverythingButASafeDeletion(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "manual.md", manual)
	framed := "It is important to note"
	cases := []struct{ name, contains, text, problem string }{
		{"unchanged", framed, "It is important to note that the client retries a request after a connection failure.", "changes no word"},
		{"punctuation only", framed, "It is important to note that the client retries a request, after a connection failure!", "changes no word"},
		{"added word", framed, "Note that the client always retries a request after a connection failure.", "adds, replaces, or reorders words"},
		{"replaced word", framed, "The client repeats a request after a connection failure.", "adds, replaces, or reorders words"},
		{"reordered", framed, "After a connection failure the client retries a request.", "adds, replaces, or reorders words"},
		{"too short", framed, "The client retries.", "removes more than half of the words"},
		{"deleted", framed, " ", "removes more than half of the words"},
		{"negation", "does not retry", "The client does retry a request that the server rejected.", "removes a negation or a modal word"},
		{"modal", "careful operator", "A really careful operator raise the \"soft\" limit.", "removes a negation or a modal word"},
		{"number", "The delay doubles", "The delay doubles after every attempt, up to seconds.", "removes a number"},
		{"markup", "careful operator", "A careful operator may raise the soft limit.", "changes markup or quotation marks"},
		{"sentence with emphasis", "bold", "A remark is never trimmed in place.", "item cannot be trimmed in place"},
		{"unknown item", "", "The client retries.", "unknown item"},
		{"sentence with code", "before the first start", "Run before the first start.", "item cannot be trimmed in place"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			proposal := review.Trim{Item: "i999", Text: tc.text}
			if tc.contains != "" {
				proposal = trim(c, prepared, tc.contains, tc.text)
			}
			accepted, rejected, err := prepared.BindTrims([]review.Trim{proposal})
			c.Assert(err, qt.IsNil)
			c.Assert(accepted, qt.HasLen, 0)
			c.Assert(rejected, qt.HasLen, 1)
			c.Assert(rejected[0].Problem, qt.Equals, tc.problem)
		})
	}
}

func TestBindTrimsTakesOneProposalPerItem(t *testing.T) {
	c := qt.New(t)
	prepared := prepare(c, t, "manual.md", manual)
	proposal := trim(c, prepared, "It is important to note", "The client retries a request after a connection failure.")
	accepted, rejected, err := prepared.BindTrims([]review.Trim{proposal, proposal})
	c.Assert(err, qt.IsNil)
	c.Assert(accepted, qt.HasLen, 1)
	c.Assert(rejected, qt.HasLen, 1)
	c.Assert(rejected[0].Problem, qt.Equals, "duplicate item")
	_, _, err = prepared.BindTrims(make([]review.Trim, review.MaxTrims+1))
	c.Assert(err, qt.ErrorMatches, "trims must contain at most 1024 entries")
}

func TestParseTrimsTakesTheLastObjectOfACorrectedAnswer(t *testing.T) {
	c := qt.New(t)
	answer := `{"trims": [{"item": "i001", "text": "The client retries, given a failure."}]}` +
		"\n\nWait, that replaces a word. Corrected:\n\n" +
		`{"trims": [{"item": "i001", "text": "The client retries."}]}`
	trims, err := review.ParseTrims(answer)
	c.Assert(err, qt.IsNil)
	c.Assert(trims, qt.DeepEquals, []review.Trim{{Item: "i001", Text: "The client retries."}})

	// An object without the field does not replace the answer.
	trims, err = review.ParseTrims(`{"trims": [{"item": "i002", "text": "It works."}]}` + "\n" + `{"done": true}`)
	c.Assert(err, qt.IsNil)
	c.Assert(trims, qt.DeepEquals, []review.Trim{{Item: "i002", Text: "It works."}})

	// A correction that withdraws every proposal leaves none.
	trims, err = review.ParseTrims(`{"trims": [{"item": "i001", "text": "x"}]}` + " Corrected: " + `{"trims": []}`)
	c.Assert(err, qt.IsNil)
	c.Assert(trims, qt.HasLen, 0)
}

func TestParseTrimsRejectsACutOffCorrection(t *testing.T) {
	complete := `{"trims": [{"item": "i001", "text": "The client retries."}]}`
	cases := []struct{ name, answer, problem string }{
		{"cut-off correction", complete + "\nCorrected:\n" + `{"trims": [{"item": "i001", "text": "The cli`,
			"the answer is not the requested JSON object: .*"},
		{"brace that starts no object", complete + "\nI left {name} alone.", "the answer is not the requested JSON object: .*"},
		{"objects without the field", `{"note": "nothing"} {"done": true}`, "the answer has no trims field"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := review.ParseTrims(tc.answer)
			c.Assert(err, qt.ErrorMatches, tc.problem)
		})
	}
}

func TestParseTrimsRejectsAnswersWithoutTrims(t *testing.T) {
	c := qt.New(t)
	trims, err := review.ParseTrims("Done.\n```json\n" + `{"trims": [{"item": "i001", "text": "The client retries."}]}` + "\n```")
	c.Assert(err, qt.IsNil)
	c.Assert(trims, qt.DeepEquals, []review.Trim{{Item: "i001", Text: "The client retries."}})
	trims, err = review.ParseTrims(`{"trims": []}`)
	c.Assert(err, qt.IsNil)
	c.Assert(trims, qt.HasLen, 0)
	for _, answer := range []string{"I cannot do that.", `{"flags": []}`, `{"trims": null}`} {
		_, err = review.ParseTrims(answer)
		c.Assert(err, qt.IsNotNil, qt.Commentf("answer %q", answer))
	}
}
