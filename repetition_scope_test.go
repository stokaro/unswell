package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestRepetitionScopesKeepIndependentVersionsAndCells(t *testing.T) {
	for _, id := range []string{"repetition.exact-sentence", "repetition.near-sentence",
		"repetition.sentence-openers", "repetition.paragraph-openers"} {
		t.Run(id, func(t *testing.T) {
			c := qt.New(t)
			items := []string{overlapParagraph, overlapParagraph, overlapParagraph}
			if id == "repetition.near-sentence" {
				items[1] = strings.Replace(overlapParagraph, "opens", "creates", 1)
				items[2] = strings.Replace(overlapParagraph, "opens", "starts", 1)
			}
			same := "# Release\n\n" + strings.Join(items, "\n\n")
			c.Assert(singleRuleResult(t, id, same, "", "").Findings, qt.HasLen, 1)
			for _, headings := range [][]string{{"## v1", "## v2", "## v3"}, {"## Notes", "## Notes", "## Notes"}} {
				var text strings.Builder
				for i, item := range items {
					text.WriteString(headings[i] + "\n\n" + item + "\n\n")
				}
				c.Assert(singleRuleResult(t, id, text.String(), "", "").Findings, qt.HasLen, 0)
			}
			table := "| Provider | Capability |\n| --- | --- |\n| A | " + items[0] + " |\n| B | " + items[1] + " |\n| C | " + items[2] + " |\n"
			c.Assert(singleRuleResult(t, id, table, "", "").Findings, qt.HasLen, 0)
			if id == "repetition.exact-sentence" || id == "repetition.near-sentence" {
				cell := "| Provider | Capability |\n| --- | --- |\n| A | " + strings.Join(items, " ") + " |\n"
				c.Assert(singleRuleResult(t, id, cell, "", "").Findings, qt.HasLen, 1)
			}
		})
	}
}

func TestRepetitionScopesKeepInstallationVariants(t *testing.T) {
	for _, id := range []string{"repetition.sentence-openers", "repetition.paragraph-openers", "repetition.near-sentence"} {
		t.Run(id, func(t *testing.T) {
			c := qt.New(t)
			var text strings.Builder
			for _, platform := range []string{"Homebrew", "MacPorts", "Chocolatey"} {
				text.WriteString("If you're a " + platform + " user, you can install the tool from the package manager.\n\n")
			}
			c.Assert(singleRuleResult(t, id, text.String(), "", "").Findings, qt.HasLen, 0)
			if id != "repetition.near-sentence" {
				repeated := strings.Repeat("If you're a Homebrew user, you can install the tool from the package manager.\n\n", 3)
				result := singleRuleResult(t, id, repeated, "", "")
				c.Assert(result.Findings, qt.HasLen, 1)
				c.Assert(result.Findings[0].Related, qt.HasLen, 2)
			}
		})
	}
}

func TestDuplicateListItemsRetainReferencesAndProtectedIdentity(t *testing.T) {
	const first = "add documentation for `Literal` type, #651 by @dmontagu"
	const plain = "add documentation for Literal type, #651 by @dmontagu"
	for _, row := range []struct {
		name, left, right, separator string
		want                         int
	}{
		{"observed change", first, plain, "\n* Document the parser boundaries, #652 by @dmontagu\n", 1},
		{"same formatting", first, first, "\n", 1},
		{"different reference", first, strings.Replace(plain, "651", "652", 1), "\n", 0},
		{"different identifier", first, strings.Replace(plain, "Literal", "NewType", 1), "\n", 0},
		{"identifier case", first, strings.Replace(plain, "Literal", "literal", 1), "\n", 0},
		{"negation", first, strings.Replace(plain, "add ", "do not add ", 1), "\n", 0},
		{"separate releases", first, plain, "\n\n## v2\n\n", 0},
		{"separate lists", first, plain, "\n\nThe next task has its own prerequisites.\n\n", 0},
		{"code command", strings.Replace(first, "Literal", "rm -rf /", 1), strings.Replace(first, "Literal", "rm -rf /", 1), "\n", 0},
		{"URL", "read the documentation at https://example.com/guide", "read the documentation at https://example.com/guide", "\n", 0},
		{"different linked targets", "read the [documentation](https://example.com/a) for this type",
			"read the [documentation](https://example.com/b) for this type", "\n", 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			text := "\ufeff## v1\r\n\r\n* " + row.left + row.separator + "* " + row.right + "\r\n"
			result := singleRuleResult(t, "repetition.duplicate-list-item", text, "", "")
			c.Assert(result.Findings, qt.HasLen, row.want)
			if row.want > 0 {
				f := result.Findings[0]
				c.Assert(f.Primary.Snippet, qt.Equals, row.left)
				c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(text, row.left))
				c.Assert(f.Related, qt.HasLen, 1)
				c.Assert(f.Related[0].Snippet, qt.Equals, row.right)
				c.Assert(f.Related[0].Span.Start, qt.Equals, strings.LastIndex(text, row.right))
				c.Assert(f.Evidence.Metrics[0].Value, qt.Equals, float64(2))
			}
		})
	}
}

func TestDuplicateListApplicabilityAndBudget(t *testing.T) {
	c := qt.New(t)
	const item = "Add documentation for Literal type, #651."
	for _, text := range []string{
		"1. " + item + "\n2. " + item,
		"- [ ] " + item + "\n- [ ] " + item,
		"- " + item + "\n  - Nested context.\n- " + item,
	} {
		c.Assert(singleRuleResult(t, "repetition.duplicate-list-item", text, "", "").Findings, qt.HasLen, 0)
	}
	text := "- " + item + "\n- " + item
	c.Assert(singleRuleResult(t, "repetition.duplicate-list-item", text, "{min_words: 20}", "").Findings, qt.HasLen, 0)
	engine := singleRuleEngine(t, "repetition.duplicate-list-item", "", "analysis: {max_candidates: 1}\n")
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	assertBudgetAbstention(t, result, "guide.md", "repetition.duplicate-list-item", "max_candidates")
}
