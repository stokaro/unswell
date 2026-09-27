package extract

import (
	"bytes"
	"fmt"
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

func (r *sourceReader) csharpLiteral(node *ts.Node, text, kind string) (literalSpec, error) {
	quote := strings.IndexByte(text, '"')
	if quote < 0 {
		return literalSpec{}, fmt.Errorf("csharp string at byte %d has no delimiter", node.StartByte())
	}
	width := len(text[quote:]) - len(strings.TrimLeft(text[quote:], "\""))
	if kind != "raw_string_literal" && (kind != "interpolated_string_expression" || width < 3) {
		spec, err := r.quotedLiteral(node, text, kind)
		if strings.Contains(text[:quote], "@") {
			spec.mode = "csharp-verbatim"
		}
		return spec, err
	}
	span := syntaxSpan(node, 0)
	span.Start += quote + width
	span.End -= width
	return csharpRawLiteral(r.doc.Source, span)
}

func csharpRawLiteral(source []byte, span document.Span) (literalSpec, error) {
	if span.End < span.Start {
		return literalSpec{}, fmt.Errorf("csharp raw string has mismatched delimiters")
	}
	spec := literalSpec{span: span, mode: "raw"}
	content := string(source[span.Start:span.End])
	first := strings.IndexByte(content, '\n')
	if first < 0 || strings.TrimSpace(content[:first]) != "" {
		return spec, nil
	}
	last := strings.LastIndexByte(content, '\n')
	spec.indent = content[last+1:]
	if strings.Trim(spec.indent, " \t") != "" {
		return literalSpec{}, fmt.Errorf("csharp raw closing delimiter must occupy its own line")
	}
	spec.span.Start += first + 1
	spec.span.End = max(spec.span.Start, span.Start+last)
	if spec.span.End > spec.span.Start && source[spec.span.End-1] == '\r' {
		spec.span.End--
	}
	return spec, nil
}

func literalSegment(builder *mapping.Builder, source []byte, start, end int, spec literalSpec) error {
	if spec.indent == "" {
		return literalSource(builder, source, start, end, spec.mode)
	}
	for start < end {
		lineEnd := end
		if newline := bytes.IndexByte(source[start:end], '\n'); newline >= 0 {
			lineEnd = start + newline + 1
		}
		if start == spec.span.Start || source[start-1] == '\n' {
			trimmed, err := csharpIndent(source[start:lineEnd], spec.indent)
			if err != nil {
				return err
			}
			start += trimmed
		}
		if err := literalSource(builder, source, start, lineEnd, spec.mode); err != nil {
			return err
		}
		start = lineEnd
	}
	return nil
}

func csharpIndent(line []byte, indent string) (int, error) {
	if bytes.HasPrefix(line, []byte(indent)) {
		return len(indent), nil
	}
	blank := strings.TrimRight(string(line), "\r\n")
	if strings.Trim(blank, " \t") == "" && strings.HasPrefix(indent, blank) {
		return len(blank), nil
	}
	return 0, fmt.Errorf("csharp raw string indentation does not match its closing delimiter")
}
