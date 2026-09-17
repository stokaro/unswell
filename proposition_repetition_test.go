package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestRestrictionReformulations(t *testing.T) {
	for _, text := range []string{
		"The client accepts only signed requests. That is, the client never accepts requests that are not signed.",
		"The client accepts only signed requests. In other words, the client does not accept requests that are not signed.",
		"The workers process only validated requests. That is, the workers do not process requests that are not validated.",
		"The proxy will forward only authenticated requests. That is, the proxy will never forward requests that are not authenticated.",
		"The scanner reads only writable files. That is, the scanner does not read files that are not writable.",
		"In particular, ripgrep's dependencies (direct and transitive) will always be limited to permissive licenses. " +
			"That is, ripgrep will never depend on code that is not permissively licensed.",
		"The client's dependencies are limited to commercial licenses. " +
			"That is, the client never depends on code that is not commercially licensed.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "repetition.repeated-claim", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Evidence.Kind, qt.Equals, "heuristic")
			c.Assert(r.Findings[0].Related, qt.HasLen, 1)
			c.Assert(r.Findings[0].Evidence.Metrics[0].Name, qt.Equals, "restriction-restatements")
			c.Assert(r.Findings[0].Evidence.Suggestion, qt.Contains, "scope")
		})
	}
}

func TestRestrictionReformulationControls(t *testing.T) {
	for _, text := range []string{
		"The client accepts only signed requests. That is, the server never accepts requests that are not signed.",
		"The client accepts only signed requests. That is, the client never accepts requests that are not validated.",
		"The client accepts only signed requests. That is, the client never forwards requests that are not signed.",
		"The client accepts only signed requests. That is, the client never accepts messages that are not signed.",
		"The client accepts only signed requests. That is, the client never accepts requests that are signed.",
		"The client accepts only signed requests. That is, the client sometimes accepts requests that are not signed.",
		"The client accepts only signed requests. That is, the client will never accept requests that are not signed.",
		"The client may accept only signed requests. That is, the client never accepts requests that are not signed.",
		"The client accepts only signed requests when connected. That is, the client never accepts requests that are not signed.",
		"The client accepts only signed requests. That is, the client never accepts requests that are not signed unless disconnected.",
		"The client accepts only signed requests. The client never accepts requests that are not signed.",
		"The client accepts only signed requests. That is, it never accepts requests that are not signed.",
		"The client accepts only `signed` requests. That is, the client never accepts requests that are not `signed`.",
		"The client accepts only signed requests.\n\n## Another configuration\n\n" +
			"That is, the client never accepts requests that are not signed.",
		"Ripgrep's direct dependencies are limited to permissive licenses. " +
			"That is, ripgrep never depends on code that is not permissively licensed.",
		"Ripgrep's dependencies (direct only) are limited to permissive licenses. " +
			"That is, ripgrep never depends on code that is not permissively licensed.",
		"Ripgrep's dependencies are limited to permissive licenses. " +
			"That is, ripgrep never depends on code that is not commercially licensed.",
		"The [client](https://example.test/one) accepts only signed requests. " +
			"That is, the [client](https://example.test/two) never accepts requests that are not signed.",
		"The report says the client accepts only signed requests. " +
			"That is, the client never accepts requests that are not signed.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "repetition.repeated-claim", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}
