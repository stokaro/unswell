package extract

import (
	"bytes"
	"context"
	"maps"
	"slices"

	ts "github.com/odvcencio/gotreesitter"

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
	for _, span := range bracketCodeSpans(source) {
		maskOpenBrackets(source[span.Start:span.End])
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
		rewriteUnfinishedTag(source[pos:end])
		expected[document.Span{Start: pos, End: end}] = "html_tag"
		pos = end - 1
	}
	return changed
}

// rewriteUnfinishedTag overwrites an unfinished tag with one short tag of the
// same length, which the grammar reads as inline HTML.
func rewriteUnfinishedTag(tag []byte) {
	for index := 1; index < len(tag)-1; index++ {
		tag[index] = 'x'
	}
	tag[len(tag)-1] = '>'
}

// protectUnfinishedTags gives a block that the grammar accepts the same
// reading of an unfinished tag as a block that it rejects. The grammar reads
// such a tag as text when it accepts the block, so one tag was prose or HTML
// by its spelling. A "<" with no ">" after it that opens a tag with a quoted
// attribute value is rewritten as in maskInlineDefects. A bracket inside a
// code span, and one without a quoted value, are left alone.
//
// The rewrite adds a closing bracket, and a tag that it completes can take a
// link or another construct around it apart. So the new tree is used only
// when it has an HTML tag at exactly those bytes and reads everything outside
// them as the first tree does.
func protectUnfinishedTags(ctx context.Context, original syntaxTree, source []byte) syntaxTree {
	tags := unfinishedTags(source)
	if len(tags) == 0 {
		return original
	}
	trial := bytes.Clone(source)
	expected := make(map[document.Span]string, len(tags))
	for _, tag := range tags {
		rewriteUnfinishedTag(trial[tag.Start:tag.End])
		expected[tag] = "html_tag"
	}
	repaired, err := parseInlineBoundaries(ctx, trial, expected)
	if err != nil {
		return original
	}
	if !sameReadingOutside(ctx, original, repaired, tags) {
		repaired.tree.Release()
		return original
	}
	original.tree.Release()
	copy(source, trial)
	return repaired
}

// readNode identifies one node that decides what the inline reader emits.
type readNode struct {
	kind string
	span document.Span
}

// sameReadingOutside reports whether two trees of one block hold the same
// nodes that the inline reader acts on, apart from those inside the regions.
func sameReadingOutside(ctx context.Context, first, second syntaxTree, regions []document.Span) bool {
	before, err := readNodesOutside(ctx, first, regions)
	if err != nil {
		return false
	}
	after, err := readNodesOutside(ctx, second, regions)
	return err == nil && maps.Equal(before, after)
}

func readNodesOutside(ctx context.Context, syntax syntaxTree, regions []document.Span) (map[readNode]int, error) {
	kinds := slices.Concat(protectedInline, inlineLinks, inlineDelimiters, inlineReferences, []string{"link_text"})
	nodes := make(map[readNode]int)
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		span := syntaxSpan(node, 0)
		if coveredBy(span, regions) {
			return true, nil
		}
		if kind := node.Type(syntax.lang); slices.Contains(kinds, kind) {
			nodes[readNode{kind: kind, span: span}]++
		}
		return false, nil
	})
	return nodes, err
}

// unfinishedTags returns the unfinished tags with an attribute value that
// stand after the last ">" of the block and outside its code spans.
func unfinishedTags(source []byte) []document.Span {
	start := bytes.LastIndexByte(source, '>') + 1
	if bytes.IndexByte(source[start:], '<') < 0 {
		return nil
	}
	code := codeSpans(source)
	var tags []document.Span
	for pos := start; pos < len(source); pos++ {
		if source[pos] != '<' || slices.ContainsFunc(code, func(span document.Span) bool { return span.Start < pos && pos < span.End }) {
			continue
		}
		if end := unfinishedTagEnd(source, pos); end > 0 {
			tags = append(tags, document.Span{Start: pos, End: end})
			pos = end - 1
		}
	}
	return tags
}

// unfinishedTagEnd reads an HTML open tag without its closing bracket at
// start: a name and attributes as CommonMark defines them. It returns the
// offset after the last attribute that has a quoted value, or -1 when there is
// none. A quoted value marks HTML source; without one the text can be prose,
// as in "n <m and k=2". Words after that attribute stay outside too.
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
		next, quoted := htmlAttributeEnd(source, pos)
		if next < 0 {
			return end
		}
		pos = next
		if quoted {
			end = pos
		}
	}
}

// htmlAttributeEnd reads white space and one attribute at pos. It returns the
// offset after the attribute and whether the attribute has a quoted value, or
// -1 when no attribute starts there.
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
	if value, quoted := htmlValueEnd(source, skipBlank(source, equals+1)); value > 0 {
		return value, quoted
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

// htmlValueEnd returns the offset after the attribute value at pos and
// whether the value is quoted, or -1 when no complete value starts there.
func htmlValueEnd(source []byte, pos int) (int, bool) {
	if pos == len(source) {
		return -1, false
	}
	if quote := source[pos]; quote == '"' || quote == '\'' {
		closing := bytes.IndexByte(source[pos+1:], quote)
		if closing < 0 {
			return -1, false
		}
		return pos + closing + 2, true
	}
	end := pos
	for end < len(source) && bytes.IndexByte([]byte(" \t\r\n\"'=<>`"), source[end]) < 0 {
		end++
	}
	if end == pos {
		return -1, false
	}
	return end, false
}

// restoreCodeSpans repairs a tree that the grammar built without an error and
// without a code span that CommonMark has. With a "<" inside the span the
// grammar can drop the span silently: its words become prose, and its
// backticks can pair with those of the next span. Such a span gets the same
// repair as in a rejected block. The repaired tree is used only when it has a
// code span at each of those spans; otherwise the reading stays as it was. A
// block whose tree has every such span is left alone.
func restoreCodeSpans(ctx context.Context, original syntaxTree, source []byte) syntaxTree {
	lost := lostCodeSpans(ctx, original, source)
	if len(lost) == 0 {
		return original
	}
	trial := bytes.Clone(source)
	expected := make(map[document.Span]string, len(lost))
	for _, span := range lost {
		maskOpenBrackets(trial[span.Start:span.End])
		expected[span] = "code_span"
	}
	repaired, err := parseInlineBoundaries(ctx, trial, expected)
	if err != nil {
		return original
	}
	original.tree.Release()
	copy(source, trial)
	return repaired
}

// lostCodeSpans returns the code spans that hold a "<" and that the tree does
// not have as a code span node at the same bytes.
func lostCodeSpans(ctx context.Context, syntax syntaxTree, source []byte) []document.Span {
	paired := bracketCodeSpans(source)
	if len(paired) == 0 {
		return nil
	}
	found := make(map[document.Span]bool)
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if node.Type(syntax.lang) != "code_span" {
			return false, nil
		}
		found[syntaxSpan(node, 0)] = true
		return true, nil
	})
	if err != nil && ctx.Err() != nil {
		return nil
	}
	// A tree that is too deep to walk hides its spans: the grammar nests one
	// level for each inline element before a span that it dropped. The spans
	// not seen count as lost, and the repaired tree decides.
	return slices.DeleteFunc(paired, func(span document.Span) bool { return found[span] })
}

// bracketCodeSpans returns the code spans of codeSpans that hold a "<".
func bracketCodeSpans(source []byte) []document.Span {
	if bytes.IndexByte(source, '`') < 0 || bytes.IndexByte(source, '<') < 0 {
		return nil
	}
	return slices.DeleteFunc(codeSpans(source), func(span document.Span) bool {
		return bytes.IndexByte(source[span.Start:span.End], '<') < 0
	})
}

// maskOpenBrackets overwrites each "<" of a code span. A ">" stays, because it
// may end markup that began before the span.
func maskOpenBrackets(code []byte) {
	for index, char := range code {
		if char == '<' {
			code[index] = 'x'
		}
	}
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
