package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestEvidencePosture(t *testing.T) {
	for _, text := range []string{
		"That is measured rather than assumed.",
		"This is verified rather than asserted.",
		"That is proven rather than assumed.",
		"That was checked rather than claimed.",
		"This is empirically tested rather than simply promised.",
		"These statements are validated rather than assumed.",
		"The conclusion is demonstrated, not guessed.",
		"The claim was verified and not merely asserted.",
		"What it covers is measured rather than asserted: the matrix records each capability.",
		"What this guide describes is tested rather than promised.",
		"The `format`, `migration` and `schema` rows stay `env`-only, and that is measured rather than assumed.",
		"The program accepts two inputs, and that is verified rather than asserted.",
		"This is checked rather than claimed, and the command returns a classified error.",
		"The result agrees; it is verified rather than asserted.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.evaluative-closure", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Message, qt.Contains, "comments on verification")
			c.Assert(r.Findings[0].Evidence.Suggestion, qt.Contains, "preserving its factual claims")
		})
	}
}

func TestEvidencePostureControls(t *testing.T) {
	for _, text := range []string{
		"The latency is measured rather than assumed.",
		"The reported dimension is verified rather than assumed.",
		"It is measured rather than assumed.",
		"The latency changes; it is measured rather than assumed.",
		"The result is 45; it is verified rather than assumed.",
		"The worker verifies the dimension rather than trusting the provider.",
		"This is verified against the recorded result rather than assumed.",
		"This is verified rather than assumed by comparing the two checksums.",
		"This is verified rather than assumed, using the two recorded checksums.",
		"This is measured rather than calculated.",
		"That is estimated rather than measured.",
		"That is not measured rather than assumed.",
		"That is only verified rather than assumed.",
		"That could be measured rather than assumed.",
		"That is measured rather than assumed if the probe succeeds.",
		"When the probe succeeds, that is measured rather than assumed.",
		"The result differs because that is measured rather than assumed.",
		"The author says that this is measured rather than assumed.",
		"The reader claimed: that is measured rather than asserted.",
		"\"This is verified rather than asserted.\"",
		"The example preserves its wording: `This is measured rather than assumed.`",
		"That is `measured` rather than assumed.",
		"This is verified rather than `asserted`.",
		"Is that measured rather than assumed?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestEvidencePostureSourceMapping(t *testing.T) {
	text := "\ufeffRésumé 🙂.\r\n\r\nThe `format` block stays local, and **that** is measured\r\nrather than assumed."
	r := singleRuleResult(t, "filler.evaluative-closure", text, "", "")
	c := qt.New(t)
	c.Assert(r.Findings, qt.HasLen, 1)
	f := r.Findings[0]
	c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(text, "that**"))
	c.Assert(f.Primary.Snippet, qt.Equals, "that** is measured\r\nrather than assumed")
	c.Assert(text[f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, f.Primary.Snippet)
}
