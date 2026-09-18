package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestInstructionNominalMethodAcrossBlocks(t *testing.T) {
	const ruleID = "filler.instruction-scaffolding"
	for _, method := range []string{
		"This is done by the routing notation.",
		"This can be achieved through the routing syntax.",
		"This is configured with the label option.",
		"This is done using the routing directive.",
		"This is done by the notation below:",
		"This is done by notation.",
		"This is configured with the **café** option.",
		"This is done using the `route_pattern` attribute.",
	} {
		t.Run(method, func(t *testing.T) {
			c := qt.New(t)
			text := "It is possible to configure a route.\n\n" + method
			result := singleRuleResult(t, ruleID, text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Related, qt.HasLen, 1)
			loc := result.Findings[0].Related[0]
			c.Assert(loc.Span.Start, qt.Equals, strings.Index(text, method))
			c.Assert(text[loc.Span.Start:loc.Span.End], qt.Equals, strings.TrimRight(method, ".:"))
		})
	}
}

func TestInstructionNominalMethodKeepsBoundaries(t *testing.T) {
	const ruleID = "filler.instruction-scaffolding"
	for _, tail := range []string{
		"\n\nThis is done by the next worker.",
		"\n\nThis is done by the command processor.",
		"\n\nThis is done by the `notation`.",
		"\n\nThis is not done by the notation.",
		"\n\nThis is done by the notation only after startup.",
		"\n\n# Another operation\n\nThis is done by the notation.",
		"\n\nThe worker starts.\n\nThis is done by the notation.",
		"\n\n- This is done by the notation.",
	} {
		t.Run(tail, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, ruleID, "It is possible to configure a route."+tail, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Related, qt.HasLen, 0)
		})
	}
	// A noun instrument alone does not turn an ordinary passive into a warning.
	qt.New(t).Assert(singleRuleResult(t, ruleID, "This is done by the notation.", "", "").Findings, qt.HasLen, 0)
}
