package extract

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

func goComments(ctx context.Context, doc *document.Document, _ Options) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, doc.Name, doc.Source, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return err
	}
	if ast.IsGenerated(file) {
		doc.Excluded = append(
			doc.Excluded,
			document.Exclusion{Span: document.Span{Start: 0, End: len(doc.Source)}, Reason: "generated-go"},
		)
		return nil
	}
	cgo := cgoGroups(file)
	for _, group := range file.Comments {
		if err := ctx.Err(); err != nil {
			return err
		}
		if cgo[group] {
			doc.Excluded = append(doc.Excluded, document.Exclusion{
				Span: document.Span{
					Start: fset.Position(group.Pos()).Offset,
					End:   originalCommentEnd(doc.Source, fset.Position(group.List[len(group.List)-1].Pos()).Offset),
				}, Reason: "cgo-preamble",
			})
			continue
		}
		if err := extractCommentGroup(doc, group, fset); err != nil {
			return err
		}
	}
	return nil
}

func cgoGroups(file *ast.File) map[*ast.CommentGroup]bool {
	result := make(map[*ast.CommentGroup]bool)
	for _, decl := range file.Decls {
		general, ok := decl.(*ast.GenDecl)
		if !ok || general.Tok != token.IMPORT {
			continue
		}
		for _, spec := range general.Specs {
			imp, ok := spec.(*ast.ImportSpec)
			if ok && imp.Path.Value == `"C"` {
				result[general.Doc] = true
				result[imp.Doc] = true
			}
		}
	}
	return result
}

func extractCommentGroup(doc *document.Document, group *ast.CommentGroup, fset *token.FileSet) error {
	var builder mapping.Builder
	for _, comment := range group.List {
		start := fset.Position(comment.Pos()).Offset
		end := originalCommentEnd(doc.Source, start)
		if unsupportedSuppression(comment.Text) {
			return fmt.Errorf("suppression directives are not implemented in this alpha")
		}
		if directive(comment.Text) {
			appendBlock(doc, builder.Build(), "comment")
			builder = mapping.Builder{}
			doc.Excluded = append(doc.Excluded, document.Exclusion{Span: document.Span{Start: start, End: end}, Reason: "go-directive"})
			continue
		}
		contentStart, contentEnd := start+2, end
		if strings.HasPrefix(comment.Text, "/*") {
			contentEnd -= 2
		}
		addCommentLines(doc, &builder, contentStart, contentEnd)
		if end < len(doc.Source) {
			builder.Add(" ", document.Span{Start: end, End: end + 1})
		}
	}
	appendBlock(doc, builder.Build(), "comment")
	return nil
}

func originalCommentEnd(source []byte, start int) int {
	if bytes.HasPrefix(source[start:], []byte("/*")) {
		if end := bytes.Index(source[start+2:], []byte("*/")); end >= 0 {
			return start + end + 4
		}
	}
	if end := bytes.IndexByte(source[start:], '\n'); end >= 0 {
		return start + end
	}
	return len(source)
}

func unsupportedSuppression(text string) bool {
	text = strings.TrimSpace(text)
	for _, prefix := range []string{"<!--", "//", "/*"} {
		text = strings.TrimSpace(strings.TrimPrefix(text, prefix))
	}
	return strings.HasPrefix(text, "unswell-disable") || strings.HasPrefix(text, "unswell-enable")
}

func directive(text string) bool {
	text = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(text, "//"), "/*"))
	for _, prefix := range []string{"go:", "+build", "nolint", "lint:", "line ", "SPDX-", "Code generated", "unswell-"} {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

func addCommentLines(doc *document.Document, builder *mapping.Builder, start, end int) {
	pos := start
	for line := range strings.SplitSeq(string(doc.Source[start:end]), "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			appendBlock(doc, builder.Build(), "comment")
			*builder = mapping.Builder{}
		case strings.HasPrefix(line, "\t"), strings.HasPrefix(line, "    "):
			builder.Add(" \x00 ", document.Span{Start: pos, End: pos + len(line)})
			doc.Excluded = append(
				doc.Excluded,
				document.Exclusion{Span: document.Span{Start: pos, End: pos + len(line)}, Reason: "comment-code"},
			)
		default:
			builder.Source(doc.Source, pos, pos+len(line), false)
		}
		pos += len(line) + 1
		if pos <= end {
			builder.Add(" ", document.Span{Start: pos - 1, End: pos})
		}
	}
}
