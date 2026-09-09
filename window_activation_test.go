package unswell_test

import (
	"slices"
	"strconv"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
)

func windowActivationIDs() []string {
	return []string{"syntax.not-only-density", "syntax.paired-contrast-density", "syntax.triad-density",
		"syntax.whether-preface-density", "syntax.rhetorical-question-density", "syntax.passive-candidate-density"}
}

func TestWindowActivationsPreserveCatalogExamples(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(windowActivationIDs(), d.ID) {
			continue
		}
		t.Run(d.ID, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(d.BlockObservations, qt.IsTrue)
			for _, example := range d.Examples {
				compareLocalExample(t, d.ID, example)
			}
		})
	}
}

type windowAbsenceCase struct{ text, parameters, extra, reason string }

func checkWindowAbsence(t *testing.T, id string, cases []windowAbsenceCase) {
	t.Helper()
	for _, test := range cases {
		t.Run(test.text+test.parameters, func(t *testing.T) {
			c := qt.New(t)
			value := localActivation(t, id, test.text, test.parameters, test.extra)
			if test.reason != "" {
				c.Assert(value.Number, qt.IsNil)
				c.Assert(value.Reason, qt.Equals, "inapplicable/"+test.reason)
			} else {
				c.Assert(value.Reason, qt.Equals, "")
				c.Assert(value.Number, qt.IsNotNil)
				c.Assert(*value.Number, qt.Equals, float64(0))
			}
		})
	}
}

func windowTerm(id, term string) string {
	return "vocabulary:\n  terms: [" + strconv.Quote(term) + "]\n  term_exemptions: [" + id + "]\n"
}

func TestWindowNotOnlyApplicability(t *testing.T) {
	checkWindowAbsence(t, "syntax.not-only-density", []windowAbsenceCase{
		{"The client starts.", "", "", ""},
		{"It not only reads but writes.", "", "", ""},
		{"not only", "", "", "no_eligible_window"},
		{"not only; but", "", "", "no_eligible_window"},
		{"It not only reads `code` but writes.", "", "", "no_eligible_window"},
		{"# The client starts\n", "", "", "unsupported_unit"},
	})
}

func TestWindowContrastApplicability(t *testing.T) {
	id := "syntax.paired-contrast-density"
	checkWindowAbsence(t, id, []windowAbsenceCase{
		{"The client starts today. The server waits.", "", "", ""},
		{"It is not about speed. It is about impact.", "", "", ""},
		{"It is not about speed. It is about impact.", "{window_sentences: 1, allowed_occurrences: 0}", "", "no_eligible_window"},
		{"It is not about speed.", "", "", "no_eligible_window"},
		{"It is not about speed. Yes.", "", "", "no_eligible_window"},
		{"It is not about speed? It is about impact.", "", "", "no_eligible_window"},
		{"It is not about speed. It is about impact?", "", "", "no_eligible_window"},
		{"It is not about speed. It is about impact.", "", windowTerm(id, "it is not about"), "no_eligible_window"},
		{"It is not about speed. It is about `impact`.", "", "", "no_eligible_window"},
	})
}

func TestWindowWhetherApplicability(t *testing.T) {
	id := "syntax.whether-preface-density"
	checkWindowAbsence(t, id, []windowAbsenceCase{
		{"The client can open connections, then read the reply.", "", "", ""},
		{"Whether you are using HTTP or HTTPS, validate the certificate policy.", "", "", ""},
		{"The client opens connections.", "", "", "no_eligible_window"},
		{"Whether you are, check.", "", "", "no_eligible_window"},
		{"Whether you are using HTTP, validate the policy.", "", windowTerm(id, "whether you are"), ""},
		{"Whether you are using HTTP, validate the policy.", "", windowTerm(id, "whether you are using HTTP,"), "no_eligible_window"},
		{"Whether you are using `HTTP` or HTTPS, validate the policy.", "", "", "no_eligible_window"},
	})
}

func TestWindowQuestionApplicability(t *testing.T) {
	id := "syntax.rhetorical-question-density"
	checkWindowAbsence(t, id, []windowAbsenceCase{
		{"The server? It waits.", "", "", ""},
		{"The result? A better experience.", "", "", ""},
		{"The result? A better experience.", "{window_sentences: 1, allowed_occurrences: 0}", "", "no_eligible_window"},
		{"The result?", "", "", "no_eligible_window"},
		{"The result? 42 requests.", "", "", "no_eligible_window"},
		{"The result? More questions?", "", "", "no_eligible_window"},
		{"The result? A better experience.", "{max_answer_words: 1}", "", "no_eligible_window"},
		{"The result of this? An answer.", "{phrases: ['the result?']}", "", "no_eligible_window"},
		{"The result? An answer.", "{phrases: []}", "", "no_patterns"},
		{"The result? An answer.", "", windowTerm(id, "the result?"), "no_eligible_window"},
		{"The result? An answer.", "", windowTerm(id, "an answer."), "no_eligible_window"},
		{"The result? `An answer`.", "", "", "no_eligible_window"},
	})
}

func TestWindowTriadApplicability(t *testing.T) {
	id := "syntax.triad-density"
	checkWindowAbsence(t, id, []windowAbsenceCase{
		{"The client starts.", "", "", ""},
		{"A powerful, seamless, innovative platform starts.", "", "", ""},
		{"powerful", "", "", ""},
		{"powerful", "{phrases: []}", "", "no_patterns"},
		{"powerful", "", windowTerm(id, "powerful"), "no_eligible_window"},
		{"`powerful seamless innovative`.", "", "", "no_eligible_window"},
		{".", "", "", "no_eligible_window"},
	})
}

func TestWindowPassiveApplicability(t *testing.T) {
	id := "syntax.passive-candidate-density"
	checkWindowAbsence(t, id, []windowAbsenceCase{
		{"The client sends a request to the server and waits for a reply.", "", "", ""},
		{"The request is carefully validated by the server before execution.", "", "", ""},
		{"The client waits. The server starts. The client reads. The server stops.", "", "", "insufficient_words"},
		{"The request is carefully validated by the server before execution. The client waits.", "", "", ""},
		{"The request is carefully validated by the server before `execution`.", "", "", "no_eligible_window"},
		{"The client opens connections", "{min_words: 0}", windowTerm(id, "the client opens connections"), "no_eligible_window"},
		{"is validated", "{min_words: 0}", windowTerm(id, "is"), ""},
		{".", "{min_words: 0}", "", "no_eligible_window"},
	})
}
