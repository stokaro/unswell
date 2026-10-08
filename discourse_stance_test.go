package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestDiscourseStanceSeparatesOperationalContinuation(t *testing.T) {
	for _, row := range []struct{ text, matched string }{
		{
			"It creates a temporary table, which is exactly the question, and the temporary table is removed after the check.",
			"which is exactly the question",
		},
		{
			"A temporary table errors with `ERROR 1061`, which is precisely the point, but the live table remains unchanged.",
			"which is precisely the point",
		},
		{
			"It publishes the text as the enum — which is the case worth having, because the dialect supports the declaration.",
			"which is the case worth having",
		},
		{"The migration is retained; that is the part worth keeping, and its rollback stays empty.", "that is the part worth keeping"},
		{"Applying that plan is what closes the loop.", "Applying that plan is what closes the loop"},
		{"Following these steps is what completes the cycle.", "Following these steps is what completes the cycle"},
		{
			"That is the honest conversion of a migration that never had a rollback:",
			"That is the honest conversion of a migration that never had a rollback",
		},
		{"This was an honest interpretation of the omitted rollback.", "This was an honest interpretation of the omitted rollback"},
		{"The non-zero status is the useful part.", "The non-zero status is the useful part"},
		{"The error is the important detail.", "The error is the important detail"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.evaluative-closure", row.text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.RuleVersion, qt.Equals, "12")
			c.Assert(f.Primary.Snippet, qt.Equals, row.matched)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(row.text, row.matched))
			c.Assert(f.Message, qt.Contains, "appraises")
			c.Assert(f.Evidence.Suggestion, qt.Contains, "absent operations")
			c.Assert(f.Evidence.Suggestion, qt.Contains, "Verify the basis before removing or weakening any claim")
		})
	}
}

func TestDiscourseStanceLiteralAndQualifiedControls(t *testing.T) {
	for _, text := range []string{
		"Applying that plan is what closes the feedback loop.",
		"The controller follows a wiring diagram; applying that plan is what closes the loop.",
		"The signal reaches the sensor, which is exactly the question, and the feedback circuit records it.",
		"Adding this edge is what closes the loop in the graph.",
		"Applying that plan closes the loop between the motor and the servo.",
		"The non-zero status is the useful part because it stops CI.",
		"The median latency is the important detail.",
		"The survey records the criteria; the error is the important detail.",
		"The error is not the important detail.",
		"That is not the honest conversion of the migration.",
		"That is the honest conversion of a migration if no rollback was declared.",
		"When the check succeeds, following those steps is what completes the cycle.",
		"Is applying that plan what closes the loop?",
		"The author says that applying that plan is what closes the loop.",
		"The author said: that is the honest conversion of a migration.",
		"\"Applying that plan is what closes the loop.\"",
		"The example preserves `The non-zero status is the useful part.`",
		"Applying `that plan` is what closes the loop.",
		"Applying that plan is what `closes` the loop.",
		"That is the `honest` conversion of a migration.",
		"The error is the `important` detail.",
		"That is a faithful conversion of a migration that never had a rollback.",
		"The protocol needs an honest majority.",
		"The parameter names the useful part of the response.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestDiscourseStanceMappingsAndPolicy(t *testing.T) {
	id := "filler.evaluative-closure"
	for _, source := range []document.Source{
		{
			Name: "guide.md", Format: document.Markdown,
			Bytes: []byte("\ufeffCafé 🙂.\r\n\r\nIt creates `row_2`, **which** is exactly the question, and the row is removed.\r\n"),
		},
		{
			Name: "guide.mdx", Format: document.MDX,
			Bytes: []byte("export const demo = 'Applying that plan is what closes the loop';\n\n" +
				"Applying that plan is what closes the loop.\n"),
		},
		{Name: "client.go", Format: document.Go, Bytes: []byte("package p\n// Applying that plan is what closes the loop.\nfunc Run() {}\n")},
		{Name: "client.py", Format: document.Python, Bytes: []byte("'''Applying that plan is what closes the loop.'''\n")},
	} {
		t.Run(source.Name, func(t *testing.T) {
			c := qt.New(t)
			r, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]), qt.Equals, f.Primary.Snippet)
			c.Assert(f.Primary.Snippet, qt.Not(qt.Contains), "row_2")
			c.Assert(f.Primary.Snippet, qt.Not(qt.Contains), "and the row")
		})
	}
	c := qt.New(t)
	text := "Applying that plan is what closes the loop. The non-zero status is the useful part."
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 2, saturation_occurrences: 3}", "").Findings, qt.HasLen, 0)
	approved := windowTerm(id, "Applying that plan is what closes the loop")
	filtered := singleRuleResult(t, id, text, "", approved)
	c.Assert(filtered.Findings, qt.HasLen, 1)
	c.Assert(filtered.Findings[0].Primary.Snippet, qt.Equals, "The non-zero status is the useful part")
	c.Assert(singleRuleResult(t, id, "Applying that plan is what closes the loop.", "", approved).Findings, qt.HasLen, 0)
	code := "The client requires a key.\n\n```text\nApplying that plan is what closes the loop.\n```"
	c.Assert(singleRuleResult(t, id, code, "", "").Findings, qt.HasLen, 0)
}

func TestDiscourseStanceBothProfilesAndMixedGuidance(t *testing.T) {
	const text = "That is the honest conversion of a migration that never had a rollback. That is measured rather than assumed."
	for _, profile := range []string{"technical", "strict"} {
		t.Run(profile, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{
				Config: []byte("version: 1\nextends: [builtin:" + profile + "]\n"), IncludeSource: true,
			})
			c.Assert(err, qt.IsNil)
			r, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
			c.Assert(err, qt.IsNil)
			found := false
			for _, f := range r.Findings {
				if f.RuleID != "filler.evaluative-closure" {
					continue
				}
				found = true
				c.Assert(f.Related, qt.HasLen, 1)
				c.Assert(f.Evidence.Suggestion, qt.Contains, "absent operations")
				c.Assert(f.Evidence.Suggestion, qt.Contains, "Verify the basis")
				c.Assert(f.Primary.Snippet, qt.Contains, "never had a rollback")
			}
			c.Assert(found, qt.IsTrue)
		})
	}
}
