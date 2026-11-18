package extract

import (
	"bytes"
	"context"
	"sort"
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

// The repairs in this file change only the bytes that the block grammar sees.
// They run after the grammar rejected the document or built a table wrongly,
// so a document that the grammar accepts is still parsed from the same bytes.
// Inline parsing and every reported span keep using the original source.

// shadowLabelEntities overwrites the entity references in a link label that
// starts the content of a line. The block grammar first reads such a label as
// a link reference definition. With an entity reference inside it, the grammar
// then rejects the list item that holds the paragraph. The label has no part
// in the block structure, so bytes of the same length are enough.
func shadowLabelEntities(source []byte) bool {
	changed := false
	for line := range bytes.SplitAfterSeq(source, []byte{'\n'}) {
		open := markdownLabelStart(line)
		if open < 0 {
			continue
		}
		if end := markdownLinkLabelEnd(line, open); end > open {
			changed = shadowEntities(line[open+1:end]) || changed
		}
	}
	return changed
}

// markdownLabelStart returns the offset of a bracket that follows only quote
// markers, indentation, and one list marker, or -1.
func markdownLabelStart(line []byte) int {
	pos := 0
	for pos < len(line) && (markdownSpace(line[pos]) || line[pos] == '>') {
		pos++
	}
	if content := markdownListContentStart(line[pos:]); content >= 0 {
		pos += content
	}
	if pos < len(line) && line[pos] == '[' {
		return pos
	}
	return -1
}

func shadowEntities(label []byte) bool {
	changed := false
	for pos := 0; pos < len(label); pos++ {
		end := entityEnd(label, pos)
		if end < 0 {
			continue
		}
		for i := pos; i < end; i++ {
			label[i] = 'x'
		}
		changed = true
		pos = end - 1
	}
	return changed
}

// entityEnd returns the offset after a named or numeric character reference
// that starts at start, or -1. It does not check the name: a longer match than
// the grammar's only shadows more of a label.
func entityEnd(label []byte, start int) int {
	if label[start] != '&' {
		return -1
	}
	pos := start + 1
	for pos < len(label) && (label[pos] == '#' || asciiLetter(label[pos]) || label[pos] >= '0' && label[pos] <= '9') {
		pos++
	}
	if pos == start+1 || pos == len(label) || label[pos] != ';' {
		return -1
	}
	return pos + 1
}

// maskOpenHTML rewrites the HTML block that holds a defect as a processing
// instruction of the same length and returns the run it rewrote. Inside a list
// item or a quote, the grammar does not end an HTML block at a blank line when
// more of the container follows, and the next line becomes an error. A
// processing instruction ends on its own last line, so the grammar closes it
// there. The block stays an HTML block with the same lines; container markers,
// white space, and line breaks keep their bytes.
func maskOpenHTML(syntax syntaxTree, defect *ts.Node, source []byte) (document.Span, bool) {
	block := defect.Parent()
	if block == nil || block.Type(syntax.lang) != "html_block" {
		return document.Span{}, false
	}
	text := htmlBlockText(syntax, block, int(defect.StartByte()), source)
	last := len(text) - 1
	// "<?" and "?>" each need two adjacent bytes of one line. A block without
	// room for both, such as a lone "<p>", is left to the caller.
	if last < 3 || text[1] != text[0]+1 || text[last-1] != text[last]-1 || !blankLineEndsHTML(source[text[0]:]) {
		return document.Span{}, false
	}
	// The known defect is the line after the blank line that ends the block.
	// An error at another place in an HTML block has another cause.
	if bytes.Count(source[text[last]:defect.StartByte()], []byte{'\n'}) < 2 {
		return document.Span{}, false
	}
	changed := false
	for index, offset := range text {
		char := instructionByte(index, last, source[offset])
		changed = changed || source[offset] != char
		source[offset] = char
	}
	return document.Span{Start: text[0], End: text[last] + 1}, changed
}

// instructionByte returns the byte at index of a processing instruction that
// takes the place of text with the given last index. White space stays; every
// other byte inside becomes inert, so nothing ends the instruction early.
func instructionByte(index, last int, char byte) byte {
	switch {
	case index == 0:
		return '<'
	case index == 1 || index == last-1:
		return '?'
	case index == last:
		return '>'
	case markdownSpace(char):
		return char
	}
	return 'x'
}

// blankLineEndsHTML reports whether an HTML block with this opening text ends
// at a blank line. A comment, a declaration, a processing instruction, and a
// "pre", "script", "style" or "textarea" element end at their own closing
// text and may hold blank lines, so such a block is never cut at one.
func blankLineEndsHTML(open []byte) bool {
	if len(open) < 2 || open[0] != '<' || open[1] == '!' || open[1] == '?' {
		return false
	}
	for _, name := range []string{"pre", "script", "style", "textarea"} {
		if rest := open[1:]; len(rest) >= len(name) && strings.EqualFold(string(rest[:len(name)]), name) &&
			(len(rest) == len(name) || !tagNameByte(rest[len(name)])) {
			return false
		}
	}
	return true
}

// tagNameByte reports whether an HTML tag name can go on with char.
func tagNameByte(char byte) bool {
	return asciiLetter(char) || char >= '0' && char <= '9' || char == '-'
}

// htmlBlockText lists the offsets of the text of an HTML block before end:
// every byte from the first to the last one that is not white space, without
// line breaks and without the container markers of a continued line.
func htmlBlockText(syntax syntaxTree, block *ts.Node, end int, source []byte) []int {
	var text []int
	pos := int(block.StartByte())
	for i := 0; i < block.ChildCount(); i++ {
		child := block.Child(i)
		if int(child.StartByte()) >= end {
			break
		}
		if child.Type(syntax.lang) != "block_continuation" {
			continue
		}
		text = appendLineText(text, source, pos, int(child.StartByte()))
		pos = max(pos, int(child.EndByte()))
	}
	text = appendLineText(text, source, pos, end)
	for len(text) > 0 && markdownSpace(source[text[len(text)-1]]) {
		text = text[:len(text)-1]
	}
	return text
}

// appendLineText adds the offsets from start to end that are not line breaks.
// White space before the first text byte is left out.
func appendLineText(text []int, source []byte, start, end int) []int {
	for pos := start; pos < end; pos++ {
		if source[pos] == '\r' || source[pos] == '\n' || len(text) == 0 && markdownSpace(source[pos]) {
			continue
		}
		text = append(text, pos)
	}
	return text
}

// rewrittenAsHTML reports whether each rewritten run lies inside one HTML
// block of the tree. A run that the grammar reads as anything else would leak
// its lines into prose.
func rewrittenAsHTML(ctx context.Context, syntax syntaxTree, runs []document.Span) (bool, error) {
	held := 0
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if node.Type(syntax.lang) != "html_block" {
			return false, nil
		}
		block := syntaxSpan(node, 0)
		for _, run := range runs {
			if block.Start <= run.Start && run.End <= block.End {
				held++
			}
		}
		return true, nil
	})
	return held == len(runs), err
}

// fillEmptyCells puts a marker byte between every two adjacent pipes, where a
// table row has a zero-width cell. The grammar builds a row with such a cell
// outside its table. The marker maps to an empty source span, like the markers
// of an empty row, so it yields no prose and moves no offset, in a table and
// anywhere else. Blanked regions are returned in the new coordinates.
func fillEmptyCells(input markdownInput, blanked []document.Span) (markdownInput, []document.Span, bool) {
	var offsets []int
	for pos := 1; pos < len(input.source); pos++ {
		if input.source[pos] == '|' && input.source[pos-1] == '|' {
			offsets = append(offsets, pos)
		}
	}
	if len(offsets) == 0 {
		return input, blanked, false
	}
	moved := make([]document.Span, 0, len(blanked))
	for _, region := range blanked {
		// A blanked region holds no pipe, so no marker lands inside one.
		moved = append(moved, document.Span{
			Start: region.Start + sort.SearchInts(offsets, region.Start),
			End:   region.End + sort.SearchInts(offsets, region.End),
		})
	}
	return input.withMarkers(offsets), moved, true
}

// withMarkers inserts a marker byte before each of the sorted offsets and
// records it as inserted, so spans still map to the original source.
func (m markdownInput) withMarkers(offsets []int) markdownInput {
	source := make([]byte, 0, len(m.source)+len(offsets))
	inserted := make([]int, 0, len(m.inserted)+len(offsets))
	marker, previous := 0, 0
	for pos, char := range m.source {
		if marker < len(offsets) && offsets[marker] == pos {
			inserted = append(inserted, len(source))
			source = append(source, 'x')
			marker++
		}
		if previous < len(m.inserted) && m.inserted[previous] == pos {
			inserted = append(inserted, len(source))
			previous++
		}
		source = append(source, char)
	}
	return markdownInput{source: source, inserted: inserted, original: m.original}
}
