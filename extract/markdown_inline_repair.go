package extract

import (
	"bytes"

	"github.com/stokaro/unswell/document"
)

// maskInlineDefects repairs two inputs that the inline grammar rejects. It
// writes to the buffer that only the grammar reads and reports whether it
// changed a byte. It runs after a rejection, so a block that the grammar
// accepts is parsed from the same bytes as before. Each rewritten region is
// added to expected, so the repaired tree counts only when the grammar reads
// the region as the node that the repair meant.
//
// A code span that holds a "<": the grammar reads the text after the bracket
// as an HTML tag and loses the span. Each "<" inside the span is overwritten.
// A ">" stays, because it may end markup that began before the span.
//
// A "<" with no ">" after it in the block: it can open no tag and no
// autolink, and the grammar fails on some of these unfinished tags. One that
// carries an attribute value is HTML source, so it is rewritten as one short
// tag and stays out of the prose like other inline HTML. Any other such
// bracket is literal text and is overwritten with other punctuation.
func maskInlineDefects(source []byte, expected map[document.Span]string) bool {
	changed := false
	for _, span := range codeSpans(source) {
		code := source[span.Start:span.End]
		if bytes.IndexByte(code, '<') < 0 {
			continue
		}
		for index, char := range code {
			if char == '<' {
				code[index] = 'x'
			}
		}
		expected[span] = "code_span"
		changed = true
	}
	for pos := bytes.LastIndexByte(source, '>') + 1; pos < len(source); pos++ {
		if source[pos] != '<' {
			continue
		}
		changed = true
		end := unfinishedTagEnd(source, pos)
		if end < 0 {
			source[pos] = ','
			continue
		}
		for index := pos + 1; index < end-1; index++ {
			source[index] = 'x'
		}
		source[end-1] = '>'
		expected[document.Span{Start: pos, End: end}] = "html_tag"
		pos = end - 1
	}
	return changed
}

// unfinishedTagEnd reads an HTML open tag without its closing bracket at
// start: a name and attributes as CommonMark defines them. It returns the
// offset after the last attribute that has a value, or -1 when there is none.
// Words after that attribute can be prose and stay outside.
func unfinishedTagEnd(source []byte, start int) int {
	pos := start + 1
	if pos == len(source) || !asciiLetter(source[pos]) {
		return -1
	}
	for pos < len(source) && tagNameByte(source[pos]) {
		pos++
	}
	end := -1
	for {
		next, valued := htmlAttributeEnd(source, pos)
		if next < 0 {
			return end
		}
		pos = next
		if valued {
			end = pos
		}
	}
}

// htmlAttributeEnd reads white space and one attribute at pos. It returns the
// offset after the attribute and whether the attribute has a value, or -1
// when no attribute starts there.
func htmlAttributeEnd(source []byte, pos int) (int, bool) {
	name := skipBlank(source, pos)
	end := htmlNameEnd(source, name)
	if name == pos || end < 0 {
		return -1, false
	}
	equals := skipBlank(source, end)
	if equals == len(source) || source[equals] != '=' {
		return end, false
	}
	if value := htmlValueEnd(source, skipBlank(source, equals+1)); value > 0 {
		return value, true
	}
	return end, false
}

// htmlNameEnd returns the offset after the attribute name at pos, or -1 when
// no name starts there.
func htmlNameEnd(source []byte, pos int) int {
	if pos == len(source) || !asciiLetter(source[pos]) && source[pos] != '_' && source[pos] != ':' {
		return -1
	}
	pos++
	for pos < len(source) && (tagNameByte(source[pos]) || bytes.IndexByte([]byte("_.:"), source[pos]) >= 0) {
		pos++
	}
	return pos
}

// htmlValueEnd returns the offset after the attribute value at pos, quoted or
// not, or -1 when no complete value starts there.
func htmlValueEnd(source []byte, pos int) int {
	if pos == len(source) {
		return -1
	}
	if quote := source[pos]; quote == '"' || quote == '\'' {
		closing := bytes.IndexByte(source[pos+1:], quote)
		if closing < 0 {
			return -1
		}
		return pos + closing + 2
	}
	end := pos
	for end < len(source) && bytes.IndexByte([]byte(" \t\r\n\"'=<>`"), source[end]) < 0 {
		end++
	}
	if end == pos {
		return -1
	}
	return end
}

// codeSpans pairs backtick runs as CommonMark does: a run opens a code span
// when a later run has the same length and is literal otherwise. Outside a
// span, a backslash makes the backtick after it literal. An HTML tag binds
// tighter than a code span, so the text of a possible tag is passed over
// whole; a backtick in an attribute value opens nothing.
func codeSpans(source []byte) []document.Span {
	var spans []document.Span
	for pos := 0; pos < len(source); {
		switch {
		case source[pos] == '<':
			pos = inlineMarkupEnd(source, pos)
		case source[pos] == '`' && !markdownEscaped(source, pos):
			open := backtickRunEnd(source, pos)
			end := closingBackticks(source, open, open-pos)
			if end > 0 {
				spans = append(spans, document.Span{Start: pos, End: end})
			}
			pos = max(open, end)
		default:
			pos++
		}
	}
	return spans
}

func backtickRunEnd(source []byte, pos int) int {
	for pos < len(source) && source[pos] == '`' {
		pos++
	}
	return pos
}

// closingBackticks returns the offset after the first run of exactly length
// backticks at or after from, or -1.
func closingBackticks(source []byte, from, length int) int {
	for pos := from; pos < len(source); pos++ {
		if source[pos] != '`' {
			continue
		}
		end := backtickRunEnd(source, pos)
		if end-pos == length {
			return end
		}
		pos = end - 1
	}
	return -1
}
