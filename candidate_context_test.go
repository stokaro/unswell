package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestCandidateActivationsRespectSelectedContext(t *testing.T) {
	const absentPair = "inapplicable/no_eligible_pair"
	const unsupported = "inapplicable/unsupported_unit"
	for _, row := range []struct {
		id, text string
		reasons  []string
	}{
		{"repetition.heading-echo", "# " + overlapParagraph + "\n\n" + overlapParagraph, []string{absentPair}},
		{"repetition.summary-echo", overlapParagraph + "\n\n# Summary\n\n" + overlapParagraph,
			[]string{absentPair, absentPair}},
		{"format.list-fragmentation", "- Clear reports\n- Useful summaries\n\n" + overlapParagraph, []string{unsupported}},
	} {
		t.Run(row.id, func(t *testing.T) {
			c := qt.New(t)
			options := unswell.Options{Features: []string{"activation/" + row.id}, Config: []byte(
				"version: 1\nextends: [builtin:custom]\nextraction: {contexts: [paragraph]}\nrules:\n  " + row.id +
					": {enabled: true}\n")}
			engine, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(row.text)}
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, 0)
			assertPhraseMeasurements(t, result, row.reasons, nil)
			options.Features = nil
			ordinary, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			want, err := ordinary.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			result.Features = nil
			c.Assert(result, qt.DeepEquals, want)
		})
	}
}
