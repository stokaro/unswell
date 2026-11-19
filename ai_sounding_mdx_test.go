package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

// mdxMarked is an MDX page with marks in three kinds of place: in its prose,
// in the text inside a component, and in the code of the page, which is an
// import, an attribute, and an expression.
var mdxMarked = "import Note from './Note.astro';\n\n# Fast client\n\n" +
	"The client connects — then it waits. It is robust → and it retries.\n\n" +
	"<Note title=\"A dash — in an attribute → here\">\n  The note is powerful — it warns.\n</Note>\n\n" +
	"The total is {items.map((item) => item.name + ' — ' + item.size)} items.\n\n" +
	"- **Fast**: it starts quickly.\n- A plain item.\n\n" + plainProse + "\n"

func TestAISoundingIndexReadsAnMDXPage(t *testing.T) {
	c := qt.New(t)
	index := aiSounding(c, t, "guide.mdx", document.MDX, mdxMarked)
	c.Assert(index, qt.IsNotNil)
	c.Assert(index.Status, qt.Equals, "available")
	// The text inside the component counts as prose. The attribute, the
	// expression, and the import are code and add no mark.
	c.Assert(counts(index), qt.DeepEquals, map[string]int{"em-dash": 2, "emoji-in-heading": 0, "emoji-at-list-item": 0,
		"arrow": 1, "stock-word": 2, "bold-label-in-list": 1})
	var lines []int
	for _, mark := range index.Marks {
		lines = append(lines, mark.Start.Line)
		c.Assert(mark.Start.Line != 7 && mark.Start.Line != 11, qt.IsTrue, qt.Commentf("a mark in the code of line %d", mark.Start.Line))
	}
	c.Assert(lines, qt.DeepEquals, []int{5, 5, 5, 8, 8, 13})
	c.Assert(index.Level, qt.Equals, unswell.AISoundingHigh)

	// The same prose without the marks is low, as it is in a Markdown page.
	plain := aiSounding(c, t, "guide.mdx", document.MDX, "import Note from './Note.astro';\n\n# Client\n\n<Note>\n  A short note.\n</Note>\n\n"+
		plainProse+"\n")
	c.Assert(plain.Index, qt.Equals, 23)
	c.Assert(plain.Level, qt.Equals, unswell.AISoundingLow)

	short := aiSounding(c, t, "guide.mdx", document.MDX, "# Client\n\n<Note>\n  It connects — then it waits.\n</Note>\n")
	c.Assert(short.Status, qt.Equals, "too_short")
}

func TestAISoundingRunsEndAtAComponent(t *testing.T) {
	c := qt.New(t)
	// Two sentences before the component, two inside it, and three after it:
	// seven like sentences, and no run reaches across a tag.
	page := "# Client\n\n" + even + " " + even + "\n\n<Note>\n  " + even + " " + even + "\n</Note>\n\n" +
		even + " " + even + "\n\n" + even + "\n\n" + plainProse + "\n"
	index := aiSounding(c, t, "guide.mdx", document.MDX, page)
	c.Assert(runs(index, "plain-sentence-run"), qt.DeepEquals, []int{2, 2, 3})
	c.Assert(runs(index, "subject-opener-run"), qt.DeepEquals, []int{3})
}

func TestRemoveAIMarksLeavesTheCodeOfAnMDXPage(t *testing.T) {
	c := qt.New(t)
	index := aiSounding(c, t, "guide.mdx", document.MDX, mdxMarked)
	revised, applied := unswell.RemoveAIMarks([]byte(mdxMarked), index.Marks)
	c.Assert(applied, qt.Equals, 4)
	want := strings.NewReplacer(
		"The client connects — then it waits. It is robust → and it retries.",
		"The client connects, then it waits. It is robust -> and it retries.",
		"The note is powerful — it warns.", "The note is powerful. It warns.",
		"- **Fast**: it starts quickly.", "- Fast: it starts quickly.",
	).Replace(mdxMarked)
	c.Assert(string(revised), qt.Equals, want)
	// The attribute and the expression keep their dash and their arrow.
	c.Assert(string(revised), qt.Contains, `<Note title="A dash — in an attribute → here">`)
	c.Assert(string(revised), qt.Contains, `{items.map((item) => item.name + ' — ' + item.size)}`)

	// The page that the fixes leave is still a complete MDX page, and only
	// the stock words are left of its marks.
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX, Bytes: revised})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
	after := result.Documents[0].AISounding
	c.Assert(after.Index < index.Index, qt.IsTrue, qt.Commentf("%d and %d", after.Index, index.Index))
	c.Assert(counts(after), qt.DeepEquals, map[string]int{"em-dash": 0, "emoji-in-heading": 0, "emoji-at-list-item": 0,
		"arrow": 0, "stock-word": 2, "bold-label-in-list": 0})
}
