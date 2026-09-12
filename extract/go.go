package extract

import (
	"bytes"
	"context"
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
		span := document.Span{Start: start, End: end}
		found, err := collectDirective(doc, span, directiveText(doc.Source, span))
		if err != nil {
			return err
		}
		if found {
			appendBlock(doc, builder.Build(), "comment")
			builder = mapping.Builder{}
			continue
		}
		if directive(comment.Text) {
			appendBlock(doc, builder.Build(), "comment")
			builder = mapping.Builder{}
			doc.Excluded = append(doc.Excluded, document.Exclusion{Span: document.Span{Start: start, End: end}, Reason: "directive"})
			continue
		}
		contentStart, contentEnd := start+2, end
		opener := commentOpener(comment.Text)
		if opener != "" {
			contentEnd -= 2
		}
		addCommentLines(doc, &builder, contentStart, contentEnd, opener)
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

// directive reports whether a Go comment is a build, generate, lint or
// generated-file instruction rather than prose.
func directive(text string) bool {
	text = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(text, "//"), "/*"))
	return toolDirective(document.Go, text)
}
