package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

// habitPage is a Markdown page of the given paragraphs under one heading.
func habitPage(paragraphs ...string) string {
	return "# Runner\n\n" + strings.Join(paragraphs, "\n\n") + "\n"
}

// habitPlaces analyzes a page and returns the text at each place of the finding of one rule, in the order of the page.
func habitPlaces(c *qt.C, t *testing.T, config, ruleID, text string) (places []string, found bool) {
	t.Helper()
	engine, err := unswell.New(unswell.Options{Config: []byte(config)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "page.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	for _, finding := range result.Findings {
		if finding.RuleID != ruleID {
			continue
		}
		c.Assert(found, qt.IsFalse, qt.Commentf("the rule reports a page once"))
		found = true
		c.Assert(finding.Severity, qt.Equals, "warning")
		for _, location := range append([]unswell.Location{finding.Primary}, finding.Related...) {
			places = append(places, text[location.Span.Start:location.Span.End])
		}
	}
	return places, found
}

// Paragraphs that end on a denial, each in one of the forms that the rule reads.
const (
	deniesPlainly    = "The runner reads the old history. It does not rewrite a file."
	deniesAfterASemi = "The check compares two schemas. The comparison runs on a copy; it does not call the service."
	deniesWithATail  = "We ran the plan on a copy. This is a rehearsal, not a migration."
	deniesContracted = "The tool prints the plan. It doesn't apply it."
	deniesTwice      = "The report lists every statement. It neither repairs SQL nor approves the next plan."
	deniesAfterADash = "The files stay as they are. The history is read from the directory — the runner never edits one."
	endsOnAStatement = "The runner reads the old history. It keeps every file as it is."
)

func TestClosingDenialDensityCountsParagraphsThatEndOnADenial(t *testing.T) {
	cases := []struct {
		name   string
		text   string
		places []string
	}{
		{
			"three paragraphs, one form each", habitPage(deniesPlainly, deniesAfterASemi, deniesWithATail),
			[]string{"It does not rewrite a file.", "The comparison runs on a copy; it does not call the service.",
				"This is a rehearsal, not a migration."},
		},
		{
			"a contraction, neither and nor, and a dash", habitPage(deniesContracted, deniesTwice, deniesAfterADash),
			[]string{"It doesn't apply it.", "It neither repairs SQL nor approves the next plan.",
				"The history is read from the directory — the runner never edits one."},
		},
		{"two are allowed", habitPage(deniesPlainly, deniesWithATail, endsOnAStatement), nil},
		{
			"a paragraph of one sentence is not counted",
			habitPage(deniesPlainly, deniesWithATail, "It does not rewrite a file."), nil,
		},
		{
			"a denial in the middle of a paragraph is not counted",
			habitPage(deniesPlainly, deniesWithATail, "The runner does not rewrite a file. It reads the old history."), nil,
		},
		{
			"a condition, a question, and an instruction are not denials",
			habitPage(deniesPlainly, deniesWithATail,
				"The runner stops. If the file is not there, it reports the path.",
				"The runner stops. Is the file not there?",
				"The runner stops. Do not edit the file by hand."), nil,
		},
		{
			"a warning to the reader and a limit are not denials",
			habitPage(deniesPlainly, deniesWithATail,
				"The runner stops. You cannot resume it from the same step.",
				"The runner stops. It is not only slower, it also locks the table.",
				"The runner stops. The fix has not been released yet."), nil,
		},
		{
			"a negation far from the start of the sentence is not counted",
			habitPage(deniesPlainly, deniesWithATail,
				"The runner stops. The runner reads every file of the old history in order and does not stop early."), nil,
		},
		{
			"list items, headings, and code are not paragraphs",
			"# Runner does not rewrite\n\n" + deniesPlainly + "\n\n" + deniesWithATail + "\n\n" +
				"- The runner stops. It does not rewrite a file.\n- The check runs. It does not call the service.\n\n" +
				"```text\nThe runner stops. It does not rewrite a file.\n```\n", nil,
		},
		{
			"a word in code is not a negation",
			habitPage(deniesPlainly, deniesWithATail, "The runner stops. The flag `not` selects the other files."), nil,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			places, found := habitPlaces(c, t, "", "syntax.closing-denial-density", test.text)
			c.Assert(places, qt.DeepEquals, test.places)
			c.Assert(found, qt.Equals, len(test.places) > 0)
		})
	}
}

func TestClosingDenialDensityNeedsAShareOfTheParagraphs(t *testing.T) {
	c := qt.New(t)
	// Three paragraphs of forty end on a denial, which is under eight in a hundred.
	paragraphs := []string{deniesPlainly, deniesAfterASemi, deniesWithATail}
	for len(paragraphs) < 40 {
		paragraphs = append(paragraphs, endsOnAStatement)
	}
	_, found := habitPlaces(c, t, "", "syntax.closing-denial-density", habitPage(paragraphs...))
	c.Assert(found, qt.IsFalse)
	// With 37 paragraphs they are over it.
	places, found := habitPlaces(c, t, "", "syntax.closing-denial-density", habitPage(paragraphs[:37]...))
	c.Assert(found, qt.IsTrue)
	c.Assert(places, qt.HasLen, 3)
	// The limits are parameters of the rule.
	strict := "version: 1\nrules:\n  syntax.closing-denial-density:\n    parameters: {allowed_occurrences: 1, onset: 5}\n"
	places, found = habitPlaces(c, t, strict, "syntax.closing-denial-density", habitPage(paragraphs...))
	c.Assert(found, qt.IsTrue)
	c.Assert(places, qt.HasLen, 3)
	places, found = habitPlaces(c, t, strict, "syntax.closing-denial-density", habitPage(deniesPlainly, deniesWithATail, endsOnAStatement))
	c.Assert(found, qt.IsTrue)
	c.Assert(places, qt.HasLen, 2)
	off := "version: 1\nrules:\n  syntax.closing-denial-density: {enabled: false}\n"
	_, found = habitPlaces(c, t, off, "syntax.closing-denial-density", habitPage(deniesPlainly, deniesAfterASemi, deniesWithATail))
	c.Assert(found, qt.IsFalse)
}

// formalSentences is running prose of the given number of sentences with no casual mark.
func formalSentences(count int) string {
	return strings.TrimSpace(strings.Repeat("The server reads the request and sends the reply to the client. ", count))
}

func TestUniformRegisterReportsAPageWithoutACasualMark(t *testing.T) {
	const first = "The server reads the request and sends the reply to the client."
	cases := []struct {
		name     string
		text     string
		reported bool
	}{
		{"twenty-five formal sentences", habitPage(formalSentences(25)), true},
		{"twenty-four are too few", habitPage(formalSentences(24)), false},
		{"sentences of several paragraphs count together", habitPage(formalSentences(10), formalSentences(10), formalSentences(5)), true},
		{"a contraction", habitPage(formalSentences(25), "It doesn't keep the request."), false},
		{"a contraction with a typographic apostrophe", habitPage(formalSentences(25), "It doesn’t keep the request."), false},
		{"a contracted pronoun", habitPage(formalSentences(25), "It's the same request."), false},
		{"a contracted verb", habitPage(formalSentences(25), "We've seen the same request twice."), false},
		{"an aside in parentheses", habitPage(formalSentences(25), "The server keeps the request (for a minute or two)."), false},
		{"a question", habitPage(formalSentences(25), "Does the server keep the request?"), false},
		{"an exclamation", habitPage(formalSentences(25), "The server keeps the request!"), false},
		{"a sentence that opens with But", habitPage(formalSentences(25), "But the server keeps the request."), false},
		{"a possessive is not a contraction", habitPage(formalSentences(25), "The client's request stays in the server's queue."), true},
		{"one word in parentheses is not an aside", habitPage(formalSentences(25), "The server answers over a socket (TCP)."), true},
		{"a contraction in code is not prose", habitPage(formalSentences(25), "The flag `don't-keep` drops the request."), true},
		{
			"list items and headings are not running prose",
			"# Isn't it fast?\n\n" + formalSentences(25) + "\n\n- It doesn't keep the request.\n- But it keeps the reply (for a while).\n", true,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			places, found := habitPlaces(c, t, "", "syntax.uniform-register", test.text)
			c.Assert(found, qt.Equals, test.reported)
			if test.reported {
				// The page is reported once, at its first sentence.
				c.Assert(places, qt.DeepEquals, []string{first})
			}
		})
	}
}

func TestUniformRegisterTakesItsLimitsFromTheConfiguration(t *testing.T) {
	c := qt.New(t)
	page := habitPage(formalSentences(25), "It doesn't keep the request.")
	_, found := habitPlaces(c, t, "", "syntax.uniform-register", page)
	c.Assert(found, qt.IsFalse)
	// One casual sentence is allowed, and a second one is not.
	loose := "version: 1\nrules:\n  syntax.uniform-register:\n    parameters: {min_sentences: 25, allowed_occurrences: 1}\n"
	_, found = habitPlaces(c, t, loose, "syntax.uniform-register", page)
	c.Assert(found, qt.IsTrue)
	_, found = habitPlaces(c, t, loose, "syntax.uniform-register", habitPage(formalSentences(25), "It doesn't keep the request. Why would it?"))
	c.Assert(found, qt.IsFalse)
	short := "version: 1\nrules:\n  syntax.uniform-register:\n    parameters: {min_sentences: 10}\n"
	_, found = habitPlaces(c, t, short, "syntax.uniform-register", habitPage(formalSentences(10)))
	c.Assert(found, qt.IsTrue)
	off := "version: 1\nrules:\n  syntax.uniform-register: {enabled: false}\n"
	_, found = habitPlaces(c, t, off, "syntax.uniform-register", habitPage(formalSentences(25)))
	c.Assert(found, qt.IsFalse)
}
