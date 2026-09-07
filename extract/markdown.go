package extract

import (
	"context"
	"fmt"
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

type markdownReader struct {
	ctx     context.Context
	doc     *document.Document
	syntax  syntaxTree
	options Options
}

func markdown(ctx context.Context, doc *document.Document, options Options) error {
	syntax, err := parseSyntax(ctx, maskFrontMatter(doc), "markdown")
	if err != nil {
		return err
	}
	defer syntax.tree.Release()
	reader := markdownReader{ctx: ctx, doc: doc, syntax: syntax, options: options}
	return walkSyntax(ctx, syntax.tree.RootNode(), 0, reader.block)
}

func (r markdownReader) block(node *ts.Node) (bool, error) {
	if len(r.doc.Blocks) > r.options.MaxBlocks {
		return true, fmt.Errorf("source exceeds %d prose blocks", r.options.MaxBlocks)
	}
	kind := node.Type(r.syntax.lang)
	if reason := markdownExclusion(kind, r.options.IncludeQuotes); reason != "" {
		span := syntaxSpan(node, 0)
		if reason == "html" && unsupportedSuppression(string(r.doc.Source[span.Start:span.End])) {
			return true, fmt.Errorf("suppression directives are not implemented in this alpha")
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

func (r markdownReader) inline(node *ts.Node, kind string) error {
	span := syntaxSpan(node, 0)
	if contextExclusion(r.doc, r.options.Policy, kind, span) {
		return nil
	}
	var continuations []document.Span
	err := walkSyntax(r.ctx, node, 0, func(child *ts.Node) (bool, error) {
		if child.Type(r.syntax.lang) == "block_continuation" && child.EndByte() > child.StartByte() {
			continuations = append(continuations, syntaxSpan(child, 0))
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	if kind == "table-cell" {
		for span.End > span.Start && strings.ContainsRune(" \t", rune(r.doc.Source[span.End-1])) {
			span.End--
		}
	}
	mapped, err := markdownInline(r.ctx, r.doc, span, continuations)
	if err == nil {
		appendBlock(r.doc, mapped, kind)
	}
	return err
}
