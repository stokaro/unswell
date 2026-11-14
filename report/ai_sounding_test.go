package report_test

import (
	"bytes"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func textReport(c *qt.C, t *testing.T, sources ...document.Source) string {
	c.Helper()
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	var text bytes.Buffer
	c.Assert(report.Write(&text, "text", result, report.Options{}), qt.IsNil)
	return text.String()
}

func markdownPage(name, text string) document.Source {
	return document.Source{Name: name, Format: document.Markdown, Bytes: []byte(text)}
}

func TestTextReportNamesThePageIndex(t *testing.T) {
	c := qt.New(t)
	plain := strings.Repeat("The server reads the request and sends the response to the client. ", 16)
	one := textReport(c, t, markdownPage("guide.md", "# Client\n\n"+plain+"\n"))
	c.Assert(one, qt.Contains, "AI-sounding index: 23 of 100 (low) in guide.md: none of its marks.\n")
	c.Assert(one, qt.Contains, "It is not a probability and does not say who wrote a page.")
	c.Assert(one, qt.Not(qt.Contains), "Markdown pages:")

	dashes := strings.Repeat("It connects — then it waits. ", 12)
	many := textReport(c, t,
		markdownPage("a.md", "# Client\n\n"+plain+"\n"),
		markdownPage("b.md", "# Client\n\n"+dashes+plain+"\n"),
		markdownPage("c.md", "# Client\n\nIt connects — then it waits. "+plain+"\n"),
		markdownPage("short.md", "# Client\n\nIt connects — then it waits.\n"),
		document.Source{Name: "client.go", Format: document.Go,
			Bytes: []byte("package client\n\n// Start opens the connection.\nfunc Start() {}\n")},
	)
	c.Assert(many, qt.Matches, `(?s).*AI-sounding index: \d+ of 100 \(high\) in b\.md: 12 em dashes\.\n.*`)
	c.Assert(many, qt.Matches, `(?s).*AI-sounding index: \d+ of 100 \(elevated\) in c\.md: 1 em dash\.\n.*`)
	c.Assert(many, qt.Contains, "AI-sounding index over 3 Markdown pages: 1 high, 1 elevated, 1 low.\n")
	// Each page above the low level lists where its marks are.
	c.Assert(many, qt.Contains, "\n  b.md:3:13: em dash; write a comma\n  b.md:3:42: em dash; write a comma\n")
	c.Assert(strings.Count(many, "\n  b.md:3:"), qt.Equals, 12)
	c.Assert(many, qt.Contains, "\n  c.md:3:13: em dash; write a comma\n")
	c.Assert(strings.Count(many, "`unswell fix --write` makes the replacements named above."), qt.Equals, 1)
	// The most marked page comes first, and a low page is not listed among several.
	c.Assert(strings.Index(many, "in b.md") < strings.Index(many, "in c.md"), qt.IsTrue)
	c.Assert(many, qt.Not(qt.Contains), "in a.md")
	c.Assert(many, qt.Not(qt.Contains), "in short.md")

	source := textReport(c, t, document.Source{Name: "client.go", Format: document.Go,
		Bytes: []byte("package client\n\n// Start opens the connection and waits for the first answer of the server.\nfunc Start() {}\n")})
	c.Assert(source, qt.Not(qt.Contains), "AI-sounding")
}

func TestTextReportListsTheMarksOfAPage(t *testing.T) {
	c := qt.New(t)
	plain := strings.Repeat("The server reads the request and sends the response to the client. ", 16)
	page := "# \U0001F680 Client\n\nA Seamless client — it retries → then it waits.\n\n" +
		"- **Fast**: it starts quickly.\n- ⚡ One connection.\n\n" + plain + "\n"
	text := textReport(c, t, markdownPage("guide.md", page))
	// Each mark is listed with the replacement that `unswell fix` would make, or with what only a person can do.
	c.Assert(text, qt.Contains, "\n  guide.md:1:3: emoji in a heading; remove it\n"+
		"  guide.md:3:3: stock word \"seamless\"; say what the thing does, or drop the word\n"+
		"  guide.md:3:19: em dash; write a comma\n"+
		"  guide.md:3:32: arrow; write \"->\"\n"+
		"  guide.md:5:3: list item that starts with a bold label; remove the bold\n"+
		"  guide.md:6:3: list item that starts with an emoji; remove it\n"+
		"`unswell fix --write` makes the replacements named above. A stock word needs your own words.\n"+
		"The index counts marks")
	c.Assert(text, qt.Not(qt.Contains), "not listed here")

	// The replacement follows the place of the dash, and a dash between two numbers has none.
	placed := textReport(c, t, markdownPage("guide.md", "# Client\n\nThe skill stays small — the build enforces it.\n\n"+
		"- Timeout — the client waits.\n\nSee pages 10—20 (once) — then stop.\n\n"+plain+"\n"))
	c.Assert(placed, qt.Contains, "\n  guide.md:3:23: em dash; write a period and start the next word with a capital letter\n"+
		"  guide.md:5:11: em dash; write a colon\n"+
		"  guide.md:7:13: em dash; leave it if it stands between two numbers\n"+
		"  guide.md:7:24: em dash; write a comma\n")

	// A low page has little to fix, so its marks are not listed.
	low := textReport(c, t, markdownPage("guide.md", "# Client\n\nIt connects — then it waits. "+strings.Repeat(plain, 8)+"\n"))
	c.Assert(low, qt.Contains, "(low) in guide.md: 1 em dash.\n")
	c.Assert(low, qt.Not(qt.Contains), ": em dash;")
	c.Assert(low, qt.Not(qt.Contains), "unswell fix --write")

	// The marks listed across a run are bounded, and the rest are counted.
	dashes := strings.Repeat("It connects — then it waits. ", 70)
	long := textReport(c, t, markdownPage("a.md", "# Client\n\n"+dashes+"\n"), markdownPage("b.md", "# Client\n\n"+dashes+"\n"))
	c.Assert(strings.Count(long, ": em dash; write a comma\n"), qt.Equals, 60)
	c.Assert(long, qt.Contains, "  10 more marks in a.md are not listed here.\n")
	c.Assert(long, qt.Contains, "  70 more marks in b.md are not listed here.\n")
}
