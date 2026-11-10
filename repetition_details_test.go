package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

const detailsOpening = "Install the tool using any of the following package managers:"

func TestRepetitionSeparatesMarkdownDetailsChoices(t *testing.T) {
	for _, id := range []string{"repetition.paragraph-openers", "repetition.sentence-openers", "repetition.near-sentence"} {
		t.Run(id, func(t *testing.T) {
			c := qt.New(t)
			text := "# Installation\n\n"
			for _, choice := range []string{"Android", "BSD", "Linux"} {
				text += "<details>\n<summary>" + choice + "</summary>\n\n" + detailsOpening +
					"\n\nUse the " + choice + " package manager to install the command.\n\n</details>\n\n"
			}
			result := singleRuleResult(t, id, text, "", "")
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Findings, qt.HasLen, 0)
		})
	}
}

func TestRepetitionStillFindsRepeatedOpenersWithinOneDetailsSection(t *testing.T) {
	for _, id := range []string{"repetition.paragraph-openers", "repetition.sentence-openers"} {
		t.Run(id, func(t *testing.T) {
			c := qt.New(t)
			text := "<details>\n<summary>One choice</summary>\n\n" +
				strings.Repeat(detailsOpening+"\n\n", 3) + "</details>\n"
			result := singleRuleResult(t, id, text, "", "")
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, detailsOpening)
		})
	}
}

func TestRepetitionIgnoresDetailsTextInCodeAndComments(t *testing.T) {
	for _, prefix := range []string{
		"```html\n<details>\n</details>\n```\n\n",
		"<!--\n<details>\n</details>\n-->\n\n",
	} {
		t.Run(prefix[:3], func(t *testing.T) {
			c := qt.New(t)
			text := prefix + strings.Repeat(detailsOpening+"\n\n", 3)
			result := singleRuleResult(t, "repetition.paragraph-openers", text, "", "")
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Findings, qt.HasLen, 1)
		})
	}
}

func TestRepetitionKeepsParentScopeAroundNestedDetails(t *testing.T) {
	c := qt.New(t)
	text := "<details open>\n<summary>All systems</summary>\n\n" + detailsOpening + "\n\n" +
		"<details>\n<summary>Android</summary>\n\n" + detailsOpening + "\n\n</details>\n\n" +
		"<details>\n<summary>BSD</summary>\n\n" + detailsOpening + "\n\n</details>\n\n" +
		strings.Repeat(detailsOpening+"\n\n", 2) + "</details>\n"
	result := singleRuleResult(t, "repetition.paragraph-openers", text, "", "")
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Related, qt.HasLen, 2)
	first := strings.Index(text, detailsOpening)
	last := strings.LastIndex(text, detailsOpening)
	previous := strings.LastIndex(text[:last], detailsOpening)
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, first)
	c.Assert(result.Findings[0].Related[0].Span.Start, qt.Equals, previous)
	c.Assert(result.Findings[0].Related[1].Span.Start, qt.Equals, last)
}

func TestRepetitionDoesNotScopeUnmatchedDetailsTag(t *testing.T) {
	c := qt.New(t)
	text := "<details>\n<summary>Unclosed</summary>\n\n" +
		strings.Repeat(detailsOpening+"\n\n", 3)
	result := singleRuleResult(t, "repetition.paragraph-openers", text, "", "")
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Findings, qt.HasLen, 1)
}
