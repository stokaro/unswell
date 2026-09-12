package english

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jdkato/prose/v3/segment"
	"github.com/jdkato/prose/v3/tokenize"
)

// Punkt treats a dotted identifier such as chi.Router or v1.2 followed by a
// period as an abbreviation and deletes the sentence break after it. The break
// is restored only when the next token is capitalized and known to the model,
// so a protected code span never restores it. These shapes are the tokens the
// tokenizer leaves whole; e.g., U.S., and Dr. keep their period attached and
// never reach this check.
var (
	dottedIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)+$`)
	dottedVersion    = regexp.MustCompile(`^v?[0-9]+(\.[0-9]+)+$`)
	// dottedAbbreviations lists lowercase abbreviation shapes that must keep
	// their sentence joined even when they look like a dotted identifier.
	dottedAbbreviations = []string{"e.g", "i.e", "u.s", "u.k", "a.m", "p.m", "ph.d"}
	openingPunctuation  = []string{`"`, "(", "[", "“", "‘"}
)

// piece is one repaired sentence inside a Punkt segment. Token offsets are
// relative to sentence.Text, matching the tokenizer's invariant.
type piece struct {
	sentence segment.Sentence
	tokens   []tokenize.Token
}

// repairBoundaries splits one Punkt segment at every interior period that ends
// a dotted identifier or version and precedes a sentence opener. A segment
// without such a point is returned unchanged.
func repairBoundaries(sent segment.Sentence, tokens []tokenize.Token) []piece {
	points := splitPoints(tokens)
	if len(points) == 0 {
		return []piece{{sentence: sent, tokens: tokens}}
	}
	result := make([]piece, 0, len(points)+1)
	first, start := 0, 0
	for _, point := range append(points, len(tokens)) {
		end := len(sent.Text)
		if point < len(tokens) {
			end = tokens[point-1].End()
		}
		result = append(result, subPiece(sent, tokens[first:point], start, end))
		if point < len(tokens) {
			first, start = point, tokens[point].Start
		}
	}
	return result
}

// splitPoints returns the token indexes that open a repaired sentence.
func splitPoints(tokens []tokenize.Token) []int {
	var points []int
	for i := 1; i+1 < len(tokens); i++ {
		if tokens[i].Text != "." || tokens[i-1].End() != tokens[i].Start {
			continue
		}
		if identifierBeforePeriod(tokens[i-1].Text) && sentenceOpener(tokens[i+1:]) {
			points = append(points, i+1)
		}
	}
	return points
}

func subPiece(sent segment.Sentence, tokens []tokenize.Token, start, end int) piece {
	rebased := make([]tokenize.Token, len(tokens))
	for i, tok := range tokens {
		tok.Start -= start
		rebased[i] = tok
	}
	return piece{sentence: segment.Sentence{Text: sent.Text[start:end], Start: sent.Start + start}, tokens: rebased}
}

func identifierBeforePeriod(text string) bool {
	if slices.Contains(dottedAbbreviations, strings.ToLower(text)) {
		return false
	}
	return dottedIdentifier.MatchString(text) || dottedVersion.MatchString(text)
}

// sentenceOpener reports whether the tokens after a period start a sentence:
// a capitalized word, optionally after opening punctuation, or a protected
// placeholder, which Punkt cannot score at all.
func sentenceOpener(tokens []tokenize.Token) bool {
	if len(tokens) > 1 && slices.Contains(openingPunctuation, tokens[0].Text) {
		tokens = tokens[1:]
	}
	if strings.ContainsRune(tokens[0].Text, 0) {
		return true
	}
	first, _ := utf8.DecodeRuneInString(tokens[0].Text)
	return unicode.IsUpper(first)
}
