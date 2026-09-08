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
}

// prepareMarkdown prevents the pinned scanner from treating a pipe-only row as
// a delimiter row. Markers also make zero-width cells visible to the block
// grammar. Mapped framing retains the first block's wrapper and finishes the
// last block at EOF. Only the block parser sees these additions; inline parsing
// uses original bytes.
func prepareMarkdown(ctx context.Context, source []byte) (markdownInput, error) {
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
	return document.Span{
		Start: start - sort.SearchInts(m.inserted, start),
		End:   end - sort.SearchInts(m.inserted, end),
	}
}
