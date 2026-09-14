package extract

import (
	"bytes"
	"context"
	"fmt"

	ts "github.com/stokaro/gotreesitter"

	"github.com/stokaro/unswell/document"
)

// prepareMDXMarkdown removes indentation only where the Markdown grammar would
// create indented code. MDX permits indented prose and uses fences for code.
// The offset table composes this removal with Markdown's inserted framing.
// List indentation remains owned by the Markdown grammar.
func prepareMDXMarkdown(ctx context.Context, source []byte) (markdownInput, error) {
	var original []int
	for range 16 {
		input, err := prepareMarkdown(ctx, source)
		if err != nil {
			return markdownInput{}, err
		}
		removed, err := mdxIndentation(ctx, input)
		if err != nil {
			return markdownInput{}, err
		}
		if len(removed) == 0 {
			input.original = original
			return input, nil
		}
		source, original = mdxDedent(source, original, removed)
	}
	return markdownInput{}, fmt.Errorf("MDX indentation exceeds 16 normalization passes")
}

func mdxIndentation(ctx context.Context, input markdownInput) ([]document.Span, error) {
	syntax, err := parseSyntax(ctx, input.source, "markdown")
	if err != nil {
		return nil, err
	}
	defer syntax.tree.Release()
	var removed []document.Span
	err = walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if node.Type(syntax.lang) != "indented_code_block" {
			return false, nil
		}
		start, end := int(node.StartByte()), int(node.EndByte())
		start = bytes.LastIndexByte(input.source[:start], '\n') + 1
		for pos := start; pos < end; {
			begin := mdxIndentStart(input.source, pos)
			stop := mdxIndentEnd(input.source, begin)
			// The temporary input has framing but no original-offset table yet.
			if stop > begin {
				removed = append(removed, document.Span{
					Start: input.offset(begin), End: input.offset(stop)})
			}
			newline := bytes.IndexByte(input.source[stop:], '\n')
			if newline < 0 {
				break
			}
			pos = stop + newline + 1
		}
		return true, nil
	})
	return removed, err
}

func mdxIndentEnd(source []byte, start int) int {
	end, columns := start, 0
	for end < len(source) && columns < 4 {
		switch source[end] {
		case ' ':
			columns++
		case '\t':
			columns += 4 - columns%4
		default:
			return end
		}
		end++
	}
	return end
}

func mdxDedent(source []byte, original []int, removed []document.Span) ([]byte, []int) {
	result := make([]byte, 0, len(source))
	positions := make([]int, 0, len(source))
	region := 0
	for pos, char := range source {
		for region < len(removed) && removed[region].End <= pos {
			region++
		}
		if region < len(removed) && pos >= removed[region].Start {
			continue
		}
		result = append(result, char)
		if original == nil {
			positions = append(positions, pos)
		} else {
			positions = append(positions, original[pos])
		}
	}
	return result, positions
}

func mdxIndentStart(source []byte, start int) int {
	for {
		next := start
		for next < len(source) && (source[next] == ' ' || source[next] == '\t') {
			next++
		}
		if next == len(source) || source[next] != '>' {
			return start
		}
		start = next + 1
		if start < len(source) && (source[start] == ' ' || source[start] == '\t') {
			start++
		}
	}
}
