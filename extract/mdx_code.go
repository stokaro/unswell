package extract

import (
	"bytes"
	"slices"
	"strings"
)

// mdxCodeEnd keeps MDX-looking examples inside a fenced code block opaque.
// This boundary pass also sees fences inside JSX, where a Markdown-only tree
// would otherwise consume the entire component as an HTML block.
func mdxCodeEnd(source []byte, start int) int {
	marker := source[start]
	if marker != '`' && marker != '~' {
		return start
	}
	lineStart := bytes.LastIndexByte(source[:start], '\n') + 1
	prefix := strings.TrimLeft(string(source[lineStart:start]), " \t>")
	if !slices.Contains([]string{"", "- ", "* ", "+ "}, prefix) {
		return start
	}
	count := mdxRun(source, start, marker)
	if count < 3 {
		return start
	}
	lineEnd := bytes.IndexByte(source[start:], '\n')
	if lineEnd < 0 {
		return len(source)
	}
	if marker == '`' && bytes.IndexByte(source[start+count:start+lineEnd], '`') >= 0 {
		return start
	}
	return mdxFenceEnd(source, start+lineEnd+1, marker, count)
}

func mdxFenceEnd(source []byte, pos int, marker byte, count int) int {
	for pos < len(source) {
		line, _, _ := bytes.Cut(source[pos:], []byte{'\n'})
		trimmed := bytes.TrimLeft(line, " \t>")
		if len(trimmed) > 0 && trimmed[0] == marker {
			size := mdxRun(trimmed, 0, marker)
			if size >= count && len(bytes.TrimSpace(trimmed[size:])) == 0 {
				return min(pos+len(line)+1, len(source))
			}
		}
		pos += len(line) + 1
	}
	return len(source)
}

func mdxCodeSpanEnd(source []byte, start int) int {
	count := mdxRun(source, start, '`')
	for pos := start + count; pos < len(source); {
		next := bytes.IndexByte(source[pos:], '`')
		if next < 0 {
			break
		}
		pos += next
		size := mdxRun(source, pos, '`')
		pos += size
		if size == count {
			return pos
		}
	}
	return start + count
}

func mdxRun(source []byte, start int, marker byte) int {
	end := start
	for end < len(source) && source[end] == marker {
		end++
	}
	return end - start
}
