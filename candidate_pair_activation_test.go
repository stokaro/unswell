package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestCandidateActivationsUseActualPairsAndWindows(t *testing.T) {
	const absentPair = "inapplicable/no_eligible_pair"
	const unsupported = "inapplicable/unsupported_unit"
	const missingWords = "inapplicable/insufficient_words"
	const missingTokens = "inapplicable/no_eligible_tokens"
	const missingList = "inapplicable/no_eligible_list"
	const missingWindow = "inapplicable/no_eligible_window"
	const heading = "A practical approach to the delivery process"
	near := overlapParagraph + "\n\n" + strings.Replace(overlapParagraph, "opens", "creates", 1)
	contrast := nearContrastSentence + "\n\n" + strings.Replace(nearContrastSentence, "may", "must", 1)
	for _, row := range []struct {
		name, id, text, parameters string
		reasons                    []string
		numbers                    []float64
		findings                   int
	}{
		{"near contrast", "repetition.near-sentence", contrast, "", []string{"", ""}, []float64{0, 0}, 0},
		{"near match", "repetition.near-sentence", near, "", []string{"", ""}, []float64{1, 1}, 1},
		{"no shared bigram", "repetition.near-sentence", overlapParagraph + "\n\n" +
			"Editors review written documents carefully before publishing updated manuals for interested readers around town.", "",
			[]string{absentPair, absentPair}, nil, 0},
		{"overlap contrast", "repetition.paragraph-overlap", contrast, "", []string{"", ""}, []float64{0, 0}, 0},
		{"overlap singleton windows", "repetition.paragraph-overlap", near, "{window_blocks: 1}",
			[]string{absentPair, absentPair}, nil, 0},
		{"overlap allowed pair", "repetition.paragraph-overlap", near, "{allowed_occurrences: 2}",
			[]string{"", ""}, []float64{0, 0}, 0},
		{"overlap protected partner", "repetition.paragraph-overlap", overlapParagraph + "\n\n" +
			strings.Replace(overlapParagraph, "credentials", "`credentials`", 1), "",
			[]string{absentPair, missingTokens}, nil, 0},
		{"heading match", "repetition.heading-echo", "# " + heading + "\n\n" + heading + ".", "",
			[]string{"", ""}, []float64{1, 1}, 1},
		{"heading missing minimum", "repetition.heading-echo", "# Retry budget\n\n" + overlapParagraph, "",
			[]string{missingWords, absentPair}, nil, 0},
		{"heading protected partner", "repetition.heading-echo", "# " + heading + "\n\n" +
			strings.Replace(overlapParagraph, "credentials", "`credentials`", 1), "",
			[]string{absentPair, missingTokens}, nil, 0},
		{"heading fence boundary", "repetition.heading-echo", "# " + heading + "\n\n```go\nx()\n```\n\n" + heading + ".",
			"", []string{absentPair, absentPair}, nil, 0},
		{"summary contrast", "repetition.summary-echo", nearContrastSentence + "\n\n# Summary\n\n" +
			strings.Replace(nearContrastSentence, "may", "must", 1), "", []string{"", unsupported, ""}, []float64{0, 0}, 0},
		{"summary missing scope", "repetition.summary-echo", near, "", []string{absentPair, absentPair}, nil, 0},
		{"summary missing earlier source", "repetition.summary-echo", "# Summary\n\n" + overlapParagraph, "",
			[]string{unsupported, absentPair}, nil, 0},
		{"list allowed", "format.list-fragmentation", "- Clear reports\n- Useful summaries", "",
			[]string{"", ""}, []float64{0, 0}, 0},
		{"list cannot fit window", "format.list-fragmentation", "- Clear reports\n- Useful summaries", "{window_blocks: 1}",
			[]string{missingWindow, missingWindow}, nil, 0},
		{"reference list", "format.list-fragmentation", "# API Reference\n\n- Clear reports\n- Useful summaries", "",
			[]string{unsupported, missingList, missingList}, nil, 0},
		{"excluded list item", "format.list-fragmentation", "- `Clear reports`\n- Useful summaries", "",
			[]string{missingList}, nil, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			parameters := row.parameters
			if parameters == "" {
				parameters = "{}"
			}
			options := unswell.Options{Features: []string{"activation/" + row.id}, Config: []byte(
				"version: 1\nextends: [builtin:custom]\nrules:\n  " + row.id + ": {enabled: true, parameters: " + parameters + "}\n")}
			engine, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(row.text)}
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, row.findings)
			assertPhraseMeasurements(t, result, row.reasons, row.numbers)
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
