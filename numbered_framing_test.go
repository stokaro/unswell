package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNumberedSectionFraming(t *testing.T) {
	for _, text := range []string{
		"Four lines of a specification decide where your corpus goes. This page explains the configuration.\n\n## The four lines",
		"Five lines of configuration determine the endpoint.\n\n## The five lines",
		"10 rules determine how a request is handled.\n\n## The ten rules",
		"Three steps control the deployment process.\n\nThe 3 steps\n-----------",
		"Just six points explain the setup.\n\n## These six points",
		"Four **lines** of a specification decide where café requests go.\n\n## The four **lines**",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.numbered-section-framing", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Related, qt.HasLen, 1)
			c.Assert(r.Findings[0].Primary.Span.Start, qt.Equals, 0)
			c.Assert(r.Findings[0].Evidence.Suggestion, qt.Contains, "Name the configuration or task")
		})
	}
}

func TestNumberedSectionFramingControls(t *testing.T) {
	for _, text := range []string{
		"Two schemes: env and file.\n\n## The two schemes",
		"Two fixed strings leave the process.\n\n## The two strings",
		"Four bytes encode the length.\n\n## The four bytes",
		"Four lines are required by the protocol.\n\n## The four lines",
		"Four lines must terminate the message.\n\n## The four lines",
		"Four lines control the endpoint; exactly four are required by the protocol.\n\n## The four lines",
		"Four lines do not control the endpoint.\n\n## The four lines",
		"Four steps initialize the database.\n\n## Four steps to initialize the database",
		"Four lines control the endpoint.\n\n## Four lines: configuration",
		"Four steps control deployment.\n\n## Four steps — setup",
		"Three rules control deployment.\n\n## Three rules; exceptions",
		"Four lines control the endpoint.\n\n## Four lines: `model`",
		"Four lines of configuration control the endpoint.\n\n## Provider configuration",
		"Four lines of configuration control the endpoint.\n\n## The five lines",
		"Four lines of configuration control the endpoint.\n\n## The four steps",
		"Four lines of configuration control the endpoint.\n\nA separate operation follows.\n\n## The four lines",
		"Four lines of configuration control the endpoint.\n\n```yaml\na: b\n```\n\n## The four lines",
		"Four lines of configuration control the endpoint.\n\n<!-- boundary -->\n\n## The four lines",
		"Four lines of configuration control the endpoint.\n\n- A boundary\n\n## The four lines",
		"Four lines of configuration control the endpoint.\n\n## Other heading\n\n## The four lines",
		"`Four lines` of configuration control the endpoint.\n\n## The four lines",
		"Four lines of configuration control the endpoint.\n\n## The `four lines`",
		"\"Four lines of configuration control the endpoint.\"\n\n## The four lines",
		"The guide says four lines control the endpoint.\n\n## The four lines",
		"Do four lines control the endpoint?\n\n## The four lines",
		"Four lines control the endpoint.\n\n## The four lines?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.numbered-section-framing", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestNumberedSectionFramingMappingAndPolicy(t *testing.T) {
	c := qt.New(t)
	const id = "filler.numbered-section-framing"
	text := "\ufeffFour **lines** decide where café requests go.\r\n\r\n## The four **lines**\r\n"
	r := singleRuleResult(t, id, text, "", "")
	c.Assert(r.Findings, qt.HasLen, 1)
	f := r.Findings[0]
	c.Assert(f.Primary.Span.Start, qt.Equals, len("\ufeff"))
	c.Assert(f.Primary.Snippet, qt.Equals, "Four **lines** decide where café requests go")
	c.Assert(f.Related[0].Span.Start, qt.Equals, strings.Index(text, "The four"))
	c.Assert(f.Related[0].Snippet, qt.Equals, "The four **lines")
	c.Assert(singleRuleResult(t, id, text, "", windowTerm(id, "The four lines")).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
}
