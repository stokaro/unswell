package extract

import (
	"bytes"
	"context"
	"errors"
	"slices"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

// Recovery bounds. A document that needs more than this is reported as a
// parse failure, as before.
const (
	maxUnparsedRegions = 8
	maxRecoveryParses  = 24
)

// markdownFramePrefix is the length of the framing that markdownInput.framed
// puts before the document. Recovery never blanks it.
const markdownFramePrefix = 3

// parseMarkdownBlocks parses the block structure of a prepared document. When
// the grammar rejects it, the run of lines at the first defect is replaced by
// line breaks of the same length and the document is parsed again. Offsets stay
// valid. The blanked runs are returned in prepared coordinates so the caller
// can record them as exclusions; nothing in them is extracted as prose.
func parseMarkdownBlocks(ctx context.Context, input markdownInput) (syntaxTree, markdownInput, []document.Span, error) {
	syntax, err := parseSyntax(ctx, input.source, "markdown")
	if err == nil || !errors.Is(err, errInvalidSyntax) {
		return syntax, input, nil, err
	}
	source := slices.Clone(input.source)
	var blanked []document.Span
	for parses := 0; parses < maxRecoveryParses && len(blanked) <= maxUnparsedRegions; parses++ {
		position, found, locateErr := firstSyntaxDefect(ctx, source)
		if locateErr != nil || !found {
			return syntaxTree{}, input, nil, err
		}
		region, ok := nextRegion(source, position, blanked)
		if !ok {
			return syntaxTree{}, input, nil, err
		}
		blankRegion(source, region)
		blanked = mergeRegions(blanked, region)
		recovered, parseErr := parseSyntax(ctx, source, "markdown")
		if parseErr == nil {
			input.source = source
			return recovered, input, blanked, nil
		}
		if !errors.Is(parseErr, errInvalidSyntax) {
			return syntaxTree{}, input, nil, parseErr
		}
	}
	return syntaxTree{}, input, nil, err
}

// orphanedDelimiterError reports a pipe token outside a table row that the
// block walk reached. Its position is in prepared coordinates.
type orphanedDelimiterError struct {
	position int
}

func (*orphanedDelimiterError) Error() string {
	return "markdown grammar returned an orphaned table delimiter"
}

// blankAt blanks the run of lines at position in a copy of the prepared input.
func blankAt(input markdownInput, position int, blanked []document.Span) (markdownInput, []document.Span, bool) {
	region, ok := nextRegion(input.source, position, blanked)
	if !ok {
		return input, blanked, false
	}
	source := slices.Clone(input.source)
	blankRegion(source, region)
	input.source = source
	return input, mergeRegions(blanked, region), true
}

// firstSyntaxDefect returns the start of the first error or missing node.
func firstSyntaxDefect(ctx context.Context, source []byte) (int, bool, error) {
	loose, err := parseSyntaxTree(ctx, source, "markdown")
	if err != nil {
		return 0, false, err
	}
	defer loose.tree.Release()
	position, found := -1, false
	err = walkSyntax(ctx, loose.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if found {
			return true, nil
		}
		if node.IsError() || node.IsMissing() {
			position, found = int(node.StartByte()), true
			return true, nil
		}
		return !node.HasErrorOrMissing(), nil
	})
	return position, found, err
}

// nextRegion picks the run of nonblank lines at position. When that run is
// already blank, it takes the next run after it, so recovery grows forward.
// A run that opens a fenced block without closing it extends to the closing
// fence: blanking half of a fence would turn prose into code.
func nextRegion(source []byte, position int, blanked []document.Span) (document.Span, bool) {
	position = max(position, markdownFramePrefix)
	for position < len(source) {
		start, end := lineRun(source, position)
		// The frame prefix stays; everything else in the run goes.
		start = max(start, markdownFramePrefix)
		if end <= start {
			return document.Span{}, false
		}
		region := closeFences(source, document.Span{Start: start, End: end})
		if !coveredBy(region, blanked) && len(bytes.TrimSpace(source[region.Start:region.End])) > 0 {
			return region, true
		}
		position = skipBlank(source, region.End)
	}
	return document.Span{}, false
}

// skipBlank returns the offset of the next byte that is not white space.
func skipBlank(source []byte, position int) int {
	for position < len(source) && bytes.IndexByte([]byte(" \t\r\n"), source[position]) >= 0 {
		position++
	}
	return position
}

// lineRun returns the maximal run of nonblank lines that contains position.
func lineRun(source []byte, position int) (int, int) {
	position = min(position, len(source)-1)
	start := lineStart(source, position)
	for start > 0 {
		previous := lineStart(source, start-1)
		if len(bytes.TrimSpace(source[previous:start])) == 0 {
			break
		}
		start = previous
	}
	end := lineEnd(source, position)
	for end < len(source) {
		next := lineEnd(source, end)
		if len(bytes.TrimSpace(source[end:next])) == 0 {
			break
		}
		end = next
	}
	return start, end
}

func lineStart(source []byte, position int) int {
	if index := bytes.LastIndexByte(source[:position], '\n'); index >= 0 {
		return index + 1
	}
	return 0
}

// lineEnd returns the offset after the newline that ends the line at position.
func lineEnd(source []byte, position int) int {
	if index := bytes.IndexByte(source[position:], '\n'); index >= 0 {
		return position + index + 1
	}
	return len(source)
}

// closeFences extends a region with an odd number of fence lines to the next
// fence line, so a fenced block is blanked whole.
func closeFences(source []byte, region document.Span) document.Span {
	fences := 0
	for line := range bytes.SplitSeq(source[region.Start:region.End], []byte{'\n'}) {
		if fenceLine(line) {
			fences++
		}
	}
	if fences%2 == 0 {
		return region
	}
	position := region.End
	for position < len(source) {
		end := lineEnd(source, position)
		if fenceLine(source[position:end]) {
			region.End = end
			return region
		}
		position = end
	}
	region.End = len(source)
	return region
}

func fenceLine(line []byte) bool {
	trimmed := bytes.TrimLeft(line, " \t>")
	return bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~"))
}

// blankRegion turns every byte of the region into a line break. Offsets stay
// valid. Lines of spaces would still be indented content for the grammar;
// empty lines are not.
func blankRegion(source []byte, region document.Span) {
	for index := region.Start; index < region.End; index++ {
		source[index] = '\n'
	}
}

func coveredBy(region document.Span, blanked []document.Span) bool {
	return slices.ContainsFunc(blanked, func(done document.Span) bool {
		return done.Start <= region.Start && region.End <= done.End
	})
}

// mergeRegions keeps blanked regions sorted and joins those that touch.
func mergeRegions(blanked []document.Span, region document.Span) []document.Span {
	blanked = append(blanked, region)
	slices.SortFunc(blanked, func(a, b document.Span) int { return a.Start - b.Start })
	merged := blanked[:1]
	for _, next := range blanked[1:] {
		last := &merged[len(merged)-1]
		if next.Start <= last.End {
			last.End = max(last.End, next.End)
			continue
		}
		merged = append(merged, next)
	}
	return merged
}
