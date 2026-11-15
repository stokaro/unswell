package report

import (
	"fmt"
	"strings"

	"github.com/stokaro/unswell"
)

// aiSignalNames gives the singular and the plural a reader sees for each
// signal of the page index.
var aiSignalNames = map[string][2]string{
	"em-dash":            {"em dash", "em dashes"},
	"emoji-in-heading":   {"emoji in a heading", "emoji in headings"},
	"emoji-at-list-item": {"list item that starts with an emoji", "list items that start with an emoji"},
	"arrow":              {"arrow", "arrows"},
	"stock-word":         {"stock word", "stock words"},
	"bold-label-in-list": {"list item that starts with a bold label", "list items that start with a bold label"},
	"plain-sentence-run": {"run of plain sentences", "runs of plain sentences"},
	"subject-opener-run": {"run of sentences that open with their subject", "runs of sentences that open with their subject"},
}

// maxAISoundingPages bounds the pages that the text reports name one by one,
// and maxAIMarkLines the marks they list across a run. The JSON report holds
// the marks of every page.
const (
	maxAISoundingPages = 20
	maxAIMarkLines     = 60
)

// aiSoundingLines reports the page index of the Markdown pages of a run: one
// line for each page above the low level, most marked first, with its two
// readings, the place of each listed mark and what to do with it, then the
// count by level and what the index is. A run without a measured page adds
// nothing.
func aiSoundingLines(result unswell.RunResult) []string {
	var measured []unswell.DocumentResult
	levels := map[string]int{}
	for _, doc := range result.Documents {
		if doc.AISounding != nil && doc.AISounding.Status == "available" {
			measured = append(measured, doc)
			levels[doc.AISounding.Level]++
		}
	}
	if len(measured) == 0 {
		return nil
	}
	lines := append(aiSoundingPages(measured), aiAdviceLines(measured)...)
	if len(measured) > 1 {
		lines = append(lines, fmt.Sprintf("AI-sounding index over %d Markdown pages: %d high, %d elevated, %d low.",
			len(measured), levels[unswell.AISoundingHigh], levels[unswell.AISoundingElevated], levels[unswell.AISoundingLow]))
	}
	return append(lines, "The index is the higher of two readings: the marks that people remove when they clean AI-sounding wording "+
		"out of documentation, and runs of sentences that look alike. It is not a probability and does not say who wrote a page.")
}

// aiSoundingPages names the measured pages one by one, the high ones first,
// each followed by the places of its marks. A low page is named only when it
// is the one page of the run.
func aiSoundingPages(measured []unswell.DocumentResult) []string {
	var lines []string
	pages, left := 0, maxAIMarkLines
	for _, level := range []string{unswell.AISoundingHigh, unswell.AISoundingElevated, unswell.AISoundingLow} {
		if level == unswell.AISoundingLow && len(measured) > 1 {
			continue
		}
		for _, doc := range measured {
			if doc.AISounding.Level != level || pages >= maxAISoundingPages {
				continue
			}
			pages++
			lines = append(lines, fmt.Sprintf("AI-sounding index: %d of 100 (%s; marks %d, rhythm %d) in %s%s",
				doc.AISounding.Index, level, doc.AISounding.Reading(unswell.AIReadingMarks),
				doc.AISounding.Reading(unswell.AIReadingRhythm), doc.Name, aiSignalText(*doc.AISounding)))
			var marks []string
			marks, left = aiMarkLines(doc, left)
			lines = append(lines, marks...)
		}
	}
	return lines
}

// aiMarkLines lists the places of the marks of one page, for each of its
// readings that is above the low level, up to the lines left for the run, and
// returns the lines left after it.
func aiMarkLines(doc unswell.DocumentResult, left int) ([]string, int) {
	marks, total := doc.AISounding.Listed()
	if len(marks) == 0 {
		return nil, left
	}
	listed := min(len(marks), left)
	lines := make([]string, 0, listed+1)
	for _, mark := range marks[:listed] {
		lines = append(lines, fmt.Sprintf("  %s:%d:%d: %s; %s",
			doc.Name, mark.Start.Line, mark.Start.Column, aiMarkName(mark), aiMarkAction(mark)))
	}
	switch rest := total - listed; {
	case rest == 1:
		lines = append(lines, fmt.Sprintf("  1 more mark in %s is not listed here.", doc.Name))
	case rest > 1:
		lines = append(lines, fmt.Sprintf("  %d more marks in %s are not listed here.", rest, doc.Name))
	}
	return lines, left - listed
}

// aiMarkName says what one mark is: its signal, with the word of a stock
// word and the length of a run of sentences.
func aiMarkName(mark unswell.AIMark) string {
	switch {
	case mark.Signal == "plain-sentence-run":
		return fmt.Sprintf("%d plain sentences in a row", mark.Sentences)
	case mark.Signal == "subject-opener-run":
		return fmt.Sprintf("%d sentences in a row open with their subject", mark.Sentences)
	case mark.Word != "":
		return fmt.Sprintf("%s %q", aiSignalNames[mark.Signal][0], mark.Word)
	}
	return aiSignalNames[mark.Signal][0]
}

// aiMarkAction says what to do with one mark: the replacement that
// `unswell fix` would make, or what only a person can do.
func aiMarkAction(mark unswell.AIMark) string {
	switch {
	case mark.Signal == "plain-sentence-run":
		return "say it as you would to a colleague: join two of them, use a contraction, put an aside in parentheses, " +
			"or start one with And or But"
	case mark.Signal == "subject-opener-run":
		return "open one of them with its condition, its time, or its reason"
	case mark.Signal == "stock-word":
		return "say what the thing does, or drop the word"
	case mark.Fix == nil:
		return "leave it if it stands between two numbers"
	case mark.Signal == "em-dash":
		return dashAction(mark.Fix.Text)
	case mark.Signal == "arrow":
		return fmt.Sprintf("write %q", mark.Fix.Text)
	case mark.Signal == "bold-label-in-list":
		return "remove the bold"
	}
	return "remove it"
}

// dashAction names the replacement of an em dash from the text of its fix.
func dashAction(text string) string {
	switch {
	case strings.HasPrefix(text, "."):
		return "write a period and start the next word with a capital letter"
	case strings.HasPrefix(text, ","):
		return "write a comma"
	case strings.HasPrefix(text, ":"):
		return "write a colon"
	}
	return "remove it"
}

// aiAdviceLines says, once for the run, that the command makes the
// replacements named at the listed marks, and that a run of sentences has
// none.
func aiAdviceLines(measured []unswell.DocumentResult) []string {
	fixable, runs := false, false
	for _, doc := range measured {
		marks, _ := doc.AISounding.Listed()
		for _, mark := range marks {
			fixable = fixable || mark.Fix != nil
			runs = runs || mark.Sentences > 0
		}
	}
	var lines []string
	if fixable {
		lines = append(lines, "`unswell fix --write` makes the replacements named above. A stock word needs your own words.")
	}
	if runs {
		lines = append(lines, "A run of sentences has no replacement that a command can make. Rewrite one or two of its sentences, "+
			"and keep what they say.",
			"Do not add an em dash anywhere: an em dash is one of the marks that this report counts.")
	}
	return lines
}

// aiSignalText lists the counted marks of one page, or says that it has none.
// Runs of sentences are named only when their reading is above the low level:
// a few plain sentences in a row are on nearly every page.
func aiSignalText(index unswell.AISounding) string {
	var parts []string
	for _, signal := range index.Signals {
		if signal.Reading == unswell.AIReadingRhythm && unswell.AILevel(index.Reading(signal.Reading)) == unswell.AISoundingLow {
			continue
		}
		if signal.Count > 0 {
			name := aiSignalNames[signal.Name][1]
			if signal.Count == 1 {
				name = aiSignalNames[signal.Name][0]
			}
			parts = append(parts, fmt.Sprintf("%d %s", signal.Count, name))
		}
	}
	if len(parts) == 0 {
		return ": none of its marks."
	}
	return ": " + strings.Join(parts, ", ") + "."
}
