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
// grammar. Only the block parser sees them; inline parsing uses original bytes.
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
		return markdownInput{source: source}, nil
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
	return markdownInput{source: modified, inserted: inserted}, nil
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
