package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestRedundantGrammaticalSupport(t *testing.T) {
	for _, text := range []string{
		"The renderer also supports custom labels as well.",
		"The renderer also sends labels to workers as well.",
		"Also, connectors can be used in reverse directions as well.",
		"In addition to labels, the renderer also supports colors.",
		"Use separators like, for example, blank lines.",
		"Create a worker and then finally start the listener.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "repetition.redundant-support", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestRedundantGrammaticalSupportControls(t *testing.T) {
	for _, text := range []string{
		"The renderer also supports custom labels as well as colors.",
		"The renderer also parses labels, but the reader accepts colors as well.",
		"The renderer supports labels, and the reader supports them as well.",
		"The renderer also supports labels that the reader parses as well.",
		"The worker also takes this opportunity to monitor the container as well.",
		"The worker also gives readers a chance to inspect the state as well.",
		"The renderer does not support labels as well as the reader.",
		"In addition to labels, the renderer supports colors.",
		"The payload looks like, for example, an event.",
		"Start the worker, then connect, and finally close the stream.",
		"The renderer `also` supports labels as well.",
		"The renderer also supports labels `as well`.",
		"The guide says the renderer also supports labels as well.",
		"\"The renderer also supports labels as well.\"",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "repetition.redundant-support", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestStructuralWordingMappingAndExemptions(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"repetition.redundant-support", "The renderer also supports café labels as well"},
	} {
		c := qt.New(t)
		text := "\ufeff" + strings.Replace(row.text, "café", "**café**", 1) + ".\r\n"
		r := singleRuleResult(t, row.id, text, "", "")
		c.Assert(r.Findings, qt.HasLen, 1)
		c.Assert(r.Findings[0].Primary.Span.Start, qt.Equals, len("\ufeff"))
		c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(strings.TrimPrefix(text, "\ufeff"), ".\r\n"))
		c.Assert(singleRuleResult(t, row.id, row.text+".", "", windowTerm(row.id, row.text)).Findings, qt.HasLen, 0)
		c.Assert(singleRuleResult(t, row.id, row.text+".", "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings,
			qt.HasLen, 0)
	}
}
