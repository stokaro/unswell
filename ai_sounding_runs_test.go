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

// even is one plain statement of twelve words that opens with its subject.
const even = "The server reads the request and sends the response to the client."

func runs(index *unswell.AISounding, signal string) []int {
	var lengths []int
	for _, mark := range index.Marks {
		if mark.Signal == signal {
			lengths = append(lengths, mark.Sentences)
		}
	}
	return lengths
}

func TestAISoundingIndexReadsTheRhythmOfSentences(t *testing.T) {
	c := qt.New(t)
	page := "# Client\n\n" + evenProse + "\n"
	index := aiSounding(c, t, "guide.md", document.Markdown, page)
	c.Assert(index.Status, qt.Equals, "available")
	// The page has none of the marks, and its index comes from the second reading.
	c.Assert(index.Readings.Marks, qt.Equals, 23)
	c.Assert(index.Readings.Rhythm >= 70, qt.IsTrue, qt.Commentf("%d", index.Readings.Rhythm))
	c.Assert(index.Index, qt.Equals, index.Readings.Rhythm)
	c.Assert(index.Level, qt.Equals, unswell.AISoundingHigh)
	c.Assert(counts(index), qt.DeepEquals, map[string]int{"em-dash": 0, "emoji-in-heading": 0, "emoji-at-list-item": 0,
		"arrow": 0, "stock-word": 0, "bold-label-in-list": 0})
	for _, signal := range index.Signals[6:] {
		c.Assert(signal.Reading, qt.Equals, unswell.AIReadingRhythm)
		c.Assert(signal.Count, qt.Equals, 1, qt.Commentf(signal.Name))
		c.Assert(signal.Sentences, qt.Equals, 16, qt.Commentf(signal.Name))
		c.Assert(signal.PerThousand, qt.Equals, 1000*16/float64(index.Words))
		c.Assert(signal.Advice, qt.Not(qt.Equals), "")
	}
	// One mark for each run: it covers the run from its first sentence to its last and has no replacement.
	c.Assert(index.Marks, qt.HasLen, 2)
	for _, mark := range index.Marks {
		c.Assert(mark.Sentences, qt.Equals, 16)
		c.Assert(mark.Fix, qt.IsNil)
		c.Assert(page[mark.Span.Start:mark.Span.End], qt.Equals, strings.TrimSpace(evenProse))
		c.Assert(mark.Start, qt.Equals, document.Position{Line: 3, Column: 1})
	}
}

func TestAISoundingRunsEndWhereTheProseEnds(t *testing.T) {
	c := qt.New(t)
	// A code block, a heading, and a list each end a run; a blank line between two paragraphs does not.
	page := "# Client\n\n" + even + " " + even + "\n\n" + even + "\n\n```text\ncode\n```\n\n" + even + " " + even + "\n\n## Next\n\n" +
		even + " " + even + "\n\n- " + even + "\n- " + even + "\n\n" + even + "\n\n" + plainProse + "\n"
	index := aiSounding(c, t, "guide.md", document.Markdown, page)
	c.Assert(runs(index, "plain-sentence-run"), qt.DeepEquals, []int{3, 2, 2})
	// Three in a row open with their subject before the code block; the other stretches are shorter.
	c.Assert(runs(index, "subject-opener-run"), qt.DeepEquals, []int{3})
}

func TestAISoundingRunsCountOnlyPlainSentencesOfMiddlingLength(t *testing.T) {
	long := "The server reads the request, checks the token of the caller against the list of the keys that the operator has " +
		"issued, and sends the response to the client."
	tests := map[string]struct {
		middle string
		plain  []int
	}{
		"a plain sentence keeps the run":    {even, []int{5}},
		"a question ends the run":           {"Does the server read the request and send the response to the client?", []int{2, 2}},
		"an exclamation ends the run":       {"The server reads the request and sends the response to the client!", []int{2, 2}},
		"a parenthesis ends the run":        {"The server reads the request (and the headers) and sends the response.", []int{2, 2}},
		"a colon ends the run":              {"The server reads two parts of the request: the headers and the body.", []int{2, 2}},
		"a semicolon ends the run":          {"The server reads the request; the client waits for the response.", []int{2, 2}},
		"an em dash ends the run":           {"The server reads the request — the client waits for the response.", []int{2, 2}},
		"a sentence of four words ends it":  {"The server then waits.", []int{2, 2}},
		"a sentence of five words keeps it": {"The server waits for requests.", []int{5}},
		"a sentence over twenty words ends": {long, []int{2, 2}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			page := "# Client\n\n" + even + " " + even + " " + test.middle + " " + even + " " + even + "\n\n" + plainProse + "\n"
			c.Assert(runs(aiSounding(c, t, "guide.md", document.Markdown, page), "plain-sentence-run"), qt.DeepEquals, test.plain)
		})
	}
}

func TestAISoundingRunsOfSubjectOpenersNeedThreeSentences(t *testing.T) {
	c := qt.New(t)
	other := "When the request arrives, the server reads it and sends the response to the client."
	page := func(sentences ...string) string {
		return "# Client\n\n" + strings.Join(sentences, " ") + "\n\n" + plainProse + "\n"
	}
	two := aiSounding(c, t, "guide.md", document.Markdown, page(even, "It keeps the connection open for the next request.", other))
	c.Assert(runs(two, "subject-opener-run"), qt.HasLen, 0)
	three := aiSounding(c, t, "guide.md", document.Markdown,
		page(even, "It keeps the connection open for the next request.", "You can close it with a call to Stop.", other))
	c.Assert(runs(three, "subject-opener-run"), qt.DeepEquals, []int{3})
	// A sentence that opens with its condition ends the run.
	broken := aiSounding(c, t, "guide.md", document.Markdown,
		page(even, "It keeps the connection open for the next request.", other, "You can close it with a call to Stop."))
	c.Assert(runs(broken, "subject-opener-run"), qt.HasLen, 0)
}

func TestAISoundingListsTheMarksOfAReadingAboveLow(t *testing.T) {
	c := qt.New(t)
	// Marks high, rhythm low: the dashes are listed, the two plain sentences in a row are not.
	dashes := "# Client\n\nIt connects — then it waits — and it closes — on a timeout.\n\n" + even + " " + even + "\n\n" + plainProse + "\n"
	index := aiSounding(c, t, "guide.md", document.Markdown, dashes)
	c.Assert(unswell.AILevel(index.Readings.Marks), qt.Not(qt.Equals), unswell.AISoundingLow)
	c.Assert(unswell.AILevel(index.Readings.Rhythm), qt.Equals, unswell.AISoundingLow)
	c.Assert(runs(index, "plain-sentence-run"), qt.DeepEquals, []int{2})
	listed, total := index.Listed()
	c.Assert(total, qt.Equals, 3)
	for _, mark := range listed {
		c.Assert(mark.Signal, qt.Equals, "em-dash")
	}
	c.Assert(listed, qt.HasLen, 3)

	// Rhythm high, marks low: the runs are listed, the one arrow of a long page is not.
	rhythm := "# Client\n\n" + strings.Repeat(evenProse, 5) + "\n\nThe request goes to the proxy → the server.\n"
	index = aiSounding(c, t, "guide.md", document.Markdown, rhythm)
	c.Assert(unswell.AILevel(index.Readings.Marks), qt.Equals, unswell.AISoundingLow)
	c.Assert(index.Level, qt.Equals, unswell.AISoundingHigh)
	listed, total = index.Listed()
	c.Assert(total, qt.Equals, len(listed))
	c.Assert(len(listed) > 0, qt.IsTrue)
	for _, mark := range listed {
		c.Assert(mark.Sentences > 0, qt.IsTrue, qt.Commentf(mark.Signal))
	}

	// Both low: nothing is listed, though the marks are counted.
	quiet := aiSounding(c, t, "guide.md", document.Markdown, "# Client\n\n"+even+" "+even+"\n\n"+plainProse+plainProse+"\n")
	c.Assert(quiet.Level, qt.Equals, unswell.AISoundingLow)
	c.Assert(quiet.Marks, qt.HasLen, 1)
	listed, total = quiet.Listed()
	c.Assert(listed, qt.HasLen, 0)
	c.Assert(total, qt.Equals, 0)
}

func TestAILevelNamesTheLevelsOfAnIndex(t *testing.T) {
	c := qt.New(t)
	c.Assert(unswell.AILevel(0), qt.Equals, unswell.AISoundingLow)
	c.Assert(unswell.AILevel(39), qt.Equals, unswell.AISoundingLow)
	c.Assert(unswell.AILevel(40), qt.Equals, unswell.AISoundingElevated)
	c.Assert(unswell.AILevel(69), qt.Equals, unswell.AISoundingElevated)
	c.Assert(unswell.AILevel(70), qt.Equals, unswell.AISoundingHigh)
	c.Assert(unswell.AISounding{}.Reading(unswell.AIReadingRhythm), qt.Equals, 0)
}

// The record of the second reading recounts its numbers with the weights it
// publishes. This test fails when the product and the record part.
func TestAISoundingRhythmUsesTheWeightsOfItsRecord(t *testing.T) {
	c := qt.New(t)
	data, err := os.ReadFile("research/reviews/2026-11-15-style-guide-text/results.json")
	c.Assert(err, qt.IsNil)
	type reading struct {
		Weights map[string]float64 `json:"weights"`
		Bias    float64            `json:"bias"`
	}
	var record struct {
		Readings struct {
			Marks      reading `json:"marks"`
			Rhythm     reading `json:"rhythm"`
			ElevatedAt int     `json:"elevated_at"`
			HighAt     int     `json:"high_at"`
		} `json:"readings"`
	}
	c.Assert(json.Unmarshal(data, &record), qt.IsNil)
	c.Assert(record.Readings.Rhythm.Weights, qt.HasLen, 2)
	c.Assert(record.Readings.Marks.Weights, qt.HasLen, 6)
	c.Assert(unswell.AILevel(record.Readings.HighAt), qt.Equals, unswell.AISoundingHigh)
	c.Assert(unswell.AILevel(record.Readings.HighAt-1), qt.Equals, unswell.AISoundingElevated)
	c.Assert(unswell.AILevel(record.Readings.ElevatedAt), qt.Equals, unswell.AISoundingElevated)
	c.Assert(unswell.AILevel(record.Readings.ElevatedAt-1), qt.Equals, unswell.AISoundingLow)

	conditional := "When the request arrives, the server sends the response to the client."
	pages := map[string]string{
		// Plain statements that open with a condition: the first signal alone.
		"plain runs": strings.Repeat(conditional+" ", 24),
		// Plain statements that open with their subject: both signals.
		"both runs": strings.Repeat(even+" ", 16),
	}
	for name, run := range pages {
		index := aiSounding(c, t, "guide.md", document.Markdown, "# Client\n\n"+run+"\n\n"+plainProse+"\n")
		score := record.Readings.Rhythm.Bias
		signals := 0
		for _, signal := range index.Signals {
			if signal.Reading != unswell.AIReadingRhythm {
				continue
			}
			weight, known := record.Readings.Rhythm.Weights[signal.Name]
			c.Assert(known, qt.IsTrue, qt.Commentf(signal.Name))
			score += weight * math.Log1p(signal.PerThousand)
			signals++
		}
		c.Assert(signals, qt.Equals, 2)
		expected := int(math.Round(100 / (1 + math.Exp(-score))))
		c.Assert(index.Readings.Rhythm, qt.Equals, expected, qt.Commentf(name))
		c.Assert(index.Index, qt.Equals, expected, qt.Commentf(name))
		// The page sits in the middle of the scale, where a changed weight or constant moves the reading.
		c.Assert(expected > 40 && expected < 90, qt.IsTrue, qt.Commentf("%s: %d", name, expected))
	}
	plain := aiSounding(c, t, "guide.md", document.Markdown, "# Client\n\n"+pages["plain runs"]+"\n\n"+plainProse+"\n")
	c.Assert(runs(plain, "subject-opener-run"), qt.HasLen, 0)
	c.Assert(runs(plain, "plain-sentence-run"), qt.DeepEquals, []int{24})
}
