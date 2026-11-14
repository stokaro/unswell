package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

// fixed analyzes a Markdown page, applies the fixes of its marks, and returns
// the new page with the number of fixes applied.
func fixed(c *qt.C, t *testing.T, page string) (string, int) {
	c.Helper()
	index := aiSounding(c, t, "guide.md", document.Markdown, page)
	c.Assert(index, qt.IsNotNil)
	c.Assert(index.Status, qt.Equals, "available")
	revised, applied := unswell.RemoveAIMarks([]byte(page), index.Marks)
	return string(revised), applied
}

func TestRemoveAIMarksReplacesEachKindOfMark(t *testing.T) {
	cases := []struct{ name, before, after string }{
		{"a dash between clauses becomes a comma", "It connects — then it waits.", "It connects, then it waits."},
		{"a dash without spaces becomes a comma", "It connects—then it waits.", "It connects, then it waits."},
		{"a dash at the end of a line", "It connects —\nthen it waits.", "It connects,\nthen it waits."},
		{"a dash after other punctuation is dropped", "It connects (once) — then it waits. Done. — Then it stops.",
			"It connects (once), then it waits. Done. Then it stops."},
		{"a dash at the start of a list item is dropped", "- — it waits.\n- It stops.", "- it waits.\n- It stops."},
		{"a dash after a short label becomes a colon", "- Timeout — the client waits this long.\n- `--retry` — the count of retries.",
			"- Timeout: the client waits this long.\n- `--retry`: the count of retries."},
		{"a dash after a longer start of an item stays a comma", "- The client waits for the server to answer — then it stops.",
			"- The client waits for the server to answer, then it stops."},
		{"a bold label loses its asterisks and its dash", "- **Timeout** — the client waits.\n- **Retry:** it retries.",
			"- Timeout: the client waits.\n- Retry: it retries."},
		{"a bold label that ends with a colon keeps one", "- **Timeout:** — the client waits.", "- Timeout: the client waits."},
		{"an emoji leaves a heading", "## \U0001F680 Start\n\nIt starts.\n\n## Stop \U0001F6D1\n\nIt stops.",
			"## Start\n\nIt starts.\n\n## Stop\n\nIt stops."},
		{"an emoji with a variation selector", "## ⚙️ Settings\n\nIt has settings.", "## Settings\n\nIt has settings."},
		{"an emoji leaves the start of a list item", "- ⚡ It keeps one connection.\n- [x] \U0001F680 It starts.",
			"- It keeps one connection.\n- [x] It starts."},
		{"an arrow becomes two characters", "The request goes client → proxy ⇒ server.", "The request goes client -> proxy => server."},
		{"a dash in a heading", "## Start — the first run\n\nIt starts.", "## Start, the first run\n\nIt starts."},
		{"a dash between two full clauses becomes a period",
			"The skill stays small — the build enforces it. All calls are mocked — no command runs.",
			"The skill stays small. The build enforces it. All calls are mocked. No command runs."},
		{"a period before a clause on the next line", "The server never compiles its own code —\nit takes the build of the client.",
			"The server never compiles its own code.\nIt takes the build of the client."},
		{"a clause that starts with markup keeps its case", "The client never retries — `retry` is off by default.",
			"The client never retries. `retry` is off by default."},
		{"a dependent phrase keeps the comma", "The client retries — which the server allows. It waits — a pause that the server sets.",
			"The client retries, which the server allows. It waits, a pause that the server sets."},
		{"two dashes around an aside stay commas", "The client — it is small — retries.", "The client, it is small, retries."},
		{"a dependent clause before the dash keeps the comma", "Although the loop is untouched — it is a new build.",
			"Although the loop is untouched, it is a new build."},
		{"an open parenthesis keeps the comma", "Unzip the folder (do not delete it — the browser needs it) and start.",
			"Unzip the folder (do not delete it, the browser needs it) and start."},
		{"a participle after the dash keeps the comma", "It is one change — a list audited end to end.",
			"It is one change, a list audited end to end."},
		{"a noun with a relative clause keeps the comma", "It is the main tool — the thing you use when the build fails.",
			"It is the main tool, the thing you use when the build fails."},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			page := "# Client\n\n" + test.before + "\n\n" + plainProse + "\n"
			revised, applied := fixed(c, t, page)
			c.Assert(revised, qt.Equals, "# Client\n\n"+test.after+"\n\n"+plainProse+"\n")
			c.Assert(applied > 0, qt.IsTrue)
			// The result has nothing left to replace.
			again, more := fixed(c, t, revised)
			c.Assert(more, qt.Equals, 0)
			c.Assert(again, qt.Equals, revised)
		})
	}
}

func TestRemoveAIMarksLeavesWhatNeedsAPerson(t *testing.T) {
	c := qt.New(t)
	page := "# Client\n\nA seamless client for pages 10—20 of the guide.\n\n```text\na — b → c\n```\n\nRun `a → b` first.\n\n" +
		plainProse + "\n"
	index := aiSounding(c, t, "guide.md", document.Markdown, page)
	c.Assert(counts(index), qt.DeepEquals, map[string]int{"em-dash": 1, "emoji-in-heading": 0, "emoji-at-list-item": 0,
		"arrow": 0, "stock-word": 1, "bold-label-in-list": 0})
	// A stock word and a dash between two numbers have no mechanical replacement, and code is not read at all.
	for _, mark := range index.Marks {
		c.Assert(mark.Fix, qt.IsNil, qt.Commentf(mark.Signal))
	}
	revised, applied := unswell.RemoveAIMarks([]byte(page), index.Marks)
	c.Assert(applied, qt.Equals, 0)
	c.Assert(string(revised), qt.Equals, page)
}

func TestRemoveAIMarksLowersTheIndex(t *testing.T) {
	c := qt.New(t)
	page := "# \U0001F680 Client\n\nIt connects — then it waits — and it retries → twice.\n\n" +
		"- **Fast**: it starts quickly.\n- ⚡ One connection.\n\n" + plainProse + "\n"
	before := aiSounding(c, t, "guide.md", document.Markdown, page)
	revised, applied := unswell.RemoveAIMarks([]byte(page), before.Marks)
	c.Assert(applied, qt.Equals, 6)
	after := aiSounding(c, t, "guide.md", document.Markdown, string(revised))
	c.Assert(before.Level, qt.Equals, unswell.AISoundingHigh)
	c.Assert(after.Level, qt.Equals, unswell.AISoundingLow)
	c.Assert(after.Index, qt.Equals, 23)
	c.Assert(after.Marks, qt.HasLen, 0)
	// Only the marks change: the words of the page stay as they were.
	c.Assert(strings.Fields(strings.NewReplacer("—", "", "→", "", "\U0001F680", "", "⚡", "", "**", "", ",", "", "->", "").Replace(page)),
		qt.DeepEquals, strings.Fields(strings.NewReplacer(",", "", "->", "").Replace(string(revised))))
}

func TestRemoveAIMarksSkipsFixesThatDoNotFit(t *testing.T) {
	c := qt.New(t)
	source := []byte("one two three")
	fix := func(start, end int, text string) unswell.AIMark {
		return unswell.AIMark{Fix: &unswell.AIMarkFix{Span: document.Span{Start: start, End: end}, Text: text}}
	}
	// The second fix overlaps the first, the third lies outside the source, and the fourth has no fix.
	revised, applied := unswell.RemoveAIMarks(source, []unswell.AIMark{
		fix(8, 13, "3"), fix(0, 3, "1"), fix(2, 5, "x"), fix(10, 40, "y"), {Signal: "stock-word"},
	})
	c.Assert(string(revised), qt.Equals, "1 two 3")
	c.Assert(applied, qt.Equals, 2)
	c.Assert(string(source), qt.Equals, "one two three")
	same, none := unswell.RemoveAIMarks(source, nil)
	c.Assert(none, qt.Equals, 0)
	c.Assert(string(same), qt.Equals, "one two three")
}
