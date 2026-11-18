package builtin

import (
	"context"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// Bounds of a closing denial, in words.
const (
	// denialReach is how far from the start of a clause its negation may stand.
	denialReach = 6
	// denialTail bounds the words after ", not" at the end of a sentence.
	denialTail = 12
	// denialLateTail bounds the words after a negation that stands late in its clause.
	denialLateTail = 9
)

// denialWords deny what a clause says. A word that ends in n't counts too.
var denialWords = []string{"not", "cannot", "never", "neither", "nor"}

// clauseOpeners start a clause that states a condition, a time, or a reason.
// A negation in such a clause belongs to the condition, not to a denial.
var clauseOpeners = []string{"if", "when", "whenever", "unless", "until", "because", "although", "though", "while",
	"since", "once", "after", "before", "even", "so", "thus", "as", "whether"}

// instructionOpeners start a sentence that tells the reader what not to do.
var instructionOpeners = []string{"do", "don't", "never", "avoid", "please", "note"}

// readerWords address the reader. A sentence about what the reader cannot do is a warning, not a denial.
var readerWords = []string{"you", "your", "let's"}

// boundedNegations follow "not" where it limits a statement and denies nothing.
var boundedNegations = []string{"only", "just", "uncommon", "unusual", "necessarily", "all", "always", "every", "yet", "quite"}

// statusWords make a negation a report of where things stand now.
var statusWords = []string{"yet", "anymore"}

// denyingVerbs stand before a negation that denies a fact: "is not", "does
// not", "can't". A contraction is cut after its verb, so "ca" and "wo" are
// here too. "May not" hedges and "should not" advises, and neither denies.
var denyingVerbs = []string{"is", "are", "was", "were", "am", "do", "does", "did", "has", "have", "had", "can", "ca", "could",
	"will", "wo", "would"}

// embeddingWords open a clause that runs to the end of its sentence. A
// negation after one of them belongs to that clause.
var embeddingWords = []string{"where", "why", "how", "what"}

// relativeWords open a clause that can close before the verb of the main clause.
var relativeWords = []string{"that", "which", "who", "whom", "whose"}

// finiteTags are the tags of a verb that a clause can stand on.
var finiteTags = []string{"VBZ", "VBP", "VBD", "MD"}

// quotationMarks open and close the words of someone else.
var quotationMarks = []string{"\"", "“", "”", "``", "''"}

// closingDenials reports a page whose paragraphs end on a denial: a last
// sentence that says what something is not, or what it does not do. The page
// is reported once, with each such sentence.
func closingDenials(ctx context.Context, view rule.View, emit rule.Emitter) error {
	var occurrences []rule.Occurrence
	paragraphs := 0
	for _, block := range view.Document.Blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if block.Excluded || block.Kind != "paragraph" || len(block.Sentences) < 2 {
			continue
		}
		paragraphs++
		if last := block.Sentences[len(block.Sentences)-1]; closingDenial(last) {
			occurrences = append(occurrences, sentenceOccurrence(last))
		}
	}
	p := view.Parameters
	count := len(occurrences)
	if count <= p.AllowedOccurrences || count*100 < p.Onset*paragraphs {
		return ctx.Err()
	}
	share := float64(count) * 100 / float64(paragraphs)
	return emit.Emit(rule.Evidence{Kind: "heuristic", Activation: 1000, Occurrences: occurrences, Metrics: []rule.Metric{
		{Name: "closing-denial-share", Value: share, Unit: "percent-of-paragraphs", Onset: float64(p.Onset), Saturation: float64(p.Onset) * 4},
		{Name: "closing-denial-count", Value: float64(count), Unit: "paragraphs"},
		{Name: "paragraph-denominator", Value: float64(paragraphs), Unit: "paragraphs"},
	}})
}

// closingDenial reports a sentence that says what something is not or does
// not do, in one of four forms: a last clause after a semicolon or a dash
// that opens with a negation, a tail of the form ", not X", a main clause
// with its negation among the first words, or a clause whose negation stands
// late, after a long subject.
func closingDenial(sentence document.Sentence) bool {
	text := strings.TrimSpace(sentence.Text)
	if strings.HasSuffix(text, "?") || strings.HasSuffix(text, ":") {
		return false
	}
	words := wordIndexes(sentence.Tokens)
	if len(words) < 3 || hasNormal(sentence.Tokens, words, statusWords) {
		return false
	}
	// The last clause is what a sentence ends on, so a late negation is read there.
	last := words
	if tail := lastClause(sentence.Tokens, words); tail != nil {
		if deniesEarly(sentence.Tokens, tail) {
			return true
		}
		last = tail
	}
	return commaNot(sentence.Tokens) || (!addressesReader(sentence.Tokens, words) && deniesEarly(sentence.Tokens, words)) ||
		deniesLate(sentence.Tokens, last)
}

// wordIndexes returns the positions of the words among the tokens.
func wordIndexes(tokens []document.Token) []int {
	words := make([]int, 0, len(tokens))
	for index, token := range tokens {
		if token.Word {
			words = append(words, index)
		}
	}
	return words
}

func hasNormal(tokens []document.Token, words []int, list []string) bool {
	for _, index := range words {
		if frameWord(tokens[index], list...) {
			return true
		}
	}
	return false
}

// lastClause returns the words after the last semicolon or dash of a
// sentence, or nil when the sentence has one clause. A hyphen is a dash only
// when it stands apart from the words around it.
func lastClause(tokens []document.Token, words []int) []int {
	separator := -1
	for index := range tokens {
		if separates(tokens, index) {
			separator = index
		}
	}
	if separator < 0 {
		return nil
	}
	start, _ := slices.BinarySearch(words, separator)
	if len(words)-start < 2 {
		return nil
	}
	return words[start:]
}

// separates reports a token that ends a clause: a semicolon, a dash, or a
// hyphen that stands apart from the words around it.
func separates(tokens []document.Token, index int) bool {
	token := tokens[index]
	if token.Word || token.Protected {
		return false
	}
	switch token.Text {
	case ";", "—", "–":
		return true
	case "-", "--":
		return index > 0 && index+1 < len(tokens) && tokens[index-1].End < token.Start && token.End < tokens[index+1].Start
	}
	return false
}

// deniesEarly reports a clause that is no condition and no instruction and
// has a negation among its first words.
func deniesEarly(tokens []document.Token, words []int) bool {
	if len(words) == 0 || frameWord(tokens[words[0]], clauseOpeners...) || frameWord(tokens[words[0]], instructionOpeners...) {
		return false
	}
	for position, index := range words[:min(denialReach, len(words))] {
		if !negation(tokens[index]) {
			continue
		}
		next := position + 1
		if next < len(words) && frameWord(tokens[words[next]], boundedNegations...) {
			return false
		}
		return true
	}
	return false
}

// deniesLate reports a clause that denies after a long subject: "Replaying an
// old history on a database that holds data is not a plan for a cutover." Its
// last negation stands past the first words, with a few words after it, and
// denies the verb of the main clause. "Not ... until" states a time, and a
// negation inside quotation marks is quoted.
func deniesLate(tokens []document.Token, words []int) bool {
	position := lastNegation(tokens, words)
	after := len(words) - 1 - position
	if position < denialReach || after < 1 || after > denialLateTail {
		return false
	}
	verb := deniedVerb(tokens, words, position)
	if verb < 0 || frameWord(tokens[words[position+1]], boundedNegations...) || hasNormal(tokens, words[position+1:], []string{"until"}) {
		return false
	}
	return !quoted(tokens, words[position]) && statesFact(tokens, words[:verb]) && mainVerb(tokens, words, verb)
}

// quoted reports a token inside quotation marks: the words of someone else,
// which the sentence repeats and does not say.
func quoted(tokens []document.Token, index int) bool {
	marks := 0
	for _, token := range tokens[:index] {
		if !token.Word && !token.Protected && slices.Contains(quotationMarks, token.Text) {
			marks++
		}
	}
	return marks%2 == 1
}

// lastNegation returns the position of the last negation among the words, or -1.
func lastNegation(tokens []document.Token, words []int) int {
	for position, index := range slices.Backward(words) {
		if negation(tokens[index]) {
			return position
		}
	}
	return -1
}

// deniedVerb returns the position of the verb that a negation denies: the
// word "cannot", or one of denyingVerbs right before the negation or
// contracted with it. It returns -1 when the negation follows no such verb.
func deniedVerb(tokens []document.Token, words []int, position int) int {
	normal := strings.ReplaceAll(tokens[words[position]].Normal, "’", "'")
	if stem, contracted := strings.CutSuffix(normal, "n't"); normal == "cannot" || (contracted && slices.Contains(denyingVerbs, stem)) {
		return position
	}
	if position > 0 && frameWord(tokens[words[position-1]], denyingVerbs...) {
		return position - 1
	}
	return -1
}

// statesFact reports the words before a verb that give no instruction, open
// no condition, time, reason, or embedded question, and do not address the
// reader.
func statesFact(tokens []document.Token, before []int) bool {
	if len(before) == 0 || frameWord(tokens[before[0]], instructionOpeners...) {
		return false
	}
	return !hasNormal(tokens, before, clauseOpeners) && !hasNormal(tokens, before, embeddingWords) && !hasNormal(tokens, before, readerWords)
}

// mainVerb reports a verb that stands in no relative clause. A clause that
// "that", "which", or "who" opens before the verb has to close first: it has a
// finite verb of its own, and "and", "but", or "or" does not join the verb to
// it. "The rows that the pass does not match" denies nothing about the rows.
func mainVerb(tokens []document.Token, words []int, verb int) bool {
	opener := -1
	for position := verb - 1; position >= 0 && opener < 0; position-- {
		if frameWord(tokens[words[position]], relativeWords...) {
			opener = position
		}
	}
	if opener < 0 {
		return true
	}
	if frameWord(tokens[words[verb-1]], "and", "but", "or") {
		return false
	}
	return slices.ContainsFunc(words[opener+1:verb], func(index int) bool {
		return !tokens[index].Protected && slices.Contains(finiteTags, tokens[index].Tag)
	})
}

// addressesReader reports a sentence that speaks to the reader in its first words.
func addressesReader(tokens []document.Token, words []int) bool {
	return hasNormal(tokens, words[:min(denialReach, len(words))], readerWords)
}

// negation reports a word that denies: one of denialWords, or a contraction of "not".
func negation(token document.Token) bool {
	if token.Protected {
		return false
	}
	normal := strings.ReplaceAll(token.Normal, "’", "'")
	return slices.Contains(denialWords, normal) || strings.HasSuffix(normal, "n't")
}

// commaNot reports a sentence that ends on ", not X": a comma, an optional
// "and" or "but", the word "not", and a few words with no comma after them.
func commaNot(tokens []document.Token) bool {
	for index := len(tokens) - 1; index > 0; index-- {
		token := tokens[index]
		if !token.Word && !token.Protected && (token.Text == "," || token.Text == ";") {
			return notAfter(tokens[index+1:])
		}
	}
	return false
}

func notAfter(tokens []document.Token) bool {
	words := wordIndexes(tokens)
	if len(words) > 0 && frameWord(tokens[words[0]], "and", "but") {
		words = words[1:]
	}
	if len(words) < 2 || len(words) > denialTail+1 || !frameWord(tokens[words[0]], "not") {
		return false
	}
	return !frameWord(tokens[words[1]], boundedNegations...)
}
