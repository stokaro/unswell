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
}

// aiSoundingSignalOrder is the order in which the reports name the signals.
var aiSoundingSignalOrder = []string{"em-dash", "emoji-in-heading", "emoji-at-list-item", "arrow", "stock-word", "bold-label-in-list"}

// maxAISoundingPages bounds the pages that the text reports name one by one,
// and maxAIMarkLines the marks they list across a run. The JSON report holds
// the marks of every page.
const (
	maxAISoundingPages = 20
	maxAIMarkLines     = 60
)

// aiSoundingLines reports the page index of the Markdown pages of a run: one
// line for each page above the low level, most marked first, with the places
// of its marks, then the count by level and what the index is. A run without
// a measured page adds nothing.
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
	return append(lines, "The index counts marks that people remove when they clean AI-sounding wording out of documentation. "+
		"It is not a probability and does not say who wrote a page.")
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
			lines = append(lines, fmt.Sprintf("AI-sounding index: %d of 100 (%s) in %s%s",
				doc.AISounding.Index, level, doc.Name, aiSignalText(*doc.AISounding)))
			var marks []string
			marks, left = aiMarkLines(doc, left)
			lines = append(lines, marks...)
		}
	}
	return lines
}

// aiMarkLines lists the places of the marks of one page above the low level,
// up to the lines left for the run, and returns the lines left after it.
func aiMarkLines(doc unswell.DocumentResult, left int) ([]string, int) {
	marks := doc.AISounding.Marks
	if doc.AISounding.Level == unswell.AISoundingLow || len(marks) == 0 {
		return nil, left
	}
	listed := min(len(marks), left)
	lines := make([]string, 0, listed+1)
	for _, mark := range marks[:listed] {
		name := aiSignalNames[mark.Signal][0]
		if mark.Word != "" {
			name += fmt.Sprintf(" %q", mark.Word)
		}
		lines = append(lines, fmt.Sprintf("  %s:%d:%d: %s", doc.Name, mark.Start.Line, mark.Start.Column, name))
	}
	total := 0
	for _, signal := range doc.AISounding.Signals {
		total += signal.Count
	}
	switch rest := total - listed; {
	case rest == 1:
		lines = append(lines, fmt.Sprintf("  1 more mark in %s is not listed here.", doc.Name))
	case rest > 1:
		lines = append(lines, fmt.Sprintf("  %d more marks in %s are not listed here.", rest, doc.Name))
	}
	return lines, left - listed
}

// aiAdviceLines says how to remove each kind of mark that a page above the
// low level has, once for the run.
func aiAdviceLines(measured []unswell.DocumentResult) []string {
	advice := map[string]string{}
	for _, doc := range measured {
		if doc.AISounding.Level == unswell.AISoundingLow {
			continue
		}
		for _, signal := range doc.AISounding.Signals {
			if signal.Count > 0 {
				advice[signal.Name] = signal.Advice
			}
		}
	}
	if len(advice) == 0 {
		return nil
	}
	lines := []string{"To lower the index, remove its marks:"}
	for _, signal := range aiSoundingSignalOrder {
		if text, found := advice[signal]; found {
			lines = append(lines, fmt.Sprintf("  %s: %s", aiSignalNames[signal][0], text))
		}
	}
	return lines
}

// aiSignalText lists the counted marks of one page, or says that it has none.
func aiSignalText(index unswell.AISounding) string {
	var parts []string
	for _, signal := range index.Signals {
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
