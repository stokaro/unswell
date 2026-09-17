package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestRhetoricalArgumentRelations(t *testing.T) {
	for _, row := range []struct{ id, text, snippet string }{
		{"filler.evaluative-closure", "Several rows carry a decision worth stating.", "a decision worth stating"},
		{"filler.evaluative-closure", "Three cases expose a distinction worth remembering.", "a distinction worth remembering"},
		{"filler.evaluative-closure", "Importing it is what this verb is for.", "Importing it is what this verb is for"},
		{"filler.evaluative-closure", "Checking that is precisely the question this feature exists to answer.",
			"Checking that is precisely the question this feature exists to answer"},
		{"filler.evaluative-closure", "The operation fails, which is the incompatibility the baseline exists to prevent.",
			"which is the incompatibility the baseline exists to prevent"},
		{"filler.evaluative-closure", "The third row is the one that matters most.", "The third row is the one that matters most"},
		{"filler.evaluative-closure", "The last of those is the one that keeps the entry honest.",
			"The last of those is the one that keeps the entry honest"},
		{"filler.evaluative-closure", "The third row is the one that matters most and the implementations agree on it.",
			"The third row is the one that matters most"},
		{"filler.evaluative-closure", "This example is what counts here.", "This example is what counts here"},
		{"filler.unscoped-assurance", "The check fails, which is what an operator reviewing the result wants.",
			"which is what an operator reviewing the result wants"},
		{"filler.unscoped-assurance", "This is what readers understand.", "This is what readers understand"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, row.id, row.text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.Primary.Snippet, qt.Equals, row.snippet)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(row.text, row.snippet))
		})
	}
}

func TestRhetoricalArgumentControls(t *testing.T) {
	for id, texts := range map[string][]string{
		"filler.evaluative-closure": {
			"No decision worth stating follows from that test.",
			"That is not a fact worth remembering.",
			"The guide says this is a decision worth stating.",
			"The program returns `a decision worth stating`.",
			"Reading it is worth the cost of the lookup.",
			"The setting is optional and worth writing to isolate the key.",
			"The comment explains a decision worth testing.",
			"Allocating it is not what this verb is for.",
			"Holding the mutex is what the lock guard is for.",
			"Allocating it is what this verb is for when the pool is empty.",
			"The benchmark measured that this is what matters most.",
			"The sensor is what counts incoming requests.",
			"The operator is the one that keeps the entry honest.",
			"The entry is the one that keeps the server honest.",
			"This is the measured result that matters most.",
			"Is importing it what this verb is for?",
		},
		"filler.unscoped-assurance": {
			"That is what the server wants.",
			"This is what the operator requested.",
			"This is what readers want if the network fails.",
			"This is what readers do not want.",
			"This is what the surveyed readers want.",
			"The survey measured what readers want.",
			"The guide says this is what readers understand.",
			"This is what the reader looking at the requested report wants.",
			"This is what users want, according to the survey.",
			"This is what a reader reviewing `the result` wants.",
		},
	} {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) {
				qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
			})
		}
	}
}

func TestRhetoricalRelationBoundsAndPolicy(t *testing.T) {
	c := qt.New(t)
	id := "filler.evaluative-closure"
	claim := "a decision worth stating"
	text := "\ufeffThree cases expose " + claim + ".\r\n"
	result := singleRuleResult(t, id, text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, strings.Index(text, claim))
	c.Assert(singleRuleResult(t, id, strings.Repeat("long ", 100)+text, "", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "", windowTerm(id, claim)).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
}
