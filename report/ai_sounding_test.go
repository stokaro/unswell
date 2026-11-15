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
	plain := variedProse
	one := textReport(c, t, markdownPage("guide.md", "# Client\n\n"+plain+"\n"))
	c.Assert(one, qt.Contains, "AI-sounding index: 23 of 100 (low; marks 23, rhythm 0) in guide.md: none of its marks.\n")
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
	c.Assert(many, qt.Matches, `(?s).*AI-sounding index: \d+ of 100 \(high; marks \d+, rhythm \d+\) in b\.md: 12 em dashes\.\n.*`)
	c.Assert(many, qt.Matches, `(?s).*AI-sounding index: \d+ of 100 \(elevated; marks \d+, rhythm \d+\) in c\.md: 1 em dash\.\n.*`)
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
	plain := variedProse
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
		"The index is the higher of two readings")
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
	c.Assert(low, qt.Matches, `(?s).*\(low; marks \d+, rhythm \d+\) in guide\.md: 1 em dash\.\n.*`)
	c.Assert(low, qt.Not(qt.Contains), ": em dash;")
	c.Assert(low, qt.Not(qt.Contains), "unswell fix --write")

	// The marks listed across a run are bounded, and the rest are counted.
	dashes := strings.Repeat("It connects — then it waits. ", 70)
	long := textReport(c, t, markdownPage("a.md", "# Client\n\n"+dashes+"\n"), markdownPage("b.md", "# Client\n\n"+dashes+"\n"))
	c.Assert(strings.Count(long, ": em dash; write a comma\n"), qt.Equals, 60)
	c.Assert(long, qt.Contains, "  10 more marks in a.md are not listed here.\n")
	c.Assert(long, qt.Contains, "  70 more marks in b.md are not listed here.\n")
}

func TestTextReportListsTheRunsOfSentences(t *testing.T) {
	c := qt.New(t)
	even := strings.Repeat("The server reads the request and sends the response to the client. ", 16)
	text := textReport(c, t, markdownPage("guide.md", "# Client\n\n"+even+"\n"))
	// A page without any of the marks is read by the rhythm of its sentences, and each run is listed with its remedy.
	c.Assert(text, qt.Matches, `(?s).*AI-sounding index: (\d+) of 100 \(high; marks 23, rhythm (\d+)\) in guide\.md: `+
		`1 run of plain sentences, 1 run of sentences that open with their subject\.\n.*`)
	c.Assert(text, qt.Contains, "\n  guide.md:3:1: 16 plain sentences in a row; say it as you would to a colleague: join two of them, "+
		"use a contraction, put an aside in parentheses, or start one with And or But\n"+
		"  guide.md:3:1: 16 sentences in a row open with their subject; open one of them with its condition, its time, or its reason\n"+
		"A run of sentences has no replacement that a command can make. Rewrite one or two of its sentences, and keep what they say.\n"+
		"The index is the higher of two readings")
	c.Assert(text, qt.Not(qt.Contains), "unswell fix --write")

	// A page that is high by its marks has two plain sentences in a row too. Its rhythm is low, so the run is not
	// named and not listed.
	plain := variedProse
	marked := textReport(c, t, markdownPage("guide.md", "# Client\n\nIt connects — then it waits — and it closes — on a timeout.\n\n"+
		"The server reads the request and sends the response. The client reads the response and closes the connection.\n\n"+plain+"\n"))
	c.Assert(marked, qt.Matches, `(?s).*in guide\.md: 3 em dashes\.\n.*`)
	c.Assert(marked, qt.Not(qt.Contains), "in a row")
	c.Assert(marked, qt.Not(qt.Contains), "A run of sentences")
	c.Assert(marked, qt.Contains, "`unswell fix --write` makes the replacements named above.")

	// Both readings above low: the marks of both are listed in the order of the page.
	both := textReport(c, t, markdownPage("guide.md", "# Client\n\nIt connects — then it waits — and it closes — on a timeout.\n\n"+even+"\n"))
	c.Assert(both, qt.Contains, "  guide.md:3:13: em dash; write a comma\n")
	c.Assert(both, qt.Contains, "  guide.md:5:1: 16 plain sentences in a row; ")
	c.Assert(both, qt.Contains, "`unswell fix --write` makes the replacements named above. A stock word needs your own words.\n"+
		"A run of sentences has no replacement that a command can make.")
}

// variedProse is ordinary documentation without any mark that the index
// counts: its sentences differ in length and in how they start.
var variedProse = "When a request arrives, the server checks the token before anything else, because an expired token is the most common " +
	"reason for a rejection and the check costs almost nothing. Why first? Everything after it needs a read from the database. " +
	"If the token is valid, the handler loads the account, compares its plan with the limits of the endpoint, and only then " +
	"reads the body, which can be several megabytes for an upload. Small bodies stay in memory. A larger one goes to a " +
	"temporary file (the handler removes it once the response is sent), so that a slow client cannot hold the memory of a " +
	"worker for minutes. Does the order matter? It does: on a busy day about one request in fifty fails the first check, and " +
	"none of those reach the database or the disk. After the body is read, the handler validates each field, writes the row " +
	"inside one transaction, and answers with the identifier of the new record together with the time the server stored it. "
