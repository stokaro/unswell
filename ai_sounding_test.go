package unswell_test

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

// plainProse is ordinary documentation without any mark that the index counts:
// its sentences differ in length and in how they start, as a person's do.
var plainProse = "When a request arrives, the server checks the token before anything else, because an expired token is the most common " +
	"reason for a rejection and the check costs almost nothing. Why first? Everything after it needs a read from the database. " +
	"If the token is valid, the handler loads the account, compares its plan with the limits of the endpoint, and only then " +
	"reads the body, which can be several megabytes for an upload. Small bodies stay in memory. A larger one goes to a " +
	"temporary file (the handler removes it once the response is sent), so that a slow client cannot hold the memory of a " +
	"worker for minutes. Does the order matter? It does: on a busy day about one request in fifty fails the first check, and " +
	"none of those reach the database or the disk. After the body is read, the handler validates each field, writes the row " +
	"inside one transaction, and answers with the identifier of the new record together with the time the server stored it. "

// evenProse is the page that the second reading is for: plain statements of
// the same length, one after the other, each opening with its subject.
var evenProse = strings.Repeat("The server reads the request and sends the response to the client. ", 16)

func aiSounding(c *qt.C, t *testing.T, name string, format document.Format, text string) *unswell.AISounding {
	c.Helper()
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: name, Format: format, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	return result.Documents[0].AISounding
}

// counts returns the count of each signal of the first reading, the marks.
func counts(index *unswell.AISounding) map[string]int {
	return countsOf(index, unswell.AIReadingMarks)
}

func countsOf(index *unswell.AISounding, reading string) map[string]int {
	result := map[string]int{}
	for _, signal := range index.Signals {
		if signal.Reading == reading {
			result[signal.Name] = signal.Count
		}
	}
	return result
}

func TestAISoundingIndexCountsItsMarks(t *testing.T) {
	c := qt.New(t)
	marked := "# \U0001F680 Fast client\n\n" +
		"The client is a powerful tool — it connects — then it waits. It is robust → and it retries.\n\n" +
		"- **Fast**: it starts quickly.\n- ⚡ It keeps one connection.\n- [ ] **Later**: more options.\n- A plain item.\n\n" +
		"```text\na — b → c \U0001F680 **bold**\n```\n\n" + plainProse + "\n"
	index := aiSounding(c, t, "guide.md", document.Markdown, marked)
	c.Assert(index, qt.IsNotNil)
	c.Assert(index.Status, qt.Equals, "available")
	c.Assert(counts(index), qt.DeepEquals, map[string]int{"em-dash": 2, "emoji-in-heading": 1, "emoji-at-list-item": 1,
		"arrow": 1, "stock-word": 2, "bold-label-in-list": 2})
	c.Assert(index.Level, qt.Equals, unswell.AISoundingHigh)
	c.Assert(index.Index >= 70 && index.Index <= 100, qt.IsTrue)
	for _, signal := range index.Signals {
		if signal.Reading == unswell.AIReadingMarks {
			c.Assert(signal.PerThousand, qt.Equals, 1000*float64(signal.Count)/float64(index.Words))
		}
	}
	c.Assert(index.Index, qt.Equals, index.Readings.Marks)
	c.Assert(countsOf(index, unswell.AIReadingRhythm), qt.DeepEquals, map[string]int{"plain-sentence-run": 0, "subject-opener-run": 0})
	// Every counted mark is listed in source order with the bytes it covers.
	var listed []string
	for _, mark := range index.Marks {
		listed = append(listed, mark.Signal+" "+marked[mark.Span.Start:mark.Span.End])
		position, err := document.Locate([]byte(marked), mark.Span.Start)
		c.Assert(err, qt.IsNil)
		c.Assert(mark.Start, qt.Equals, position)
	}
	c.Assert(listed, qt.DeepEquals, []string{"emoji-in-heading \U0001F680", "stock-word powerful", "em-dash —", "em-dash —",
		"stock-word robust", "arrow →", "bold-label-in-list **Fast**", "emoji-at-list-item ⚡", "bold-label-in-list **Later**"})
	c.Assert(index.Marks[1].Word, qt.Equals, "powerful")
	c.Assert(index.Marks[2].Word, qt.Equals, "")
}

func TestAISoundingIndexBoundsTheListedMarks(t *testing.T) {
	c := qt.New(t)
	index := aiSounding(c, t, "guide.md", document.Markdown, "# Client\n\n"+strings.Repeat("It connects — then it waits. ", 260)+"\n")
	c.Assert(counts(index)["em-dash"], qt.Equals, 260)
	c.Assert(index.Marks, qt.HasLen, 200)
}

func TestAISoundingIndexIsLowOnPlainProse(t *testing.T) {
	c := qt.New(t)
	index := aiSounding(c, t, "guide.md", document.Markdown, "# Client\n\n"+plainProse+"\n\n- A plain item.\n- Another item.\n")
	c.Assert(index.Status, qt.Equals, "available")
	c.Assert(index.Level, qt.Equals, unswell.AISoundingLow)
	c.Assert(index.Index, qt.Equals, 23)
	for _, signal := range index.Signals {
		c.Assert(signal.Count, qt.Equals, 0)
	}
}

func TestAISoundingIndexFallsWhenMarksAreRemoved(t *testing.T) {
	c := qt.New(t)
	before := "# Client\n\nIt connects — then it waits — and it closes — on a timeout.\n\n" + plainProse + "\n"
	after := strings.ReplaceAll(before, " —", ",")
	first := aiSounding(c, t, "guide.md", document.Markdown, before)
	second := aiSounding(c, t, "guide.md", document.Markdown, after)
	c.Assert(first.Index > second.Index, qt.IsTrue, qt.Commentf("%d and %d", first.Index, second.Index))
	c.Assert(second.Index, qt.Equals, 23)
}

func TestAISoundingIndexIsNotGivenWhereItWasNotMeasured(t *testing.T) {
	c := qt.New(t)
	short := aiSounding(c, t, "guide.md", document.Markdown, "# Client\n\nIt connects — then it waits.\n")
	c.Assert(short.Status, qt.Equals, "too_short")
	c.Assert(short.Level, qt.Equals, "")
	c.Assert(short.Signals, qt.HasLen, 0)
	source := "package client\n\n// Start opens the connection — then it waits.\nfunc Start() {}\n"
	c.Assert(aiSounding(c, t, "client.go", document.Go, source), qt.IsNil)
	c.Assert(aiSounding(c, t, "notes.txt", document.Plain, plainProse+plainProse), qt.IsNil)
}

// The record of the index recounts its reading with the weights it publishes.
// This test fails when the product and the record part.
func TestAISoundingIndexUsesTheWeightsOfItsRecord(t *testing.T) {
	c := qt.New(t)
	data, err := os.ReadFile("research/reviews/2026-11-12-human-reference/results.json")
	c.Assert(err, qt.IsNil)
	var record struct {
		PageIndex struct {
			Weights map[string]float64 `json:"weights"`
			Bias    float64            `json:"bias"`
		} `json:"page_index"`
	}
	c.Assert(json.Unmarshal(data, &record), qt.IsNil)
	c.Assert(record.PageIndex.Weights, qt.HasLen, 6)
	pages := map[string]string{
		"em-dash":            "It connects — then it waits. It sends — then it reads. It closes — then it stops.",
		"emoji-in-heading":   "## \U0001F680 Start\n\nIt starts.\n\n## \U0001F680 Speed\n\nIt runs.\n\n## \U0001F680 Stop\n\nIt stops.",
		"emoji-at-list-item": "- \U0001F680 It starts.\n- \U0001F680 It runs.\n- \U0001F680 It stops.",
		"arrow":              "The request goes from the client → the proxy → the server → the store.",
		"stock-word":         "It is a seamless client with a robust parser and a powerful cache.",
		"bold-label-in-list": "- **Start**: it starts.\n- **Run**: it runs.\n- **Stop**: it stops.",
	}
	for name, marked := range pages {
		index := aiSounding(c, t, "guide.md", document.Markdown, "# Client\n\n"+marked+"\n\n"+plainProse+"\n")
		c.Assert(counts(index)[name], qt.Equals, 3, qt.Commentf(name))
		score := record.PageIndex.Bias
		for _, signal := range index.Signals {
			if signal.Reading != unswell.AIReadingMarks {
				continue
			}
			weight, known := record.PageIndex.Weights[signal.Name]
			c.Assert(known, qt.IsTrue, qt.Commentf(signal.Name))
			score += weight * math.Log1p(signal.PerThousand)
		}
		expected := int(math.Round(100 / (1 + math.Exp(-score))))
		c.Assert(index.Readings.Marks, qt.Equals, expected, qt.Commentf(name))
		c.Assert(index.Index, qt.Equals, expected, qt.Commentf(name))
		// The page sits in the middle of the scale, where a changed weight moves the index.
		c.Assert(expected > 40 && expected < 90, qt.IsTrue, qt.Commentf("%s: %d", name, expected))
	}
}
