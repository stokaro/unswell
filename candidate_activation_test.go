package unswell_test

import (
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
)

func candidateActivationIDs() []string {
	return []string{"format.list-fragmentation", "repetition.heading-echo", "repetition.near-sentence",
		"repetition.ngram-density", "repetition.paragraph-overlap", "repetition.summary-echo", "repetition.syntax-template"}
}

func TestCandidateActivationsPreserveCatalogExamples(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(candidateActivationIDs(), d.ID) {
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

func TestCandidateActivationsDistinguishMissingFromZero(t *testing.T) {
	for _, row := range []struct{ id, text, parameters, extra, reason string }{
		{"repetition.near-sentence", overlapParagraph, "", "", "no_eligible_pair"},
		{"repetition.near-sentence", overlapParagraph + " " + overlapParagraph, "", "", ""},
		{"repetition.near-sentence", overlapParagraph + " " + overlapParagraph + " Short.", "", "", ""},
		{"repetition.near-sentence", nearContrastSentence + " " + strings.Replace(nearContrastSentence, "may", "must", 1), "", "", ""},
		{"repetition.near-sentence", "The client starts.", "", "", "insufficient_words"},
		{"repetition.near-sentence", "The client waits. The server starts. The client reads. The server stops.",
			"", "", "insufficient_words"},
		{"repetition.near-sentence", strings.Replace(overlapParagraph, "credentials", "`credentials`", 1),
			"", "", "no_eligible_tokens"},
		{"repetition.ngram-density", overlapParagraph, "", "", ""},
		{"repetition.ngram-density", "Use `code`. " + overlapParagraph, "", "", ""},
		{"repetition.ngram-density", "The client starts.", "", "", "insufficient_words"},
		{"repetition.ngram-density", strings.Repeat("client ", 12) + ".", "", "", "no_eligible_tokens"},
		{"repetition.ngram-density", "Alpha beta gamma.", "{min_words: 0}",
			"vocabulary:\n  terms: [alpha beta gamma]\n  term_exemptions: [repetition.ngram-density]\n", "no_eligible_tokens"},
		{"repetition.syntax-template", "The careful writer describes the simple process for the entire local team.", "", "", ""},
		{"repetition.syntax-template", "Open the connection to the server and wait for the complete response.",
			"", "", "no_eligible_tokens"},
		{"repetition.syntax-template", strings.Replace(overlapParagraph, "credentials", "`credentials`", 1),
			"", "", "no_eligible_tokens"},
		{"repetition.paragraph-overlap", overlapParagraph, "", "", "no_eligible_pair"},
		{"repetition.paragraph-overlap", "The client starts.", "", "", "insufficient_words"},
		{"repetition.paragraph-overlap", strings.Repeat("client ", 12) + ".", "", "", "no_eligible_tokens"},
		{"repetition.paragraph-overlap", strings.Replace(overlapParagraph, "credentials", "`credentials`", 1),
			"", "", "no_eligible_tokens"},
		{"repetition.paragraph-overlap", overlapParagraph, "{min_words: 14}",
			"vocabulary:\n  terms: [opens a connection]\n  term_exemptions: [repetition.paragraph-overlap]\n", "insufficient_words"},
		{"repetition.heading-echo", overlapParagraph, "", "", "no_eligible_pair"},
		{"repetition.heading-echo", "# A practical approach to the delivery process", "", "", "no_eligible_pair"},
		{"repetition.summary-echo", overlapParagraph, "", "", "no_eligible_pair"},
		{"repetition.summary-echo", "# Summary", "", "", "unsupported_unit"},
		{"format.list-fragmentation", overlapParagraph, "", "", "unsupported_unit"},
		{"format.list-fragmentation", "- Clear reports", "", "", ""},
		{"format.list-fragmentation", "1. Clear reports", "", "", "no_eligible_list"},
		{"format.list-fragmentation", "- [ ] Clear reports", "", "", "no_eligible_list"},
		{"format.list-fragmentation", "- Clear reports.", "", "", "no_eligible_list"},
		{"format.list-fragmentation", "- Open connections", "", "", "no_eligible_list"},
		{"format.list-fragmentation", "- Clear `reports`", "", "", "no_eligible_list"},
		{"format.list-fragmentation", "- Clear reports", "",
			"vocabulary:\n  terms: [clear reports]\n  term_exemptions: [format.list-fragmentation]\n", "no_eligible_list"},
	} {
		t.Run(row.id+"/"+row.text, func(t *testing.T) {
			c := qt.New(t)
			value := localActivation(t, row.id, row.text, row.parameters, row.extra)
			if row.reason != "" {
				c.Assert(value.Number, qt.IsNil)
				c.Assert(value.Reason, qt.Equals, "inapplicable/"+row.reason)
			} else {
				c.Assert(value.Number, qt.IsNotNil)
				c.Assert(*value.Number, qt.Equals, float64(0))
				c.Assert(value.Reason, qt.Equals, "")
			}
		})
	}
}
