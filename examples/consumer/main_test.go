package main

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestPublicConsumer(t *testing.T) {
	c := qt.New(t)
	result, err := analyze(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].RuleID, qt.Equals, "team.avoid-magic")
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, 17)
}
