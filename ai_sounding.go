package unswell

import (
	"bytes"
	"math"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/glyph"
)

// AISounding is a page-level index of marks that people remove when they
// clean AI-sounding wording out of documentation. It is an index from 0 to
// 100, never a probability, and it says nothing about who wrote the page.
// ADR 0059 records the signals, the weights, and what the index was measured
// against.
type AISounding struct {
	// Status is "available", or says why there is no index: "too_short" for a
	// page with fewer prose words than the measurement covers.
	Status string `json:"status"`
	// Index is 0 to 100 and meaningful only when Status is "available".
	Index int `json:"index"`
	// Level is "low", "elevated", or "high". At most one in twenty pages that
	// people wrote before language models reach "high".
	Level string `json:"level,omitempty"`
	// Words is the number of prose words the rates are taken over.
	Words int `json:"prose_words"`
	// Signals lists every signal with its count, in a fixed order.
	Signals []AISignal `json:"signals,omitempty"`
	// Marks lists the counted marks in source order, so that a person or an
	// assistant can go to each one. A page with more than 200 marks lists the
	// first 200; Signals keep the full counts.
	Marks []AIMark `json:"marks,omitempty"`
}

// AISignal is one counted mark of an AISounding index.
type AISignal struct {
	Name        string  `json:"name"`
	Count       int     `json:"count"`
	PerThousand float64 `json:"per_thousand_words"`
	// Advice says how to remove a mark of this signal.
	Advice string `json:"advice"`
}

// AIMark is one counted mark and its place in the source.
type AIMark struct {
	// Signal names the signal that the mark counts for.
	Signal string `json:"signal"`
	// Span is the mark in the original bytes, and Start is its position.
	Span  document.Span     `json:"span"`
	Start document.Position `json:"start"`
	// Word is the matched word of the fixed list, in lower case, for a
	// stock-word mark. Other marks carry no text.
	Word string `json:"word,omitempty"`
	// Fix is the replacement that removes the mark without a rewrite of the
	// sentence. A stock word has none: only a person or a model can say what
	// the thing does. A dash between two numbers has none either.
	Fix *AIMarkFix `json:"fix,omitempty"`

	// clauses marks an em dash between two full clauses of one sentence,
	// where a comma would be a splice.
	clauses bool
}

// AIMarkFix replaces a range of the original bytes with a text. The range can
// reach past the mark to the spaces around it.
type AIMarkFix struct {
	Span document.Span `json:"span"`
	Text string        `json:"text"`
}

// Levels of an AISounding index.
const (
	AISoundingLow      = "low"
	AISoundingElevated = "elevated"
	AISoundingHigh     = "high"
)

// aiSoundingMinWords is the shortest page the index is computed for. The
// measurement used pages of 150 prose words or more.
const aiSoundingMinWords = 150

// aiSoundingMaxMarks bounds the marks that one page lists.
const aiSoundingMaxMarks = 200

// aiSignal is one signal of the index: its name, its weight on the natural
// logarithm of one plus its count per 1,000 prose words, how to remove one of
// its marks, and how its marks are found in one block. A found mark holds its
// span, and its word where the signal has one.
type aiSignal struct {
	name   string
	weight float64
	advice string
	find   func(block document.Block, source []byte) []AIMark
}

// stockWords are words of praise and stock verbs that documentation written
// by language models leans on. The list is fixed; it is a signal of the page
// index and reports no finding of its own.
var stockWords = regexp.MustCompile(`(?i)\b(?:seamless(?:ly)?|effortless(?:ly)?|powerful|robust|robustly|comprehensive|cutting-edge|` +
	`blazing(?:ly)?|lightning[- ]fast|production-ready|world-class|elegant|delightful|stunning|sleek|intuitive|game-chang\w+|` +
	`revolutionary|next-generation|state-of-the-art|enterprise-grade|battle-tested|supercharg\w+|unlock\w*|elevat\w+|empower\w*|` +
	`streamlin\w+|leverag\w+|harness\w*|delve\w*|showcas\w+|pivotal|crucial|meticulous\w*|vibrant|ever-evolving|landscape)\b`)

// aiSignals are the signals of the index with the weights fitted on the
// development part of the reference; aiSoundingBias is the constant term.
var aiSignals = []aiSignal{
	{"em-dash", 0.94, "Use a comma, a period, a colon, or parentheses, whichever the sentence needs.",
		func(block document.Block, _ []byte) []AIMark {
			marks := literalMarks(block, "—")
			for index := range marks {
				marks[index].clauses = betweenClauses(block, marks[index].Span)
			}
			return marks
		}},
	{"emoji-in-heading", 0.69, "Remove the emoji and keep the words of the heading.", func(block document.Block, _ []byte) []AIMark {
		if block.Kind != "heading" {
			return nil
		}
		return pictographMarks(block)
	}},
	{"emoji-at-list-item", 0.71, "Remove the emoji and start the item with its words.", func(block document.Block, source []byte) []AIMark {
		if block.Kind != "list-item" {
			return nil
		}
		first, size := utf8.DecodeRune(blockSource(block, source))
		if !glyph.Pictograph(first) {
			return nil
		}
		return []AIMark{{Span: document.Span{Start: block.Span.Start, End: block.Span.Start + size}}}
	}},
	{"arrow", 0.81, `Write the step in words, such as "then" or "gives". In a menu path, write ">".`,
		func(block document.Block, _ []byte) []AIMark {
			return append(literalMarks(block, "→"), literalMarks(block, "⇒")...)
		}},
	{"stock-word", 0.62, "Say what the thing does, with a fact, or drop the word.", func(block document.Block, _ []byte) []AIMark {
		var marks []AIMark
		for _, match := range stockWords.FindAllStringIndex(block.Text, -1) {
			marks = append(marks, AIMark{
				Span: document.Bounds(block.Spans(match[0], match[1])),
				Word: strings.ToLower(block.Text[match[0]:match[1]]),
			})
		}
		return marks
	}},
	{"bold-label-in-list", 0.44, "Remove the bold and keep the words, or make the label part of the sentence.", boldLabelMarks},
}

// boldLabelMarks finds the bold label that opens a list item. The block of a
// list item starts after its marker and after the opening marks of a bold
// label, so the label shows as two asterisks before the block and two more on
// its first line.
func boldLabelMarks(block document.Block, source []byte) []AIMark {
	body := blockSource(block, source)
	if block.Kind != "list-item" || body == nil || block.Span.Start < 2 ||
		string(source[block.Span.Start-2:block.Span.Start]) != "**" {
		return nil
	}
	line, _, _ := bytes.Cut(body, []byte("\n"))
	closing := bytes.Index(line, []byte("**"))
	if closing < 0 {
		return nil
	}
	return []AIMark{{Span: document.Span{Start: block.Span.Start - 2, End: block.Span.Start + closing + 2}}}
}

// literalMarks finds every occurrence of a literal in the text of a block.
func literalMarks(block document.Block, literal string) []AIMark {
	var marks []AIMark
	for offset := 0; ; {
		at := strings.Index(block.Text[offset:], literal)
		if at < 0 {
			return marks
		}
		offset += at + len(literal)
		marks = append(marks, AIMark{Span: document.Bounds(block.Spans(offset-len(literal), offset))})
	}
}

// pictographMarks finds the decorative emoji in the text of a block.
func pictographMarks(block document.Block) []AIMark {
	var marks []AIMark
	for offset, r := range block.Text {
		if glyph.Pictograph(r) {
			marks = append(marks, AIMark{Span: document.Bounds(block.Spans(offset, offset+utf8.RuneLen(r)))})
		}
	}
	return marks
}

const aiSoundingBias = -1.23

// Thresholds of the levels on the index.
const (
	aiSoundingElevatedAt = 40
	aiSoundingHighAt     = 70
)

// blockSource returns the source bytes of a block, or nil for a span outside
// the source.
func blockSource(block document.Block, source []byte) []byte {
	if block.Span.Start < 0 || block.Span.End > len(source) || block.Span.Start >= block.Span.End {
		return nil
	}
	return source[block.Span.Start:block.Span.End]
}

// aiSounding computes the index of one Markdown page. It returns nil for a
// document that is not a Markdown page, where the index was not measured.
func aiSounding(doc document.Document) *AISounding {
	if doc.Format != document.Markdown {
		return nil
	}
	counts := make([]int, len(aiSignals))
	var marks []AIMark
	words := 0
	for _, block := range doc.Blocks {
		if block.Excluded || block.Words == 0 {
			continue
		}
		words += block.Words
		for index, signal := range aiSignals {
			found := signal.find(block, doc.Source)
			counts[index] += len(found)
			marks = appendMarks(marks, signal.name, found, doc.Source)
		}
	}
	result := &AISounding{Status: "too_short", Words: words}
	if words < aiSoundingMinWords {
		return result
	}
	score := aiSoundingBias
	result.Signals = make([]AISignal, 0, len(aiSignals))
	for index, signal := range aiSignals {
		rate := 1000 * float64(counts[index]) / float64(words)
		score += signal.weight * math.Log1p(rate)
		result.Signals = append(result.Signals, AISignal{Name: signal.name, Count: counts[index], PerThousand: rate, Advice: signal.advice})
	}
	slices.SortStableFunc(marks, func(a, b AIMark) int { return a.Span.Start - b.Span.Start })
	result.Marks = marks
	result.Status = "available"
	result.Index = int(math.Round(100 / (1 + math.Exp(-score))))
	result.Level = AISoundingLow
	if result.Index >= aiSoundingHighAt {
		result.Level = AISoundingHigh
	} else if result.Index >= aiSoundingElevatedAt {
		result.Level = AISoundingElevated
	}
	return result
}

// appendMarks adds the found marks of one signal with their positions, up to
// the limit of one page. A mark whose span cannot be located is counted but
// not listed.
func appendMarks(marks []AIMark, signal string, found []AIMark, source []byte) []AIMark {
	for _, mark := range found {
		if len(marks) >= aiSoundingMaxMarks {
			break
		}
		if !mark.Span.Valid(len(source)) {
			continue
		}
		position, err := document.Locate(source, mark.Span.Start)
		if err != nil {
			continue
		}
		mark.Signal, mark.Start = signal, position
		mark.Fix = aiMarkFix(signal, mark, source)
		marks = append(marks, mark)
	}
	return marks
}
