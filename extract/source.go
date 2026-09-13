package extract

import (
	"context"
	"fmt"
	"slices"
	"strings"

	ts "github.com/stokaro/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

type sourceReader struct {
	ctx             context.Context
	doc             *document.Document
	syntax          syntaxTree
	options         Options
	comments        []document.Span
	exceptions      []compiledException
	yamlValues      map[int]yamlValue
	structureLabels map[structureKey]string
}

func sourceProse(ctx context.Context, doc *document.Document, options Options) error {
	if skip, err := prepareGo(ctx, doc, options); skip || err != nil {
		return err
	}
	syntax, err := sourceSyntax(ctx, doc)
	if err != nil {
		return err
	}
	defer syntax.tree.Release()
	reader := sourceReader{ctx: ctx, doc: doc, syntax: syntax, options: options}
	if doc.Format == document.YAML {
		reader.yamlValues, err = yamlValues(ctx, doc.Source)
		if err != nil {
			return err
		}
	}
	reader.exceptions, err = compileExceptions(options.Policy)
	if err != nil {
		return err
	}
	reader.excludeGoComments()
	if err := walkSyntax(ctx, syntax.tree.RootNode(), 0, reader.node); err != nil {
		return err
	}
	if err := reader.commentGroups(); err != nil {
		return err
	}
	return reader.finish()
}

func (r *sourceReader) finish() error {
	doc := r.doc
	slices.SortStableFunc(doc.Blocks, func(a, b document.Block) int { return a.Span.Start - b.Span.Start })
	for i := range doc.Blocks {
		doc.Blocks[i].ID = i
	}
	if r.options.IncludeStructure {
		return r.assignSourceContexts()
	}
	return nil
}

func prepareGo(ctx context.Context, doc *document.Document, options Options) (bool, error) {
	if doc.Format != document.Go {
		return false, nil
	}
	if err := goComments(ctx, doc, options); err != nil {
		return false, err
	}
	for _, excluded := range doc.Excluded {
		if excluded.Reason == "generated-go" {
			return true, nil
		}
	}
	return false, nil
}

func sourceTerminator(doc *document.Document) []byte {
	source := doc.Source
	if doc.Format == document.Fish && len(source) > 0 && source[len(source)-1] != '\n' {
		// Fish's grammar requires a final line terminator. The added token has
		// no prose, and literal and comment nodes retain their original spans.
		return append(slices.Clone(source), '\n')
	}
	return source
}

func (r *sourceReader) node(node *ts.Node) (bool, error) {
	if !node.IsNamed() {
		return true, nil
	}
	if len(r.doc.Blocks)+len(r.comments) > r.options.MaxBlocks {
		return true, fmt.Errorf("source exceeds %d prose regions", r.options.MaxBlocks)
	}
	kind := node.Type(r.syntax.lang)
	if kind == "string_start" && (node.Parent() == nil || node.Parent().Type(r.syntax.lang) != "string") {
		return true, fmt.Errorf("incomplete string at byte %d", node.StartByte())
	}
	if slices.Contains([]string{"comment", "line_comment", "block_comment"}, kind) {
		r.commentNode(node)
		return true, nil
	}
	if r.doc.Format == document.YAML {
		return r.yamlNode(node)
	}
	if !literalNode(kind, r.doc.Format) {
		return false, nil
	}
	return r.stringNode(node)
}

func (r *sourceReader) commentNode(node *ts.Node) {
	if r.doc.Format == document.Go {
		return
	}
	span := syntaxSpan(node, 0)
	if directiveCandidate(directiveText(r.doc.Source, span)) {
		r.comments = append(r.comments, span)
		return
	}
	if contextExclusion(r.doc, r.options.Policy, "comment", span) {
		return
	}
	if reason := r.exception("comment", node); reason != "" {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: reason})
	} else {
		r.comments = append(r.comments, span)
	}
}

func (r *sourceReader) stringNode(node *ts.Node) (bool, error) {
	span := syntaxSpan(node, 0)
	if contextExclusion(r.doc, r.options.Policy, "string", span) {
		return true, nil
	}
	if reason := r.exception("string", node); reason != "" {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: reason})
		return true, nil
	}
	if reason := technicalLiteral(node, r.syntax.lang, r.doc.Format); reason != "" {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: reason})
		return true, nil
	}
	spec, err := r.literalSpec(node)
	if err != nil {
		return true, err
	}
	// An excluded program is not decoded, so its escapes are never validated.
	if programFormat(r.doc.Format) && programLiteral(string(r.doc.Source[spec.span.Start:spec.span.End])) {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: "embedded-program"})
		return true, nil
	}
	mapped, err := r.literal(node, spec)
	if err != nil {
		return true, err
	}
	appendBlock(r.doc, mapped, "string")
	// Continue into interpolation expressions to find their own nested literals.
	return false, nil
}

func (r *sourceReader) excludeGoComments() {
	if r.doc.Format != document.Go {
		return
	}
	reason := r.exception("comment", nil)
	if reason == "" {
		return
	}
	for _, block := range r.doc.Blocks {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: block.Span, Reason: reason})
	}
	r.doc.Blocks = []document.Block{}
}

func literalNode(kind string, format document.Format) bool {
	if format == document.CSharp {
		return slices.Contains([]string{
			"string_literal", "verbatim_string_literal", "raw_string_literal", "interpolated_string_expression",
		}, kind)
	}
	if format == document.PowerShell {
		return slices.Contains([]string{
			"expandable_string_literal", "verbatim_string_characters", "expandable_here_string_literal", "verbatim_here_string_characters",
		}, kind)
	}
	return slices.Contains([]string{
		"interpreted_string_literal", "raw_string_literal", "string_literal", "string", "template_string",
		"raw_string", "ansi_c_string", "heredoc_body", "double_quote_string", "single_quote_string",
	}, kind)
}

func technicalLiteral(node *ts.Node, lang *ts.Language, format document.Format) string {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		kind := parent.Type(lang)
		if slices.Contains([]string{"import_spec", "import_statement", "preproc_include"}, kind) {
			return "import-path"
		}
		if format == document.Go && kind == "field_declaration" {
			return "struct-tag"
		}
	}
	return ""
}

func (r *sourceReader) commentGroups() error {
	var builder mapping.Builder
	previous := 0
	for _, span := range r.comments {
		if previous > 0 && !adjacentComments(r.doc.Source[previous:span.Start]) {
			appendBlock(r.doc, builder.Build(), "comment")
			builder = mapping.Builder{}
		}
		content := commentContent(r.doc.Source, span)
		found, err := collectDirective(r.doc, span, directiveText(r.doc.Source, span))
		if err != nil {
			return err
		}
		if found {
			appendBlock(r.doc, builder.Build(), "comment")
			builder = mapping.Builder{}
			previous = span.End
			continue
		}
		original := string(r.doc.Source[span.Start:span.End])
		if r.commentDirective(original, string(r.doc.Source[content.Start:content.End])) {
			appendBlock(r.doc, builder.Build(), "comment")
			builder = mapping.Builder{}
			r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: span, Reason: "directive"})
		} else {
			addCommentLines(r.doc, &builder, content.Start, content.End, commentOpener(original))
			if span.End < len(r.doc.Source) {
				builder.Add(" ", document.Span{Start: span.End, End: span.End + 1})
			}
		}
		previous = span.End
	}
	appendBlock(r.doc, builder.Build(), "comment")
	return nil
}

func adjacentComments(gap []byte) bool {
	text := string(gap)
	return strings.TrimSpace(text) == "" && strings.Count(text, "\n") <= 1
}

// commentContent narrows a comment span to its text. A terminator belongs to the
// block comment that opened it. A line comment that ends with the same bytes
// keeps them. A comment whose opener and terminator overlap, such as "/*/",
// leaves an empty span rather than an inverted one.
func commentContent(source []byte, span document.Span) document.Span {
	text := string(source[span.Start:span.End])
	opener := ""
	for _, marker := range []string{"///", "//!", "//", "/*", "<#", "#"} {
		if strings.HasPrefix(text, marker) {
			span.Start += len(marker)
			opener = marker
			break
		}
	}
	terminators := map[string]string{"/*": "*/", "<#": "#>"}
	if terminator := terminators[opener]; terminator != "" &&
		len(text) >= len(opener)+len(terminator) && strings.HasSuffix(text, terminator) {
		span.End -= len(terminator)
	}
	return span
}

// commentDirective reports whether a comment is a shebang or a tool directive
// for the document's format.
func (r *sourceReader) commentDirective(original, content string) bool {
	return strings.HasPrefix(original, "#!") || toolDirective(r.doc.Format, content)
}
