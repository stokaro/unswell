package builtin

import (
	"bytes"
	"fmt"
	"slices"
	"sort"

	"github.com/stokaro/unswell/document"
)

// detailRanges holds only matched, standalone Markdown HTML details boundaries.
// Only the first tag of a grammar-excluded HTML block is considered, so code
// examples, comments, and tags nested in other raw HTML cannot create a scope.
type detailRanges []detailRange

type detailRange struct {
	start  int
	end    int
	parent int
}

type detailBoundary struct {
	pos  int
	open bool
}

func repetitionDetails(doc *document.Document) detailRanges {
	if doc.Format != document.Markdown && doc.Format != document.MDX {
		return nil
	}
	boundaries := detailsBoundaries(doc)
	slices.SortFunc(boundaries, func(a, b detailBoundary) int { return a.pos - b.pos })
	return pairDetails(boundaries)
}

func detailsBoundaries(doc *document.Document) []detailBoundary {
	var boundaries []detailBoundary
	for _, excluded := range doc.Excluded {
		if excluded.Reason != "html" || excluded.Span.Start < 0 || excluded.Span.End > len(doc.Source) ||
			excluded.Span.Start >= excluded.Span.End {
			continue
		}
		raw := doc.Source[excluded.Span.Start:excluded.Span.End]
		trimmed := bytes.TrimLeft(raw, " \t\r\n")
		line, _, _ := bytes.Cut(trimmed, []byte{'\n'})
		kind := standaloneDetailsTag(line)
		if kind == 0 {
			continue
		}
		pos := excluded.Span.Start + len(raw) - len(trimmed)
		if kind > 0 {
			pos += len(line)
		}
		boundaries = append(boundaries, detailBoundary{pos: pos, open: kind > 0})
	}
	return boundaries
}

func pairDetails(boundaries []detailBoundary) detailRanges {
	var ranges detailRanges
	var stack []int
	for _, boundary := range boundaries {
		if boundary.open {
			parent := -1
			if len(stack) > 0 {
				parent = stack[len(stack)-1]
			}
			ranges = append(ranges, detailRange{start: boundary.pos, parent: parent})
			stack = append(stack, len(ranges)-1)
			continue
		}
		if len(stack) == 0 {
			continue
		}
		index := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		ranges[index].end = boundary.pos
	}
	return ranges
}

func (ranges detailRanges) suffix(span document.Span) string {
	if detail := ranges.at(span); detail != 0 {
		return fmt.Sprintf("/details:%d", detail)
	}
	return ""
}

// at returns a distinct comparison owner for the innermost complete details
// element containing the whole prose block. An unmatched tag does not scope text.
func (ranges detailRanges) at(span document.Span) int {
	index := sort.Search(len(ranges), func(i int) bool { return ranges[i].start > span.Start }) - 1
	for index >= 0 {
		r := ranges[index]
		if r.end > r.start && span.End <= r.end {
			return index + 1
		}
		index = r.parent
	}
	return 0
}

// standaloneDetailsTag accepts only a complete tag on its own line. Other HTML
// is left to the Markdown grammar and does not affect repetition comparisons.
func standaloneDetailsTag(line []byte) int {
	line = bytes.TrimSpace(line)
	if bytes.EqualFold(line, []byte("</details>")) {
		return -1
	}
	const prefix = "<details"
	if len(line) < len(prefix)+1 || !bytes.EqualFold(line[:len(prefix)], []byte(prefix)) {
		return 0
	}
	if line[len(prefix)] != '>' && line[len(prefix)] != ' ' && line[len(prefix)] != '\t' {
		return 0
	}
	end := bytes.IndexByte(line, '>')
	if end < 0 || end != len(line)-1 || line[end-1] == '/' {
		return 0
	}
	return 1
}
