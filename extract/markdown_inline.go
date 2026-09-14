package extract

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"sort"

	ts "github.com/stokaro/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

type markdownInlineReader struct {
	ctx       context.Context
	doc       *document.Document
	syntax    syntaxTree
	offset    int
	skipped   []document.Span
	protected []document.Span
	limit     int // end of the inline span being read
	consumed  int // source already held back by an inline code element
	builder   mapping.Builder
}

func markdownInlineProtected(ctx context.Context, doc *document.Document, span document.Span,
	skipped, protected []document.Span) (document.MappedText, error) {
	first := sort.Search(len(protected), func(i int) bool { return protected[i].End > span.Start })
	last := sort.Search(len(protected), func(i int) bool { return protected[i].Start >= span.End })
	protected = protected[first:last]
	source := bytes.Clone(doc.Source[span.Start:span.End])
	for _, region := range protected {
		if region.End > span.Start && region.Start < span.End {
			maskRange(source, max(region.Start-span.Start, 0), min(region.End-span.Start, len(source)))
		}
	}
	for _, skip := range skipped {
		maskRange(source, skip.Start-span.Start, skip.End-span.Start)
	}
	syntax, err := inlineSyntax(ctx, source)
	if err != nil {
		return document.MappedText{}, err
	}
	defer syntax.tree.Release()
	reader := markdownInlineReader{ctx: ctx, doc: doc, syntax: syntax, offset: span.Start, limit: span.End,
		skipped: skipped, protected: protected}
	err = reader.read(syntax.tree.RootNode(), 0)
	mapped := reader.builder.Build()
	if doc.Format == document.MDX {
		mapped = trimMDXTags(doc.Source, mapped)
	}
	return mapped, err
}

// The inline node kinds each rule of read covers.
var (
	protectedInline  = []string{"code_span", "image", "uri_autolink", "email_autolink", "html_tag"}
	inlineLinks      = []string{"inline_link", "full_reference_link", "collapsed_reference_link", "shortcut_link"}
	inlineDelimiters = []string{"emphasis_delimiter", "strikethrough_delimiter"}
	inlineReferences = []string{"entity_reference", "numeric_character_reference", "backslash_escape"}
)

func (r *markdownInlineReader) read(node *ts.Node, depth int) error {
	if err := r.ctx.Err(); err != nil {
		return err
	}
	if depth > 128 {
		return fmt.Errorf("inline syntax nesting exceeds 128")
	}
	kind := node.Type(r.syntax.lang)
	span := syntaxSpan(node, r.offset)
	if r.codeElement(kind, span) {
		return nil
	}
	switch {
	case slices.Contains(protectedInline, kind):
		return r.protect(span, kind)
	case slices.Contains(inlineLinks, kind):
		return r.linkText(node, depth)
	case slices.Contains(inlineDelimiters, kind):
		return nil
	case slices.Contains(inlineReferences, kind):
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
	start = max(start, r.consumed)
	if start >= end {
		return
	}
	for _, skip := range r.skipped {
		if skip.End <= start || skip.Start >= end {
			continue
		}
		r.unprotected(start, max(start, skip.Start))
		start = min(skip.End, end)
	}
	r.unprotected(start, end)
}

func (r *markdownInlineReader) unprotected(start, end int) {
	for _, span := range r.protected {
		if span.End <= start || span.Start >= end {
			continue
		}
		r.normalized(start, max(start, span.Start))
		r.builder.Add(" \x00 ", document.Span{Start: max(start, span.Start), End: min(end, span.End)})
		start = min(span.End, end)
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

// htmlCodeElement names the element that an opening tag starts, when the
// element carries code rather than prose, and returns "" otherwise. The tags
// of an inline "pre" or "code" are protected on their own; without this the
// example they enclose is read as prose.
func htmlCodeElement(tag string) string {
	for _, name := range []string{"pre", "code"} {
		if markupOpen(tag, "<"+name) {
			return name
		}
	}
	return ""
}

// codeElement reports whether a node belongs to an inline code element: the
// opening tag of one, which it holds back together with the element's body,
// or a node the last such hold already covered.
func (r *markdownInlineReader) codeElement(kind string, span document.Span) bool {
	if span.End <= r.consumed {
		return true
	}
	if kind != "html_tag" {
		return false
	}
	name := htmlCodeElement(string(r.doc.Source[span.Start:span.End]))
	if name == "" {
		return false
	}
	r.holdElement(span, name)
	return true
}

// holdElement replaces an inline code element and everything up to its
// closing tag with one break, the same treatment a code span receives. An
// element left open runs to the end of the inline span.
func (r *markdownInlineReader) holdElement(span document.Span, name string) {
	closer := "</" + name + ">"
	region := document.Span{Start: span.Start, End: r.limit}
	if i := indexFold(string(r.doc.Source[span.End:r.limit]), closer); i >= 0 {
		region.End = span.End + i + len(closer)
	}
	r.builder.Add(" \x00 ", region)
	r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: region, Reason: "inline-protected"})
	r.consumed = region.End
}

func (r *markdownInlineReader) protect(span document.Span, kind string) error {
	if kind == "html_tag" {
		if err := htmlDirectives(r.doc, span); err != nil {
			return err
		}
	}
	r.builder.Add(" \x00 ", span)
	r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: "inline-protected"})
	return nil
}

// linkText reads the visible label of a link. A link without one carries no
// prose, so it becomes a protected boundary like an image. The inline grammar
// also reads bracket pairs in ordinary prose as a reference link, as in the
// Java type long[][], and those pairs have no label either.
func (r *markdownInlineReader) linkText(node *ts.Node, depth int) error {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Type(r.syntax.lang) == "link_text" {
			return r.read(child, depth+1)
		}
	}
	return r.protect(syntaxSpan(node, r.offset), node.Type(r.syntax.lang))
}
