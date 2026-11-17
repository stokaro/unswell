package builtin

import (
	"context"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// casualOpeners start a sentence the way a person does when talking.
var casualOpeners = []string{"and", "but", "so", "or", "yet"}

// contractedSubjects are the words whose 's is a contraction of "is" or "has" and not a possessive.
var contractedSubjects = []string{"it", "that", "there", "here", "what", "who", "he", "she", "let"}

// minAsideWords is how many words an aside in parentheses has at least. One
// word in parentheses is an abbreviation or a unit.
const minAsideWords = 2

// uniformRegister reports a page of running prose that never leaves its one
// formal register: among its sentences there is no contraction, no aside in
// parentheses, no question, no exclamation, and no sentence that opens with
// And, But, or So. The page is reported once, at its first sentence.
func uniformRegister(ctx context.Context, view rule.View, emit rule.Emitter) error {
	sentences := runningSentences(view.Document)
	casual := 0
	for _, sentence := range sentences {
		if err := ctx.Err(); err != nil {
			return err
		}
		if casualSentence(*sentence) {
			casual++
		}
	}
	p := view.Parameters
	if len(sentences) == 0 || len(sentences) < p.MinSentences || casual > p.AllowedOccurrences {
		return ctx.Err()
	}
	return emit.Emit(rule.Evidence{Kind: "heuristic", Activation: 1000, Occurrences: []rule.Occurrence{sentenceOccurrence(*sentences[0])},
		Metrics: []rule.Metric{
			{Name: "casual-sentences", Value: float64(casual), Unit: "sentences"},
			{Name: "sentence-denominator", Value: float64(len(sentences)), Unit: "sentences", Onset: float64(p.MinSentences)},
		}})
}

// runningSentences returns the sentences of the running paragraphs of a
// document, in the order of the page. A sentence without a word is left out.
func runningSentences(doc *document.Document) []*document.Sentence {
	var sentences []*document.Sentence
	for index := range doc.Blocks {
		block := &doc.Blocks[index]
		if block.Excluded || block.Kind != "paragraph" {
			continue
		}
		for position := range block.Sentences {
			if block.Sentences[position].Words > 0 {
				sentences = append(sentences, &block.Sentences[position])
			}
		}
	}
	return sentences
}

// casualSentence reports a sentence with one of the marks of a person
// talking: a question, an exclamation, an opening And, But, or So, a
// contraction, or an aside in parentheses.
func casualSentence(sentence document.Sentence) bool {
	text := strings.TrimRight(strings.TrimSpace(sentence.Text), "\"'”’)]*_")
	if strings.HasSuffix(text, "?") || strings.HasSuffix(text, "!") {
		return true
	}
	opened := false
	for index, token := range sentence.Tokens {
		if token.Word && !opened {
			opened = true
			if frameWord(token, casualOpeners...) {
				return true
			}
		}
		if contraction(sentence.Tokens, index) {
			return true
		}
	}
	return aside(sentence.Tokens)
}

// contraction reports a contracted word: n't, 're, 've, 'll, 'd, 'm, and 's
// after a pronoun. A possessive is not one.
func contraction(tokens []document.Token, index int) bool {
	token := tokens[index]
	if token.Protected {
		return false
	}
	stem, suffix, found := strings.Cut(strings.ReplaceAll(token.Normal, "’", "'"), "'")
	if !found {
		return false
	}
	switch suffix {
	case "t":
		return strings.HasSuffix(stem, "n")
	case "re", "ve", "ll", "d", "m":
		return true
	case "s":
		if stem == "" && index > 0 {
			stem = tokens[index-1].Normal
		}
		return slices.Contains(contractedSubjects, stem)
	}
	return false
}

// aside reports parentheses that hold at least minAsideWords words.
func aside(tokens []document.Token) bool {
	inside, open := 0, false
	for _, token := range tokens {
		switch {
		case !token.Word && token.Text == "(":
			inside, open = 0, true
		case !token.Word && token.Text == ")":
			if open && inside >= minAsideWords {
				return true
			}
			open = false
		case open && token.Word:
			inside++
		}
	}
	return false
}
