package extract

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

type literalSpec struct {
	span   document.Span
	mode   string
	binary bool
	indent string
}

func (r *sourceReader) literal(node *ts.Node, spec literalSpec) (document.MappedText, error) {
	if spec.binary {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: syntaxSpan(node, 0), Reason: "byte-literal"})
		return document.MappedText{}, nil
	}
	holes, err := r.interpolations(node)
	if err != nil {
		return document.MappedText{}, err
	}
	var builder mapping.Builder
	pos := spec.span.Start
	for _, hole := range holes {
		if err := literalSegment(&builder, r.doc.Source, pos, hole.Start, spec); err != nil {
			return document.MappedText{}, err
		}
		builder.Add(" \x00 ", hole)
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: hole, Reason: "interpolation"})
		pos = hole.End
	}
	if err := literalSegment(&builder, r.doc.Source, pos, spec.span.End, spec); err != nil {
		return document.MappedText{}, err
	}
	mapped := builder.Build()
	if !utf8.ValidString(mapped.Text) {
		r.doc.Excluded = append(r.doc.Excluded, document.Exclusion{Span: syntaxSpan(node, 0), Reason: "non-utf8-literal"})
		return document.MappedText{}, nil
	}
	return mapped, nil
}

func (r *sourceReader) interpolations(node *ts.Node) ([]document.Span, error) {
	var spans []document.Span
	err := walkSyntax(r.ctx, node, 0, func(child *ts.Node) (bool, error) {
		if slices.Contains([]string{
			"template_substitution", "interpolation", "simple_expansion", "expansion", "command_substitution",
			"variable_expansion", "variable", "sub_expression", "arithmetic_expansion",
		}, child.Type(r.syntax.lang)) {
			spans = append(spans, syntaxSpan(child, 0))
			return true, nil
		}
		return false, nil
	})
	return spans, err
}

func (r *sourceReader) literalSpec(node *ts.Node) (literalSpec, error) {
	span := syntaxSpan(node, 0)
	kind := node.Type(r.syntax.lang)
	text := string(r.doc.Source[span.Start:span.End])
	if r.doc.Format == document.CSharp {
		return r.csharpLiteral(node, text, kind)
	}
	if kind == "heredoc_body" {
		return literalSpec{span: span, mode: r.heredocMode(node)}, nil
	}
	if kind == "raw_string_literal" && r.doc.Format != document.Go {
		return r.rawLiteral(node)
	}
	return r.quotedLiteral(node, text, kind)
}

func (r *sourceReader) quotedLiteral(node *ts.Node, text, kind string) (literalSpec, error) {
	span := syntaxSpan(node, 0)
	quote := strings.IndexAny(text, "\"'`")
	if quote < 0 {
		return literalSpec{}, fmt.Errorf("string at byte %d has no opening delimiter", span.Start)
	}
	width := quoteWidth(r.doc.Format, text[quote:])
	span.Start += quote + width
	span.End -= width
	if strings.Contains(kind, "here_string") {
		span.End-- // PowerShell closes here-strings with a quote followed by @.
	}
	if span.End < span.Start {
		return literalSpec{}, fmt.Errorf("string at byte %d has mismatched delimiters", node.StartByte())
	}
	mode := literalMode(r.doc.Format, kind, text[:quote])
	binary := r.doc.Format == document.Python && strings.Contains(strings.ToLower(text[:quote]), "b")
	return literalSpec{span: span, mode: mode, binary: binary}, nil
}

func quoteWidth(format document.Format, text string) int {
	if slices.Contains([]document.Format{document.Java, document.Python}, format) && strings.HasPrefix(text, strings.Repeat(text[:1], 3)) {
		return 3
	}
	return 1
}

func (r *sourceReader) rawLiteral(node *ts.Node) (literalSpec, error) {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if slices.Contains([]string{"string_content", "raw_string_content"}, child.Type(r.syntax.lang)) {
			return literalSpec{span: syntaxSpan(child, 0), mode: "raw"}, nil
		}
	}
	span := syntaxSpan(node, 0)
	text := string(r.doc.Source[span.Start:span.End])
	start, end := strings.IndexByte(text, '"')+1, strings.LastIndexByte(text, '"')
	if r.doc.Format == document.CPP {
		start, end = strings.IndexByte(text, '(')+1, strings.LastIndexByte(text, ')')
	}
	if start > 0 && start == end {
		return literalSpec{span: document.Span{Start: span.Start + start, End: span.Start + end}, mode: "raw"}, nil
	}
	return literalSpec{}, fmt.Errorf("raw string at byte %d has no content node", node.StartByte())
}

func (r *sourceReader) heredocMode(node *ts.Node) string {
	parent := node.Parent()
	for i := 0; parent != nil && i < parent.ChildCount(); i++ {
		child := parent.Child(i)
		if child.Type(r.syntax.lang) == "heredoc_start" {
			span := syntaxSpan(child, 0)
			if strings.ContainsAny(string(r.doc.Source[span.Start:span.End]), "'\"\\") {
				return "raw"
			}
		}
	}
	return "shell"
}

func literalMode(format document.Format, kind, prefix string) string {
	modes := map[string]string{
		"raw_string_literal": "raw", "raw_string": "raw", "ansi_c_string": "ansi",
		"single_quote_string": "fish-single", "double_quote_string": "fish-double",
		"verbatim_string_characters": "powershell-single", "verbatim_here_string_characters": "powershell-single",
	}
	if mode, ok := modes[kind]; ok {
		return mode
	}
	if format == document.Python && strings.Contains(strings.ToLower(prefix), "r") {
		return "raw"
	}
	if slices.Contains([]document.Format{document.Bash, document.Shell, document.Zsh}, format) {
		return "shell"
	}
	return string(format)
}
