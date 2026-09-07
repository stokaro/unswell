package extract

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

type markdownInlineReader struct {
	ctx     context.Context
	doc     *document.Document
	syntax  syntaxTree
	offset  int
	skipped []document.Span
	builder mapping.Builder
}

func markdownInline(ctx context.Context, doc *document.Document, span document.Span, skipped []document.Span) (document.MappedText, error) {
	source := bytes.Clone(doc.Source[span.Start:span.End])
	for _, skip := range skipped {
		maskRange(source, skip.Start-span.Start, skip.End-span.Start)
	}
	syntax, err := parseSyntax(ctx, source, "markdown_inline")
	if err != nil {
		return document.MappedText{}, err
	}
	defer syntax.tree.Release()
	reader := markdownInlineReader{ctx: ctx, doc: doc, syntax: syntax, offset: span.Start, skipped: skipped}
	err = reader.read(syntax.tree.RootNode(), 0)
	return reader.builder.Build(), err
}

func (r *markdownInlineReader) read(node *ts.Node, depth int) error {
	if err := r.ctx.Err(); err != nil {
		return err
	}
	if depth > 128 {
		return fmt.Errorf("inline syntax nesting exceeds 128")
	}
	kind := node.Type(r.syntax.lang)
	span := syntaxSpan(node, r.offset)
	if slices.Contains([]string{"code_span", "image", "uri_autolink", "email_autolink", "html_tag"}, kind) {
		return r.protect(span, kind)
	}
	if slices.Contains([]string{"inline_link", "full_reference_link", "collapsed_reference_link", "shortcut_link"}, kind) {
		return r.linkText(node, depth)
	}
	if kind == "emphasis_delimiter" || kind == "strikethrough_delimiter" {
		return nil
	}
	if kind == "entity_reference" || kind == "numeric_character_reference" || kind == "backslash_escape" {
		r.builder.Source(r.doc.Source, span.Start, span.End, true)
		return nil
	}
	return r.children(node, depth)
}

func (r *markdownInlineReader) children(node *ts.Node, depth int) error {
	span := syntaxSpan(node, r.offset)
	pos := span.Start
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		childSpan := syntaxSpan(child, r.offset)
		r.source(pos, childSpan.Start)
		if err := r.read(child, depth+1); err != nil {
			return err
		}
		pos = childSpan.End
	}
	r.source(pos, span.End)
	return nil
}

func (r *markdownInlineReader) source(start, end int) {
	for _, skip := range r.skipped {
		if skip.End <= start || skip.Start >= end {
			continue
		}
		r.normalized(start, max(start, skip.Start))
		start = min(skip.End, end)
	}
	r.normalized(start, end)
}

func (r *markdownInlineReader) normalized(start, end int) {
	pos := start
	for pos < end {
		newline := bytes.IndexAny(r.doc.Source[pos:end], "\r\n")
		if newline < 0 {
			r.builder.Source(r.doc.Source, pos, end, true)
			return
		}
		lineEnd := pos + newline
		r.builder.Source(r.doc.Source, pos, lineEnd, true)
		pos = lineEnd + 1
		if r.doc.Source[lineEnd] == '\r' && pos < end && r.doc.Source[pos] == '\n' {
			pos++
		}
		r.builder.Add(" ", document.Span{Start: lineEnd, End: pos})
	}
}

func (r *markdownInlineReader) protect(span document.Span, kind string) error {
	if kind == "html_tag" && unsupportedSuppression(string(r.doc.Source[span.Start:span.End])) {
		return fmt.Errorf("suppression directives are not implemented in this alpha")
	}
	r.builder.Add(" \x00 ", span)
	r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: "inline-protected"})
	return nil
}

func (r *markdownInlineReader) linkText(node *ts.Node, depth int) error {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Type(r.syntax.lang) == "link_text" {
			return r.read(child, depth+1)
		}
	}
	return fmt.Errorf("link at byte %d has no visible label", r.offset+int(node.StartByte()))
}
