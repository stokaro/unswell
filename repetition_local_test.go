package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestAdjacentWordRepetition(t *testing.T) {
	for _, row := range []struct {
		text, duplicate string
	}{
		{"The scanner requires requires an external decoder.", "requires"},
		{"See See the release notes.", "See"},
		{"The the client retries once.", "the"},
		{"The client had had enough time.", ""},
		{"We know that that request failed.", ""},
		{"They do do extra checks.", ""},
		{"The cache is very very small.", ""},
		{"The police police the area.", ""},
		{"The fish fish in shallow water.", ""},
		{"We saw her duck duck.", ""},
		{"The scanner requires, requires an external decoder.", ""},
		{"The scanner `requires requires` an external decoder.", ""},
		{"The example says \"See See the release notes\".", ""},
		{"The example says 'See See the release notes'.", ""},
		{"No, no retry is needed.", ""},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "repetition.adjacent-word", row.text, "", "")
			if row.duplicate == "" {
				c.Assert(result.Findings, qt.HasLen, 0)
				return
			}
			c.Assert(result.Findings, qt.HasLen, 1)
			f := result.Findings[0]
			c.Assert(f.Related, qt.HasLen, 1)
			c.Assert(strings.ToLower(f.Primary.Snippet), qt.Equals, strings.ToLower(row.duplicate))
			c.Assert(f.Related[0].Snippet, qt.Equals, row.duplicate)
			c.Assert(f.Primary.Span.End < f.Related[0].Span.Start, qt.IsTrue)
		})
	}
}

func TestRepeatedClaimsRetainExactOperandsAndConditions(t *testing.T) {
	const first = "The command exits `1` because drift was found"
	for _, row := range []struct {
		name, left, right, separator string
		want                         int
	}{
		{"complete assertions", first + ".", first + ".", "\n\n", 1},
		{"relative tail", first + ", which confirms the check worked.", first + ".", "\n\n", 1},
		{"across output", first + ", which confirms the check worked.", first + ".", "\n\n```text\nchanged\n```\n\n", 1},
		{"both partial", first + ", which produces a report.", first + ", which produces an audit record.", "\n\n", 0},
		{"different exit code", first + ".", strings.Replace(first, "`1`", "`0`", 1) + ".", "\n\n", 0},
		{"different condition", first + ".", strings.Replace(first, "drift", "damage", 1) + ".", "\n\n", 0},
		{"negated", first + ".", strings.Replace(first, "was found", "was not found", 1) + ".", "\n\n", 0},
		{"later condition", first + ".", first + " only when enabled.", "\n\n", 0},
		{"qualified tail", first + ", which happens only when enabled.", first + ".", "\n\n", 0},
		{"modal", first + ".", strings.Replace(first, "exits", "may exit", 1) + ".", "\n\n", 0},
		{"different section", first + ".", first + ".", "\n\n## A separate command\n\n", 0},
		{"quoted examples", "\"" + first + ".\"", "\"" + first + ".\"", "\n\n", 0},
		{"arbitrary code", strings.Replace(first, "`1`", "`a + b`", 1) + ".", strings.Replace(first, "`1`", "`a + b`", 1) + ".", "\n\n", 0},
		{"separate operands", "The decoder returns `Token` on failure.", "The decoder returns `token` on failure.", "\n\n", 0},
		{"separate links", "The decoder returns [errors](a) on failure.", "The decoder returns [errors](b) on failure.", "\n\n", 0},
		{"independent instructions", "- The command exits on failure.", "- The command exits on failure.", "\n", 0},
		{"short unprotected", "The client retries the request.", "The client retries the request.", "\n\n", 1},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			text := "\ufeff## Checking\r\n\r\n" + row.left + row.separator + row.right
			result := singleRuleResult(t, "repetition.repeated-claim", text, "", "")
			c.Assert(result.Findings, qt.HasLen, row.want)
			if row.want == 0 {
				return
			}
			f := result.Findings[0]
			c.Assert(f.Related, qt.HasLen, 1)
			c.Assert(f.Primary.Snippet, qt.Equals, f.Related[0].Snippet)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(text, f.Primary.Snippet))
			c.Assert(f.Related[0].Span.Start, qt.Equals, strings.LastIndex(text, f.Related[0].Snippet))
		})
	}
}

func TestRepeatedClaimsBoundWindowAndMinimum(t *testing.T) {
	c := qt.New(t)
	const sentence = "The client retries the request."
	text := sentence + " The server closes connections. " + sentence
	c.Assert(singleRuleResult(t, "repetition.repeated-claim", text, "{window_sentences: 2}", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, "repetition.repeated-claim", text, "{window_sentences: 3}", "").Findings, qt.HasLen, 1)
	c.Assert(singleRuleResult(t, "repetition.repeated-claim", text, "{min_words: 6}", "").Findings, qt.HasLen, 0)
}

func TestExplanatoryRestarts(t *testing.T) {
	for _, row := range []struct {
		text string
		want int
	}{
		{"The process is complex, and it is complex because several callbacks share state.", 1},
		{"The reasons for why this is are quite complex, and they are complex because the optimizations " +
			"that ripgrep uses to implement fast search are complex.", 1},
		{"The operation was slow, and it was slow because the disk was full.", 1},
		{"The process is complex because several callbacks share state.", 0},
		{"The process is slow, and it is fast because the cache is warm.", 0},
		{"The process is slow, and it was slow because the disk was full.", 0},
		{"The process is not slow, and it is slow because the disk was full.", 0},
		{"The process is slow, and it is slow only when the disk is full.", 0},
		{"The cache and the process are slow, and they are slow because the disk is full.", 0},
		{"The process is slow, and `it is slow` because the disk is full.", 0},
		{"If the disk is full, the process is slow, and it is slow because cleanup is running.", 0},
		{"The process is slow, and it is very slow because the disk is full.", 0},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "repetition.explanatory-restart", row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, row.want)
			if row.want > 0 {
				c.Assert(result.Findings[0].Related, qt.HasLen, 1)
				c.Assert(result.Findings[0].Related[0].Snippet, qt.Not(qt.Contains), "because")
			}
		})
	}
}

func TestRepeatedClaimsChargeOnlyMappingWalksThatRun(t *testing.T) {
	c := qt.New(t)
	const unsupported = "The engine uses `open file` to read the selected source and preserve its original contents.\n\n"
	text := strings.Repeat(unsupported, 200) + "The client retries the request. The client retries the request."
	result := singleRuleResult(t, "repetition.repeated-claim", text, "", "analysis: {max_candidates: 10000}\n")
	c.Assert(result.Abstentions, qt.HasLen, 0)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "The client retries the request")
}

func TestLocalRepetitionLeavesUnresolvedParaphrasesUnclaimed(t *testing.T) {
	for _, text := range []string{
		"Definitions come from one source, `docs/site/src/glossary.ts`, so the same term cannot come to mean two " +
			"different things on two pages.\n\nThen link here from the pages that use it.\n\n" +
			"The definition stays in the map and is rendered in one place, which is what keeps a word from meaning " +
			"one thing on one page and something else on the next.",
		"In particular, ripgrep's dependencies (direct and transitive) will always be limited to permissive licenses. " +
			"That is, ripgrep will never depend on code that is not permissively licensed.",
		"The client uses permissive licenses for direct dependencies. " +
			"The server uses permissive licenses for transitive dependencies.",
	} {
		for _, id := range []string{"repetition.repeated-claim", "repetition.explanatory-restart"} {
			t.Run(id+"/"+text, func(t *testing.T) {
				c := qt.New(t)
				c.Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
			})
		}
	}
}
