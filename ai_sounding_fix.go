package unswell

import (
	"bytes"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
)

// RemoveAIMarks applies the fixes of the marks to the source they were found
// in and returns the new source and the number of fixes applied. A fix that
// overlaps an earlier one, or that does not fit the source, is left out; a
// second analysis of the result finds what remains. Marks without a fix stay.
func RemoveAIMarks(source []byte, marks []AIMark) ([]byte, int) {
	fixes := make([]AIMarkFix, 0, len(marks))
	for _, mark := range marks {
		if mark.Fix != nil {
			fixes = append(fixes, *mark.Fix)
		}
	}
	slices.SortStableFunc(fixes, func(a, b AIMarkFix) int { return a.Span.Start - b.Span.Start })
	var result bytes.Buffer
	applied, offset := 0, 0
	for _, fix := range fixes {
		if fix.Span.Start < offset || fix.Span.End < fix.Span.Start || fix.Span.End > len(source) {
			continue
		}
		result.Write(source[offset:fix.Span.Start])
		result.WriteString(fix.Text)
		offset = fix.Span.End
		applied++
	}
	if applied == 0 {
		return source, 0
	}
	result.Write(source[offset:])
	return result.Bytes(), applied
}

// aiMarkFix returns the replacement that removes one mark, or nil when the
// mark needs a person: a stock word, or a dash between two numbers.
func aiMarkFix(signal string, mark AIMark, source []byte) *AIMarkFix {
	span := mark.Span
	switch signal {
	case "em-dash":
		return dashFix(span, source, mark.clauses)
	case "emoji-in-heading", "emoji-at-list-item":
		return emojiFix(span, source)
	case "arrow":
		text := "->"
		if string(source[span.Start:span.End]) == "⇒" {
			text = "=>"
		}
		return &AIMarkFix{Span: span, Text: text}
	case "bold-label-in-list":
		// The span covers the label with its asterisks; the label stays.
		return &AIMarkFix{Span: span, Text: string(source[span.Start+2 : span.End-2])}
	}
	return nil
}

// lineOpening matches the part of a line before its first word: indentation,
// a quote mark, a list marker, and a task box.
var lineOpening = regexp.MustCompile(`^[ \t>]*(?:(?:[-*+]|\d+[.)])[ \t]+)?(?:\[[ xX]\][ \t]+)?$`)

// listOpening matches the opening of a line that starts a list item, and
// codeSpan a run of inline code.
var (
	listOpening = regexp.MustCompile(`^[ \t>]*(?:[-*+]|\d+[.)])[ \t]+(?:\[[ xX]\][ \t]+)?`)
	codeSpan    = regexp.MustCompile("`[^`]*`")
)

// maxLabelWords bounds the words of a list item that read as its label.
const maxLabelWords = 4

// listLabel reports whether the text of a line before a dash is the short
// label of a list item, such as a name, a bold term, or a command.
func listLabel(before []byte) bool {
	opening := listOpening.FindIndex(before)
	if opening == nil {
		return false
	}
	label := before[opening[1]:]
	words := len(bytes.Fields(label))
	return words >= 1 && words <= maxLabelWords && !bytes.ContainsAny(codeSpan.ReplaceAll(label, nil), ".!?,;:—")
}

// dashFix replaces an em dash and the spaces around it. After the short label
// of a list item the dash becomes a colon; at the start of a line or after
// other punctuation it is dropped; between two full clauses it becomes a
// period, and the next word takes a capital letter; elsewhere it becomes a
// comma.
func dashFix(span document.Span, source []byte, clauses bool) *AIMarkFix {
	left, right := span.Start, span.End
	for left > 0 && blank(source[left-1]) {
		left--
	}
	for right < len(source) && blank(source[right]) {
		right++
	}
	lineStart := bytes.LastIndexByte(source[:left], '\n') + 1
	if lineOpening.Match(source[lineStart:span.Start]) {
		// The line keeps its opening and the spaces that were there.
		return &AIMarkFix{Span: document.Span{Start: span.Start, End: right}}
	}
	after := byte('\n')
	if right < len(source) {
		after = source[right]
	}
	text, replaced := dashReplacement(source[lineStart:left], after)
	if !replaced {
		return nil
	}
	fix := &AIMarkFix{Span: document.Span{Start: left, End: right}, Text: text}
	if clauses && strings.HasPrefix(text, ",") {
		sentenceBreak(fix, source)
	}
	return fix
}

// dashReplacement chooses what stands for a dash from the text of its line
// before it, which is not empty, and the character after it. It reports
// false for a dash between two numbers, which is a range.
func dashReplacement(before []byte, after byte) (string, bool) {
	space := " "
	if after == '\n' || after == '\r' {
		space = ""
	}
	// The character before the dash, behind any closing emphasis.
	last := bytes.TrimRight(before, "*_")
	switch {
	case digit(before[len(before)-1]) && digit(after):
		return "", false
	case len(last) > 0 && strings.IndexByte(",;:.!?(", last[len(last)-1]) >= 0:
		return space, true
	case listLabel(before):
		return ":" + space, true
	}
	return "," + space, true
}

// sentenceBreak turns the comma of a fix into a period and gives the word
// after it a capital letter. A word that starts with markup or code, or that
// is not a lower-case letter, stays as it is.
func sentenceBreak(fix *AIMarkFix, source []byte) {
	next := fix.Span.End
	for next < len(source) && (blank(source[next]) || source[next] == '\n' || source[next] == '\r') {
		next++
	}
	fix.Text = "." + fix.Text[1:]
	if next < len(source) && source[next] >= 'a' && source[next] <= 'z' {
		fix.Text += string(source[fix.Span.End:next]) + string(source[next]-'a'+'A')
		fix.Span.End = next + 1
	}
}

// betweenClauses reports whether an em dash stands between two full clauses
// of one sentence, so that a period can stand for it. The conditions are
// strict, because a wrong period makes a fragment: the sentence has one dash,
// does not open with a subordinating word, has a finite verb and no open
// parenthesis before the dash, and after the dash a subject followed within a
// few words by a verb in the present tense or a modal.
func betweenClauses(block document.Block, dash document.Span) bool {
	for _, sentence := range block.Sentences {
		if sentence.Span.Start <= dash.Start && sentence.Span.End >= dash.End {
			return clausePair(sentence.Tokens, dash)
		}
	}
	return false
}

// maxSubjectWords bounds the words between a dash and the verb of the clause
// after it.
const maxSubjectWords = 5

func clausePair(tokens []document.Token, dash document.Span) bool {
	verbBefore, dashes, open := false, 0, 0
	var before, after []document.Token
	for _, token := range tokens {
		span := document.Bounds(token.Spans)
		switch {
		case strings.Contains(token.Text, "—"):
			dashes++
		case span.End <= dash.Start:
			before = append(before, token)
			verbBefore = verbBefore || finiteVerb(token.Tag)
			open += strings.Count(token.Text, "(") - strings.Count(token.Text, ")")
		case span.Start >= dash.End:
			after = append(after, token)
		}
	}
	return dashes == 1 && verbBefore && open <= 0 && !subordinate(before) && opensClause(after)
}

// subordinate reports a clause that opens with a word such as "although",
// "if", or "when", which cannot stand as a sentence.
func subordinate(tokens []document.Token) bool {
	for _, token := range tokens {
		if token.Word {
			return token.Tag == "IN" || token.Tag == "WRB"
		}
	}
	return false
}

// opensClause reports whether the tokens start with a subject and reach a
// verb in the present tense or a modal before anything that would make them
// a dependent phrase. The past tense is left out: the tagger gives it to
// participles as well.
func opensClause(tokens []document.Token) bool {
	words := 0
	for _, token := range tokens {
		if token.Protected {
			// Inline code stands for a name.
			words++
			continue
		}
		if !token.Word || (words == 0 && !subjectStart(token.Tag)) {
			return false
		}
		if presentVerb(token.Tag) {
			return words > 0
		}
		if words >= maxSubjectWords || endsSubject(token.Tag, words) {
			return false
		}
		words++
	}
	return false
}

func presentVerb(tag string) bool { return tag == "VBZ" || tag == "VBP" || tag == "MD" }

func finiteVerb(tag string) bool { return tag == "VBZ" || tag == "VBP" || tag == "VBD" || tag == "MD" }

func subjectStart(tag string) bool {
	switch tag {
	case "PRP", "PRP$", "DT", "EX", "NN", "NNS", "NNP", "NNPS":
		return true
	}
	return false
}

// endsSubject reports a tag that cannot stand inside the subject of a main
// clause after the given number of its words: a word that opens a dependent
// phrase, a past form, or a pronoun after the first word, which opens a
// relative clause as in "the thing you use".
func endsSubject(tag string, words int) bool {
	switch tag {
	case "WDT", "WP", "WP$", "WRB", "IN", "CC", "TO", "VBD", "VBN":
		return true
	case "PRP":
		return words > 0
	}
	return false
}

// emojiFix removes an emoji with the marks that belong to it and the spaces
// after it, or the spaces before it when it ends the line.
func emojiFix(span document.Span, source []byte) *AIMarkFix {
	right := span.End
	for right < len(source) {
		next, size := utf8.DecodeRune(source[right:])
		if !emojiPart(next) {
			break
		}
		right += size
	}
	left := span.Start
	end := right
	for end < len(source) && blank(source[end]) {
		end++
	}
	if end == len(source) || source[end] == '\n' || source[end] == '\r' {
		for left > 0 && blank(source[left-1]) {
			left--
		}
		return &AIMarkFix{Span: document.Span{Start: left, End: right}}
	}
	return &AIMarkFix{Span: document.Span{Start: left, End: end}}
}

// emojiPart reports a rune that continues the emoji before it: a variation
// selector, a joiner, a skin tone, or a combining mark.
func emojiPart(r rune) bool {
	return r == 0xFE0F || r == 0x200D || (r >= 0x1F3FB && r <= 0x1F3FF) || unicode.Is(unicode.Mn, r)
}

func blank(b byte) bool { return b == ' ' || b == '\t' }

func digit(b byte) bool { return b >= '0' && b <= '9' }
