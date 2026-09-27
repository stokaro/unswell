package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestUnnamedNumericalChoice(t *testing.T) {
	for _, text := range []string{
		"Four answers, and the right one depends on whether you control the writes.",
		"3 options, but the best one depends on your deployment environment.",
		"Twenty choices, and the right one depends on whether retries are disabled.",
		"99 answers, and the best one depends on the requested precision.",
		"Four answers, and the right one depends on whether you do not control writes.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.unnamed-numerical-choice", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
			c.Assert(r.Findings[0].Related, qt.HasLen, 0)
			c.Assert(r.Findings[0].Evidence.Suggestion, qt.Contains, "Preserve the stated condition")
		})
	}
}

func TestUnnamedNumericalChoiceControls(t *testing.T) {
	for _, text := range []string{
		"Choose a consistency mode based on whether you control source writes.",
		"Two schemes: env and file.",
		"Four migration paths are available: validate, baseline, checkpoint, and replay.",
		"Four answers are required by the protocol, and the right one depends on the flags.",
		"Four answers, and the right one depends on exactly four required fields.",
		"Four answers, and the right one depends on the flags; exactly four answers are required.",
		"Four bytes encode the length.",
		"Three simultaneous conditions are required for a data race.",
		"Four answers, and the right one does not depend on whether you control writes.",
		"Four answers, and the right one depends on whether you control writes?",
		"The guide says: Four answers, and the right one depends on whether you control writes.",
		"\"Four answers, and the right one depends on whether you control writes.\"",
		"`Four answers`, and the right one depends on whether you control writes.",
		"Four answers, and the `right one` depends on whether you control writes.",
		"## Four answers, and the right one depends on whether you control writes",
		"- Four answers, and the right one depends on whether you control writes.",
		"The quiz asks how deployment works. Four answers, and the right one depends on whether you control writes.",
		"Four answers. The right one depends on whether you control writes.",
		"Four options, and the right one depends on the mode: local, hosted, gateway, or offline.",
		"Four answers, and the right one depends on the mode; local or hosted.",
		"One answer, and the right one depends on whether you control writes.",
		"100 options, and the right one depends on whether you control writes.",
		"Four answers, and the right one depends on.",
		"Four answers, and the right one depends on " + strings.Repeat("the deployment configuration ", 20) + ".",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unnamed-numerical-choice", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestUnnamedNumericalChoiceMappingAndPolicy(t *testing.T) {
	c := qt.New(t)
	const id = "filler.unnamed-numerical-choice"
	const text = "\ufeffFour **answers**, and the right one depends on\r\nwhether café writes are enabled.\r\n"
	r := singleRuleResult(t, id, text, "", "")
	c.Assert(r.Findings, qt.HasLen, 1)
	f := r.Findings[0]
	c.Assert(f.Primary.Span.Start, qt.Equals, len("\ufeff"))
	c.Assert(f.Primary.Span.End, qt.Equals, strings.Index(text, "."))
	c.Assert(f.Primary.Snippet, qt.Equals, text[f.Primary.Span.Start:f.Primary.Span.End])
	c.Assert(singleRuleResult(t, id, text, "", windowTerm(id, "Four answers, and the right one depends on whether café writes are enabled")).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
}
