package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestPurposeAndReaderConstructions(t *testing.T) {
	for _, row := range []struct{ id, text, snippet string }{
		{"filler.evaluative-closure", "That second case is what the finding is for.", "That second case is what the finding is for"},
		{"filler.evaluative-closure", "This is the fact the whole feature exists for.",
			"This is the fact the whole feature exists for"},
		{"filler.evaluative-closure", "The result survives, which is the part that matters:", "which is the part that matters"},
		{"filler.evaluative-closure", "The query returns wrong rows, which is the failure mode worth knowing about before you meet it.",
			"which is the failure mode worth knowing about before you meet it"},
		{"filler.evaluative-closure", "That direction is deliberate: loading happens before dialect selection.",
			"That direction is deliberate"},
		{"filler.evaluative-closure", "The rendered SQL proves Ptah understood the desired schema.",
			"The rendered SQL proves Ptah understood the desired schema"},
		{"filler.evaluative-closure", "The result survives, which is the case worth having because the backend lacks enums.",
			"which is the case worth having"},
		{"filler.evaluative-closure", "The two variants differ because they share storage, but that distinction is the whole value of the verb:",
			"that distinction is the whole value of the verb"},
		{"filler.unscoped-assurance", "That is the right answer nearly always, and the wrong one exactly where storage differs:",
			"That is the right answer nearly always"},
		{"filler.unscoped-assurance", "Most people want the fallback.", "Most people want the fallback"},
		{"filler.unscoped-assurance", "That is what most people do.", "That is what most people do"},
		{"filler.unscoped-assurance", "That is almost never what you want.", "That is almost never what you want"},
		{"filler.unscoped-assurance", "The defaults are safe for most runs.", "The defaults are safe for most runs"},
		{"filler.unscoped-assurance", "The check uses the policy most production environments run.",
			"most production environments run"},
		{"filler.document-justification", "This section is dedicated to runtime authors.",
			"This section is dedicated to runtime authors"},
		{"filler.document-justification", "The rest of this document looks at content storage.",
			"The rest of this document looks at content storage"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, row.id, row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			f := result.Findings[0]
			c.Assert(f.Primary.Snippet, qt.Equals, row.snippet)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(row.text, row.snippet))
		})
	}
}

func TestPurposeAndReaderControls(t *testing.T) {
	for id, texts := range map[string][]string{
		"filler.evaluative-closure": {
			"The command starts a new batch, which is what the wrapper is for.",
			"That second case is not what the finding is for.",
			"Is that what the finding is for?",
			"That is what the finding is for when tracing is enabled.",
			"The result is the failure mode worth testing before release.",
			"The output does not prove the parser understood the desired schema.",
			"No output proves the parser understood the desired schema.",
			"The rendered SQL proves the parser copied the required identifier.",
			"The measured output proves the parser understood the desired schema.",
			"The guide says the rendered SQL proves Ptah understood the desired schema.",
			"That is the `fact the whole feature exists for`.",
			"The example is \"That direction is deliberate\".",
		},
		"filler.unscoped-assurance": {
			"Most surveyed users want the fallback.",
			"Most users want the fallback, according to the survey.",
			"Most users want the fallback only when the primary fails.",
			"Most users do not want the fallback.",
			"The survey measured that most users want the fallback.",
			"That is nearly always what you want if the cache is empty.",
			"That is what most people do, according to 70 respondents.",
			"The defaults are safe for most runs in the measured benchmark.",
			"Most environments run the test before deployment.",
			"The writer claims most production environments run this policy.",
		},
		"filler.document-justification": {
			"This section is not dedicated to runtime authors.",
			"The next process will describe the stored schema.",
			"The rest of this document looks at content storage when troubleshooting.",
		},
	} {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) {
				qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
			})
		}
	}
}

func TestPurposeTailLengthAndSourceMapping(t *testing.T) {
	c := qt.New(t)
	id := "filler.evaluative-closure"
	tail := "which is exactly the question"
	premise := "The server " + strings.Repeat("retries the request and ", 12) + "returns the answer, "
	text := "\ufeff" + premise + tail + ".\r\n"
	r := singleRuleResult(t, id, text, "", "")
	c.Assert(r.Findings, qt.HasLen, 1)
	c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, tail)
	c.Assert(r.Findings[0].Primary.Span.Start, qt.Equals, strings.Index(text, tail))
	c.Assert(singleRuleResult(t, id, strings.Repeat("long ", 100)+text, "", "").Findings, qt.HasLen, 0)
	claim := "That is what the finding is for"
	c.Assert(singleRuleResult(t, id, claim+".", "", windowTerm(id, claim)).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, claim+".", "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
}
