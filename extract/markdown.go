package extract

import (
	"context"
	"fmt"
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

type markdownReader struct {
	ctx      context.Context
	doc      *document.Document
	syntax   syntaxTree
	input    markdownInput
	options  Options
	headings map[markdownScope][6]string
	lists    map[document.Span]document.ListContext
}

func markdown(ctx context.Context, doc *document.Document, options Options) error {
	input, err := prepareMarkdown(ctx, maskFrontMatter(doc))
	if err != nil {
		return err
	}
	syntax, err := parseSyntax(ctx, input.source, "markdown")
	if err != nil {
		return err
	}
	defer syntax.tree.Release()
	reader := markdownReader{ctx: ctx, doc: doc, syntax: syntax, input: input, options: options,
		headings: make(map[markdownScope][6]string), lists: make(map[document.Span]document.ListContext)}
	if options.IncludeStructure {
		if err := reader.indexLists(); err != nil {
			return err
		}
	}
	return walkSyntax(ctx, syntax.tree.RootNode(), 0, reader.block)
}

func (r *markdownReader) block(node *ts.Node) (bool, error) {
	if len(r.doc.Blocks) > r.options.MaxBlocks {
		return true, fmt.Errorf("source exceeds %d prose blocks", r.options.MaxBlocks)
	}
	kind := node.Type(r.syntax.lang)
	if kind == "|" && !markdownTableDelimiter(node.Parent(), r.syntax.lang) {
		return true, fmt.Errorf("markdown grammar returned an orphaned table delimiter")
	}
	if reason := markdownExclusion(kind, r.options.IncludeQuotes); reason != "" {
		span := r.input.span(node)
		if reason == "html" {
			if err := htmlDirectives(r.doc, span); err != nil {
				return true, err
			}
		}
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: reason})
		return true, nil
	}
	if kind == "pipe_table_cell" {
		return true, r.inline(node, "table-cell")
	}
	if kind != "inline" {
		return false, nil
	}
	return true, r.inline(node, markdownKind(node, r.syntax.lang))
}

func markdownTableDelimiter(parent *ts.Node, lang *ts.Language) bool {
	if parent == nil {
		return false
	}
	switch parent.Type(lang) {
	case "pipe_table_row", "pipe_table_header", "pipe_table_delimiter_row":
		return true
	default:
		return false
	}
}

func markdownExclusion(kind string, quotes bool) string {
	switch kind {
	case "fenced_code_block", "indented_code_block":
		return "code"
	case "html_block":
		return "html"
	case "link_reference_definition":
		return "link-definition"
	case "block_quote":
		if !quotes {
			return "quote"
		}
	}
	return ""
}

func markdownKind(node *ts.Node, lang *ts.Language) string {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		switch parent.Type(lang) {
		case "atx_heading", "setext_heading":
			return "heading"
		case "list_item":
			return "list-item"
		}
	}
	return "paragraph"
}

func (r *markdownReader) inline(node *ts.Node, kind string) error {
	span := r.inlineSpan(node, kind)
	if span.End == span.Start {
		return nil
	}
	excluded := contextExclusion(r.doc, r.options.Policy, kind, span)
	if excluded && (kind != "heading" || !r.options.IncludeStructure) {
		return nil
	}
	mapped, err := r.inlineMapping(node, span)
	if err != nil {
		return err
	}
	return r.appendInline(node, mapped, kind, excluded)
}

func (r *markdownReader) inlineSpan(node *ts.Node, kind string) document.Span {
	span := r.input.span(node)
	if kind == "table-cell" {
		for span.End > span.Start && strings.ContainsRune(" \t", rune(r.doc.Source[span.End-1])) {
			span.End--
		}
	}
	return span
}

func (r *markdownReader) inlineMapping(node *ts.Node, span document.Span) (document.MappedText, error) {
	var continuations []document.Span
	err := walkSyntax(r.ctx, node, 0, func(child *ts.Node) (bool, error) {
		if child.Type(r.syntax.lang) == "block_continuation" && child.EndByte() > child.StartByte() {
			continuations = append(continuations, r.input.span(child))
		}
		return false, nil
	})
	if err != nil {
		return document.MappedText{}, err
	}
	return markdownInline(r.ctx, r.doc, span, continuations)
}

func (r *markdownReader) appendInline(node *ts.Node, mapped document.MappedText, kind string, excluded bool) error {
	if kind == "heading" && r.options.IncludeStructure {
		if err := r.headingContext(node, mapped); err != nil {
			return err
		}
	}
	if excluded {
		return nil
	}
	before := len(r.doc.Blocks)
	appendBlock(r.doc, mapped, kind)
	if r.options.IncludeStructure && len(r.doc.Blocks) > before {
		r.doc.Blocks[before].Context = r.sectionContext(node)
		if kind == "list-item" {
			list, err := r.listContext(node)
			if err != nil {
				return err
			}
			r.doc.Blocks[before].List = list
		}
	}
	return nil
}
