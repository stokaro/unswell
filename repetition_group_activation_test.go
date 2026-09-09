package unswell_test

import (
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
)

func repetitionGroupActivationIDs() []string {
	return []string{"repetition.exact-sentence", "repetition.sentence-openers", "repetition.paragraph-openers"}
}

func TestRepetitionGroupActivationsPreserveCatalogExamples(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(repetitionGroupActivationIDs(), d.ID) {
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

func TestRepetitionGroupActivationsDistinguishMissingFromZero(t *testing.T) {
	const prose = "The client opens a connection to the server and sends the request with its credentials."
	for _, test := range []struct{ id, text, parameters, reason string }{
		{"repetition.exact-sentence", prose, "", ""},
		{"repetition.exact-sentence", "The client starts.", "", "insufficient_words"},
		{"repetition.exact-sentence", "The client waits. The server starts. The client reads. The server stops.", "", "insufficient_words"},
		{"repetition.exact-sentence", "The client opens a connection to the server and sends the request with `credentials`.",
			"", "no_eligible_tokens"},
		{"repetition.exact-sentence", "`code`.", "{min_words: 0}", "no_eligible_tokens"},
		{"repetition.exact-sentence", ".", "{min_words: 0}", ""},
		{"repetition.exact-sentence", "# " + prose, "", ""},
		{"repetition.exact-sentence", "- " + prose, "", ""},
		{"repetition.sentence-openers", prose, "", ""},
		{"repetition.sentence-openers", "The client starts.", "", "insufficient_words"},
		{"repetition.sentence-openers", "The client waits. The server starts. The client reads. The server stops.", "", "insufficient_words"},
		{"repetition.sentence-openers", "The client waits.", "{min_words: 0, opener_words: 4}", "no_eligible_tokens"},
		{"repetition.sentence-openers", "`code`.", "{min_words: 0}", "no_eligible_tokens"},
		{"repetition.sentence-openers", "The client `code` opens connections.", "{min_words: 0}", ""},
		{"repetition.sentence-openers", "Hi. " + prose, "", ""},
		{"repetition.sentence-openers", "# " + prose, "", "unsupported_unit"},
		{"repetition.sentence-openers", "- " + prose, "", "unsupported_unit"},
		{"repetition.paragraph-openers", prose, "", ""},
		{"repetition.paragraph-openers", "Hi. " + prose, "", "insufficient_words"},
		{"repetition.paragraph-openers", "`code`. " + prose, "{min_words: 0}", "no_eligible_tokens"},
		{"repetition.paragraph-openers", "The client waits.", "{min_words: 0, opener_words: 4}", "no_eligible_tokens"},
		{"repetition.paragraph-openers", "The client `code` opens connections.", "{min_words: 0}", ""},
		{"repetition.paragraph-openers", "# " + prose, "", "unsupported_unit"},
	} {
		t.Run(test.id+"/"+test.text, func(t *testing.T) {
			c := qt.New(t)
			value := localActivation(t, test.id, test.text, test.parameters, "")
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
