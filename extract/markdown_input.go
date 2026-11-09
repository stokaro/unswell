package extract

import (
	"bytes"
	"context"
	"sort"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

type markdownInput struct {
	source   []byte
	inserted []int
	original []int
}

// prepareMarkdown prevents the pinned scanner from treating a pipe-only row as
// a delimiter row. Markers also make zero-width cells visible to the block
// grammar. Mapped framing retains the first block's wrapper and finishes the
// last block at EOF. The block grammar also needs a same-length shadow of
// escapes in list-item link labels; otherwise it rejects the following list
// marker. Only the block parser sees these changes; inline parsing uses
// original bytes.
func prepareMarkdown(ctx context.Context, source []byte) (markdownInput, error) {
	source = maskListLinkEscapes(source)
	var inserted []int
	start := 0
	for line := range bytes.SplitAfterSeq(source, []byte{'\n'}) {
		if err := ctx.Err(); err != nil {
			return markdownInput{}, err
		}
		if emptyMarkdownRow(bytes.TrimRight(line, "\r\n")) {
			first := bytes.IndexByte(line, '|')
			for i := first + 1; i < len(line); i++ {
				if line[i] == '|' {
					inserted = append(inserted, start+i)
				}
			}
		}
		start += len(line)
	}
	if len(inserted) == 0 {
		return (markdownInput{source: source}).framed(), nil
	}
	modified := make([]byte, 0, len(source)+len(inserted))
	start = 0
	for i, offset := range inserted {
		modified = append(modified, source[start:offset]...)
		inserted[i] = len(modified)
		modified = append(modified, 'x')
		start = offset
	}
	modified = append(modified, source[start:]...)
	return (markdownInput{source: modified, inserted: inserted}).framed(), nil
}

// maskListLinkEscapes leaves all source offsets unchanged. The block grammar
// misclassifies a following list marker after escaped punctuation in an inline
// link label. The inline grammar still receives the untouched document source.
func maskListLinkEscapes(source []byte) []byte {
	var masked []byte
	lineStart := 0
	for line := range bytes.SplitAfterSeq(source, []byte{'\n'}) {
		maskListLineLinkEscapes(source, line, lineStart, &masked)
		lineStart += len(line)
	}
	if masked != nil {
		return masked
	}
	return source
}

func maskListLineLinkEscapes(source, line []byte, lineStart int, masked *[]byte) {
	content := markdownListContentStart(line)
	if content < 0 {
		return
	}
	lastCloseParen := bytes.LastIndexByte(line, ')')
	for pos := content; pos < len(line); pos++ {
		if line[pos] != '[' || escapedMarkdownBracket(line, pos) {
			continue
		}
		close := markdownLinkLabelEnd(line, pos)
		if close < 0 {
			return
		}
		if close+1 < lastCloseParen && line[close+1] == '(' {
			maskLinkLabelEscapes(source, line, lineStart, pos+1, close, masked)
		}
		pos = close
	}
}

func maskLinkLabelEscapes(source, line []byte, lineStart, start, end int, masked *[]byte) {
	for pos := start; pos+1 < end; pos++ {
		if line[pos] != '\\' || !asciiPunctuation(line[pos+1]) {
			continue
		}
		if *masked == nil {
			*masked = bytes.Clone(source)
		}
		(*masked)[lineStart+pos], (*masked)[lineStart+pos+1] = 'x', 'x'
		pos++
	}
}

func escapedMarkdownBracket(line []byte, pos int) bool {
	backslashes := 0
	for pos > 0 && line[pos-1] == '\\' {
		backslashes++
		pos--
	}
	return backslashes%2 != 0
}

func markdownListContentStart(line []byte) int {
	i := 0
	for i < len(line) && markdownSpace(line[i]) {
		i++
	}
	i = markdownListMarkerEnd(line, i)
	if i < 0 || i >= len(line) || !markdownSpace(line[i]) {
		return -1
	}
	for i < len(line) && markdownSpace(line[i]) {
		i++
	}
	return i
}

func markdownListMarkerEnd(line []byte, start int) int {
	if start >= len(line) {
		return -1
	}
	if markdownBullet(line[start]) {
		return start + 1
	}
	return markdownOrderedMarkerEnd(line, start)
}

func markdownBullet(char byte) bool {
	switch char {
	case '-', '*', '+':
		return true
	default:
		return false
	}
}

func markdownOrderedMarkerEnd(line []byte, start int) int {
	i := start
	for i < len(line) && i-start < 9 && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == start || i >= len(line) {
		return -1
	}
	if line[i] != '.' && line[i] != ')' {
		return -1
	}
	return i + 1
}

func markdownSpace(char byte) bool { return char == ' ' || char == '\t' }

func markdownLinkLabelEnd(line []byte, open int) int {
	depth := 1
	for i := open + 1; i < len(line); i++ {
		if line[i] == '\\' && i+1 < len(line) {
			i++
			continue
		}
		switch line[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func asciiPunctuation(char byte) bool {
	return char >= '!' && char <= '/' || char >= ':' && char <= '@' ||
		char >= '[' && char <= '`' || char >= '{' && char <= '~'
}

func (m markdownInput) framed() markdownInput {
	if len(m.source) == 0 {
		return m
	}
	const prefix = "x\n\n"
	source := make([]byte, 0, len(prefix)+len(m.source)+1)
	source = append(source, prefix...)
	source = append(source, m.source...)
	inserted := make([]int, 0, len(prefix)+len(m.inserted)+1)
	for offset := range len(prefix) {
		inserted = append(inserted, offset)
	}
	for _, offset := range m.inserted {
		inserted = append(inserted, offset+len(prefix))
	}
	if source[len(source)-1] != '\n' {
		inserted = append(inserted, len(source))
		source = append(source, '\n')
	}
	return markdownInput{source: source, inserted: inserted}
}

func emptyMarkdownRow(line []byte) bool {
	pipes := 0
	for _, char := range line {
		switch char {
		case ' ', '\t':
		case '|':
			pipes++
		case '>':
			if pipes != 0 {
				return false
			}
		default:
			return false
		}
	}
	// A lone pipe is not an empty GFM table row. The grammar must still see it.
	return pipes >= 2
}

func (m markdownInput) span(node *ts.Node) document.Span {
	start, end := int(node.StartByte()), int(node.EndByte())
	span := document.Span{
		Start: start - sort.SearchInts(m.inserted, start),
		End:   end - sort.SearchInts(m.inserted, end),
	}
	if m.original != nil && span.End > span.Start {
		span.Start, span.End = m.original[span.Start], m.original[span.End-1]+1
	}
	return span
}

func (m markdownInput) offset(pos int) int {
	return pos - sort.SearchInts(m.inserted, pos)
}
