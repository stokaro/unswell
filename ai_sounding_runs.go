package unswell

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// aiRunSignal is a signal of the index over sentences in a row. A run is a
// sequence of consecutive sentences of running paragraphs that all fit one
// condition. A run of at least `least` sentences is one mark, and the signal
// weighs the sentences in its runs per 1,000 prose words.
//
// A page that a model wrote under a style guide has few of the marks that the
// other signals count, and still reads as a model's: its sentences have the
// same length and the same shape, one after the other. The run signals are a
// second reading of the page, the rhythm of its sentences, and the index is
// the higher of the two readings. ADR 0060 records what they were measured
// against.
type aiRunSignal struct {
	name   string
	weight float64
	advice string
	least  int
	fits   func(sentence document.Sentence) bool
}

// aiRunSignals are the run signals of the index with the weights fitted on
// the fitting half of their record; aiRhythmBias is the constant term of
// their reading.
var aiRunSignals = []aiRunSignal{
	{"plain-sentence-run", 3.2, "Say it as you would to a colleague: join two of these sentences, use a contraction, " +
		"put an aside in parentheses, or start one with And or But.", 2, plainSentence},
	{"subject-opener-run", 0.35, "Open one of these sentences with its condition, its time, or its reason.", 3, opensWithSubject},
}

const aiRhythmBias = -12.3

// Bounds of a plain sentence, in words.
const (
	plainSentenceMinWords = 5
	plainSentenceMaxWords = 20
)

// plainSentence reports a plain statement of middling length: five to twenty
// words, with no question, exclamation, parenthesis, colon, semicolon, or
// dash to break its line.
func plainSentence(sentence document.Sentence) bool {
	return sentence.Words >= plainSentenceMinWords && sentence.Words <= plainSentenceMaxWords &&
		!strings.ContainsAny(sentence.Text, "?!();:—")
}

// subjectOpeners are the first words of a sentence that starts with its
// subject and nothing before it.
var subjectOpeners = map[string]bool{
	"the": true, "this": true, "that": true, "these": true, "those": true, "it": true,
	"you": true, "we": true, "they": true, "each": true, "a": true, "an": true,
}

// opensWithSubject reports a sentence whose first word is an article, a
// demonstrative, or a personal pronoun.
func opensWithSubject(sentence document.Sentence) bool {
	for _, token := range sentence.Tokens {
		if token.Word {
			return subjectOpeners[strings.ToLower(token.Text)]
		}
	}
	return false
}

// paragraphStretches returns the sentences of the running paragraphs in the
// order of the page, cut into stretches. A stretch ends at a heading, a list
// item, a table, a quote, and at anything between two paragraphs that is not
// read as prose, such as a code block.
func paragraphStretches(doc document.Document) [][]document.Sentence {
	var stretches [][]document.Sentence
	var current []document.Sentence
	end := -1
	flush := func() {
		if len(current) > 0 {
			stretches = append(stretches, current)
		}
		current, end = nil, -1
	}
	for _, block := range doc.Blocks {
		if block.Excluded || block.Kind != "paragraph" || block.Words == 0 || !block.Span.Valid(len(doc.Source)) {
			flush()
			continue
		}
		if end >= 0 && (block.Span.Start < end || strings.TrimSpace(string(doc.Source[end:block.Span.Start])) != "") {
			flush()
		}
		current = append(current, block.Sentences...)
		end = block.Span.End
	}
	flush()
	return stretches
}

// sentenceRuns finds the runs of one signal on a page: one mark for each
// run, and the number of sentences in all of them.
func sentenceRuns(stretches [][]document.Sentence, signal aiRunSignal) ([]AIMark, int) {
	var marks []AIMark
	total := 0
	for _, stretch := range stretches {
		start := -1
		for index := 0; index <= len(stretch); index++ {
			if index < len(stretch) && signal.fits(stretch[index]) {
				if start < 0 {
					start = index
				}
				continue
			}
			if start >= 0 && index-start >= signal.least {
				marks = append(marks, AIMark{
					Span:      document.Span{Start: stretch[start].Span.Start, End: stretch[index-1].Span.End},
					Sentences: index - start,
				})
				total += index - start
			}
			start = -1
		}
	}
	return marks, total
}
