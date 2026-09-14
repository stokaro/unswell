package extract

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/stokaro/unswell/document"
)

// mdxMasker removes only syntax owned by MDX. Candidate boundaries are checked
// with the JavaScript/JSX grammar, so braces and angle brackets in literals,
// comments, regular expressions, and nested expressions cannot end a region.
// Markdown still owns the prose and its source map. No component is evaluated.
type mdxMasker struct {
	ctx       context.Context
	doc       *document.Document
	source    []byte
	protected []document.Span
	stack     []string
	remaining int
}

func maskMDX(ctx context.Context, doc *document.Document, source []byte) ([]document.Span, error) {
	m := mdxMasker{ctx: ctx, doc: doc, source: source, remaining: 16*len(source) + 65536}
	for pos := 0; pos < len(source); {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		next, err := m.region(pos)
		if err != nil {
			return nil, fmt.Errorf("parse mdx at byte %d: %w", pos, err)
		}
		pos = next
	}
	if len(m.stack) != 0 {
		return nil, fmt.Errorf("parse mdx: unclosed JSX element %q", m.stack[len(m.stack)-1])
	}
	return m.protected, nil
}

func (m *mdxMasker) region(pos int) (int, error) {
	if m.source[pos] == '\\' {
		return min(pos+2, len(m.source)), nil
	}
	if end := mdxCodeEnd(m.source, pos); end > pos {
		m.exclude(pos, end, "code", false)
		return end, nil
	}
	if m.source[pos] == '`' {
		return mdxCodeSpanEnd(m.source, pos), nil
	}
	kind, delimiter := "", byte(0)
	switch {
	case mdxESMStart(m.source, pos):
		kind, delimiter = "esm", '\n'
	case m.source[pos] == '{':
		kind, delimiter = "expression", '}'
	case m.source[pos] == '<':
		kind, delimiter = "tag", '>'
	default:
		return pos + 1, nil
	}
	end, err := m.syntaxEnd(pos, delimiter, kind)
	if err != nil {
		return pos, err
	}
	m.exclude(pos, end, "mdx-"+kind, !mdxFlow(m.source, pos, end))
	return end, nil
}

func (m *mdxMasker) exclude(start, end int, reason string, inline bool) {
	span := document.Span{Start: start, End: end}
	m.doc.Excluded = append(m.doc.Excluded, document.Exclusion{Span: span, Reason: reason})
	m.protected = append(m.protected, span)
	maskRange(m.source, start, end)
	if inline {
		// A leading inline tag must not turn its following prose into an
		// indented code block. The marker is excluded again during inline parsing.
		for pos := start; pos < end; pos++ {
			if m.source[pos] == ' ' && (pos == start || m.source[pos-1] == '\n') {
				m.source[pos] = 'x'
			}
		}
	}
}

func mdxFlow(source []byte, start, end int) bool {
	lineStart := bytes.LastIndexByte(source[:start], '\n') + 1
	lineEnd := end + bytes.IndexByte(source[end:], '\n')
	if lineEnd < end {
		lineEnd = len(source)
	}
	return len(bytes.TrimSpace(source[lineStart:start])) == 0 && len(bytes.TrimSpace(source[end:lineEnd])) == 0
}

func mdxESMStart(source []byte, pos int) bool {
	if pos > 0 && source[pos-1] != '\n' {
		return false
	}
	if pos > 0 {
		previous := bytes.LastIndexByte(source[:pos-1], '\n') + 1
		if len(bytes.TrimSpace(source[previous:pos])) != 0 {
			return false
		}
	}
	for _, keyword := range []string{"import", "export"} {
		if bytes.HasPrefix(source[pos:], []byte(keyword)) && pos+len(keyword) < len(source) &&
			strings.ContainsRune(" \t\r\n", rune(source[pos+len(keyword)])) {
			return true
		}
	}
	return false
}

// trimMDXTags leaves a visible block's first token at its start even when a
// component wraps it. Interior tags and all dynamic expressions stay boundaries.
func trimMDXTags(source []byte, mapped document.MappedText) document.MappedText {
	edge := func(pos int) bool {
		char := mapped.Text[pos]
		if char == 0 {
			return source[mapped.Map[pos].Start] == '<'
		}
		return char == ' ' || char == '\t' || char == '\r' || char == '\n'
	}
	start, end := 0, len(mapped.Text)
	for start < end && edge(start) {
		start++
	}
	for end > start && edge(end-1) {
		end--
	}
	return document.MappedText{Text: mapped.Text[start:end], Map: mapped.Map[start:end]}
}
