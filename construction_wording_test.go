package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestConstructionWording(t *testing.T) {
	for id, texts := range map[string][]string{
		"filler.evaluative-closure": {
			"The two constraints are worth knowing because the queue accepts a single writer.",
			"The second difference is worth stating before startup.",
			"The reason is worth knowing because it changes the retry limit.",
			"Requests are buffered, and the difference is worth knowing before a run.",
			"It should be noted that the client retries only after a timeout.",
			"It can also be noted that a script must preserve the exit code.",
		},
		"filler.instruction-scaffolding": {
			"It is possible to apply a theme to the dashboard.",
			"It is also possible to predefine classes in the style sheet.",
			"Using the preview mode, it is also possible to set labels on edges.",
			"There is the possibility to use bidirectional links.",
			"This is done by defining the label option.",
			"By helping to triage the issue: This can be done by providing a test case.",
			"Attachment of a class to a node is done as per below:",
			"If you need a horizontal divider you just need to put three dashes on a line.",
			"If we are just interested in the total across applications, we could simply write:",
		},
		"filler.document-justification": {
			"The count is not repeated here.",
			"Neither definition is repeated here.",
			"Read the reference instead of being repeated here.",
			"They are listed here rather than left to be discovered.",
			"Use the reference rather than this page restating a list that drifts.",
		},
		"repetition.definition-echo":     {"The subset is still a subset, and the reference lists its members."},
		"repetition.redundant-predicate": {"The category label describes the category."},
	} {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) {
				c := qt.New(t)
				r := singleRuleResult(t, id, text, "", "")
				c.Assert(r.Findings, qt.HasLen, 1)
				f := r.Findings[0]
				c.Assert(text[f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, f.Primary.Snippet)
				c.Assert(f.Evidence.Suggestion, qt.Not(qt.Equals), "")
			})
		}
	}
}

func TestConstructionWordingControls(t *testing.T) {
	for id, texts := range map[string][]string{
		"filler.evaluative-closure": {
			"It must be noted in the audit log before the process exits.",
			"It should not be noted that the request succeeded.",
			"The result is worth measuring because the cache changes latency.",
			"No constraint is worth knowing.", "The constraint is not worth knowing.",
			"The log contains `It should be noted that`.",
			"The author says \"The constraint is worth knowing\".",
			"It is worth noting that the client retries.",
		},
		"filler.instruction-scaffolding": {
			"This is done by the next worker.",
			"This is done by `defining` the interface.",
			"Configuration of the queue is performed by an administrator.",
			"The possibility to lose data is documented.",
			"If we have a pending request, we can retry it.",
			"If you need permission, configure the access policy.",
			"If you want to display a warning, use the template only after validation.",
			"The example says \"This is done by defining a label\".",
		},
		"filler.document-justification": {
			"The count is not repeated in the cache.",
			"The list is repeated here only when the token expires.",
			"They are listed here as required by the interface contract.",
			"The log contains `The count is not repeated here`.",
		},
		"repetition.definition-echo": {
			"The subset is still a proper subset.", "The subset is still a subset of the cached entries.",
			"The `subset` is still a `subset`.", "The subset is not a subset.",
			"The subset is still a subset when every member is present.",
		},
		"repetition.redundant-predicate": {
			"The category label describes the resource category.",
			"The category field identifies the category.",
			"The `category` label describes the category.",
		},
	} {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) { qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0) })
		}
	}
}

func TestConstructionNoticePreservesConditionsAndOffsets(t *testing.T) {
	c := qt.New(t)
	id := "filler.evaluative-closure"
	text := "\ufeffThe queue accepts two writers, and the **difference** is worth knowing because stale clients must retry.\r\n"
	r := singleRuleResult(t, id, text, "", "")
	c.Assert(r.Findings, qt.HasLen, 1)
	f := r.Findings[0]
	c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(text, "the **difference**"))
	c.Assert(f.Primary.Snippet, qt.Equals, "the **difference** is worth knowing because stale clients must retry")
	phrase := "The difference is worth knowing because stale clients must retry"
	c.Assert(singleRuleResult(t, id, phrase+".", "", windowTerm(id, phrase)).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, phrase+".", "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
}

func TestConstructionRewritesKeepConditions(t *testing.T) {
	for _, row := range []struct{ id, before, after string }{
		{"filler.evaluative-closure", "It should be noted that the client retries only after a timeout.", "The client retries only after a timeout."},
		{"filler.evaluative-closure", "Two type mappings are worth knowing because the engine has no Boolean type.", "The engine has no Boolean type; use these two type mappings."},
		{"filler.instruction-scaffolding", "It is possible to apply `compact.css` to the dashboard.", "You can apply `compact.css` to the dashboard."},
		{"filler.instruction-scaffolding", "This is done by defining `retry_limit` as `3`.", "Define `retry_limit` as `3` to configure retries."},
		{"filler.instruction-scaffolding", "If you need a horizontal divider you just need to put three dashes on a line.", "To insert a horizontal divider, put three dashes on a line."},
		{"filler.document-justification", "Use the reference rather than this page restating a list that drifts.", "Use the reference for the current list."},
	} {
		t.Run(row.id+"/"+row.before, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(singleRuleResult(t, row.id, row.before, "", "").Findings, qt.HasLen, 1)
			c.Assert(singleRuleResult(t, row.id, row.after, "", "").Findings, qt.HasLen, 0)
		})
	}
}
