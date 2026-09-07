package extract

import (
	"bytes"
	"context"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

func markdown(ctx context.Context, doc *document.Document, options Options) error {
	source := maskFrontMatter(doc)
	parser := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser()
	root := parser.Parse(text.NewReader(source))
	depth := 0
	return ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			depth--
			return ast.WalkContinue, nil
		}
		if err := ctx.Err(); err != nil {
			return ast.WalkStop, err
		}
		depth++
		if depth > 128 {
			return ast.WalkStop, fmt.Errorf("markdown nesting exceeds 128")
		}
		return extractNode(doc, node, options)
	})
}

func extractNode(doc *document.Document, node ast.Node, options Options) (ast.WalkStatus, error) {
	if reason := excludedNode(node, options); reason != "" {
		if blockSuppression(doc, node, reason) {
			return ast.WalkStop, fmt.Errorf("suppression directives are not implemented in this alpha")
		}
		recordExcluded(doc, node, reason)
		return ast.WalkSkipChildren, nil
	}
	if kind := proseKind(node); kind != "" {
		mapped, err := inlineText(doc, node)
		if err != nil {
			return ast.WalkStop, err
		}
		appendBlock(doc, mapped, kind)
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func maskFrontMatter(doc *document.Document) []byte {
	source := bytes.Clone(doc.Source)
	start := 0
	if bytes.HasPrefix(source, []byte{0xef, 0xbb, 0xbf}) {
		start = 3
		copy(source[:3], "   ")
	}
	if !bytes.HasPrefix(source[start:], []byte("---\n")) && !bytes.HasPrefix(source[start:], []byte("---\r\n")) {
		return source
	}
	end := frontMatterEnd(source, start)
	if end == 0 {
		return source
	}
	doc.Excluded = append(doc.Excluded, document.Exclusion{Span: document.Span{Start: start, End: end}, Reason: "front-matter"})
	maskRange(source, start, end)
	return source
}

func frontMatterEnd(source []byte, start int) int {
	pos := start + bytes.IndexByte(source[start:], '\n') + 1
	for pos < len(source) {
		end := bytes.IndexByte(source[pos:], '\n')
		if end < 0 {
			end = len(source) - pos
		}
		line := bytes.TrimSpace(source[pos : pos+end])
		pos = min(len(source), pos+end+1)
		if bytes.Equal(line, []byte("---")) || bytes.Equal(line, []byte("...")) {
			return pos
		}
	}
	return 0
}

func maskRange(source []byte, start, end int) {
	for i := start; i < end; i++ {
		if source[i] != '\n' && source[i] != '\r' {
			source[i] = ' '
		}
	}
}

func excludedNode(node ast.Node, options Options) string {
	switch node.Kind() {
	case ast.KindCodeBlock, ast.KindFencedCodeBlock:
		return "code"
	case ast.KindHTMLBlock:
		return "html"
	case ast.KindBlockquote:
		if !options.IncludeQuotes {
			return "quote"
		}
	}
	return ""
}

func proseKind(node ast.Node) string {
	switch node.Kind() {
	case ast.KindHeading:
		return "heading"
	case extast.KindTableCell:
		return "table-cell"
	case ast.KindParagraph, ast.KindTextBlock:
		for parent := node.Parent(); parent != nil; parent = parent.Parent() {
			if parent.Kind() == ast.KindListItem {
				return "list-item"
			}
		}
		return "paragraph"
	}
	return ""
}

func recordExcluded(doc *document.Document, node ast.Node, reason string) {
	span := nodeBounds(node)
	if span.Valid(len(doc.Source)) {
		doc.Excluded = append(doc.Excluded, document.Exclusion{Span: span, Reason: reason})
	}
}

func nodeBounds(node ast.Node) document.Span {
	start, end := node.Pos(), node.Pos()
	if node.Type() != ast.TypeInline && node.Lines().Len() > 0 {
		start = node.Lines().At(0).Start
		end = node.Lines().At(node.Lines().Len() - 1).Stop
	}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		span := nodeBounds(child)
		if start < 0 || span.Start < start {
			start = span.Start
		}
		end = max(end, span.End)
	}
	if value, ok := node.(*ast.Text); ok {
		start, end = value.Segment.Start, value.Segment.Stop
	}
	return document.Span{Start: start, End: end}
}

func inlineText(doc *document.Document, root ast.Node) (document.MappedText, error) {
	var builder mapping.Builder
	err := ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch value := node.(type) {
		case *ast.CodeSpan, *ast.AutoLink, *ast.RawHTML, *ast.Image:
			if inlineSuppression(doc, node) {
				return ast.WalkStop, fmt.Errorf("suppression directives are not implemented in this alpha")
			}
			span := nodeBounds(node)
			if !span.Valid(len(doc.Source)) {
				span = document.Span{Start: max(0, node.Pos()), End: min(len(doc.Source), max(0, node.Pos())+1)}
			}
			builder.Add(" \x00 ", span)
			recordExcluded(doc, node, "inline-protected")
			return ast.WalkSkipChildren, nil
		case *ast.Text:
			appendInlineText(doc, &builder, value)
		case *ast.String:
			return ast.WalkStop, fmt.Errorf("unsupported synthetic Markdown text at byte %d", node.Pos())
		}
		return ast.WalkContinue, nil
	})
	return builder.Build(), err
}

func appendInlineText(doc *document.Document, builder *mapping.Builder, value *ast.Text) {
	builder.Source(doc.Source, value.Segment.Start, value.Segment.Stop, true)
	if !value.SoftLineBreak() && !value.HardLineBreak() {
		return
	}
	pos := value.Segment.Stop
	if pos < len(doc.Source) {
		builder.Add(" ", document.Span{Start: pos, End: pos + 1})
	}
}

func blockSuppression(doc *document.Document, node ast.Node, reason string) bool {
	span := nodeBounds(node)
	return reason == "html" && span.Valid(len(doc.Source)) && unsupportedSuppression(string(doc.Source[span.Start:span.End]))
}

func inlineSuppression(doc *document.Document, node ast.Node) bool {
	raw, ok := node.(*ast.RawHTML)
	if !ok {
		return false
	}
	var content bytes.Buffer
	for i := 0; i < raw.Segments.Len(); i++ {
		segment := raw.Segments.At(i)
		content.Write(segment.Value(doc.Source))
	}
	return unsupportedSuppression(content.String())
}
