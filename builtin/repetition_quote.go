package builtin

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
)

// adjacentQuoteScope follows an outer quotation across sentences in one block.
// Other quotation styles inside it cannot end it. Unclosed quotations protect
// the rest of that block. Protected tokens never alter the quotation state.
type adjacentQuoteScope struct {
	closing   rune
	skipUntil int
}

func (q *adjacentQuoteScope) tokens(text string, tokens []document.Token) []bool {
	quoted := make([]bool, len(tokens))
	for i, token := range tokens {
		quoted[i] = q.token(text, token)
	}
	return quoted
}

func (q *adjacentQuoteScope) token(text string, token document.Token) bool {
	if token.Protected {
		return true
	}
	quoted := q.closing != 0
	for pos := max(token.Start, q.skipUntil); pos < token.End; {
		char, size := utf8.DecodeRuneInString(text[pos:])
		quoted = q.delimiter(text, pos, char) || quoted
		pos = max(pos+size, q.skipUntil)
	}
	return quoted
}

func (q *adjacentQuoteScope) delimiter(text string, pos int, char rune) bool {
	if q.closing != 0 {
		return q.close(text, pos, char)
	}
	if char == '\'' && adjacentApostrophe(text, pos) || adjacentLetterBefore(text, pos) {
		return false
	}
	switch char {
	case '"', '\'':
		q.closing = char
	case '“':
		q.closing = '”'
	case '‘':
		q.closing = '’'
	case '`':
		if strings.HasPrefix(text[pos:], "``") {
			q.closing, q.skipUntil = '`', pos+2
		}
	}
	return q.closing != 0
}

func (q *adjacentQuoteScope) close(text string, pos int, char rune) bool {
	if q.closing == '`' && strings.HasPrefix(text[pos:], "''") {
		q.closing, q.skipUntil = 0, pos+2
		return true
	}
	if char != q.closing || (char == '\'' || char == '’') && adjacentApostrophe(text, pos) {
		return false
	}
	q.closing = 0
	return true
}

func adjacentLetterBefore(text string, pos int) bool {
	before, _ := utf8.DecodeLastRuneInString(text[:pos])
	return unicode.IsLetter(before) || unicode.IsNumber(before)
}

// An apostrophe within a word is not a quote. A plural possessive inside a
// single-quoted passage is ambiguous; retain protection when a word follows.
func adjacentApostrophe(text string, pos int) bool {
	if !adjacentLetterBefore(text, pos) {
		return false
	}
	char, size := utf8.DecodeRuneInString(text[pos:])
	after, _ := utf8.DecodeRuneInString(text[pos+size:])
	if unicode.IsLetter(after) || unicode.IsNumber(after) {
		return true
	}
	before, _ := utf8.DecodeLastRuneInString(text[:pos])
	if char != '\'' && char != '’' || before != 's' && before != 'S' || !unicode.IsSpace(after) {
		return false
	}
	next, _ := utf8.DecodeRuneInString(strings.TrimLeftFunc(text[pos+size:], unicode.IsSpace))
	return unicode.IsLetter(next)
}
