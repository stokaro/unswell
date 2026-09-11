package extract

import (
	"bytes"
	"context"
	"strings"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

func inlineSyntax(ctx context.Context, source []byte) (syntaxTree, error) {
	original, err := parseSyntaxTree(ctx, source, "markdown_inline")
	if err != nil || !original.tree.RootNode().HasErrorOrMissing() {
		return original, err
	}
	defer original.tree.Release()
	// The inline grammar rejects double hyphens in HTML comments, including code
	// examples. Normalize only grammar-identified boundaries in this owned buffer.
	comments, err := maskInlineBoundaries(ctx, original, source)
	if err != nil {
		return syntaxTree{}, err
	}
	if len(comments) == 0 {
		return syntaxTree{}, invalidSyntax("markdown_inline")
	}
	parsed, err := parseSyntax(ctx, source, "markdown_inline")
	if err != nil {
		return syntaxTree{}, err
	}
	if err := verifyInlineBoundaries(ctx, parsed, comments); err != nil {
		parsed.tree.Release()
		return syntaxTree{}, err
	}
	return parsed, nil
}

func maskInlineBoundaries(ctx context.Context, syntax syntaxTree, source []byte) (map[document.Span]string, error) {
	comments := make(map[document.Span]string)
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		kind := node.Type(syntax.lang)
		if (kind != "html_tag" && kind != "code_span") || !node.HasErrorOrMissing() {
			return false, nil
		}
		span := syntaxSpan(node, 0)
		if !span.Valid(len(source)) || len(comments) >= 1000 {
			return true, nil
		}
		body := normalizableInlineBody(syntax, node, source)
		if len(body) == 0 {
			return true, nil
		}
		for i := 0; i+1 < len(body); i++ {
			if body[i] == '-' && body[i+1] == '-' {
				body[i], body[i+1] = ' ', ' '
			}
		}
		comments[span] = kind
		return true, nil
	})
	return comments, err
}

func normalizableInlineBody(syntax syntaxTree, node *ts.Node, source []byte) []byte {
	span := syntaxSpan(node, 0)
	if node.Type(syntax.lang) == "code_span" {
		return delimitedCodeBody(syntax, node, source)
	}
	raw := source[span.Start:span.End]
	if !bytes.HasPrefix(raw, []byte("<!--")) || !bytes.HasSuffix(raw, []byte("-->")) || len(raw) < 7 {
		return nil
	}
	body := raw[4 : len(raw)-3]
	if !directiveCandidate(string(body)) {
		return nil
	}
	return body
}

func delimitedCodeBody(syntax syntaxTree, node *ts.Node, source []byte) []byte {
	if node.ChildCount() < 2 {
		return nil
	}
	first, last := node.Child(0), node.Child(node.ChildCount()-1)
	if first.Type(syntax.lang) != "code_span_delimiter" || last.Type(syntax.lang) != "code_span_delimiter" {
		return nil
	}
	left, right := syntaxSpan(first, 0), syntaxSpan(last, 0)
	if !left.Valid(len(source)) || !right.Valid(len(source)) || left.End > right.Start ||
		!bytes.Equal(source[left.Start:left.End], source[right.Start:right.End]) {
		return nil
	}
	return source[left.End:right.Start]
}

func verifyInlineBoundaries(ctx context.Context, syntax syntaxTree, expected map[document.Span]string) error {
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		kind := node.Type(syntax.lang)
		if kind == "html_tag" || kind == "code_span" {
			span := syntaxSpan(node, 0)
			if expected[span] == kind {
				delete(expected, span)
			}
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	if len(expected) != 0 {
		return invalidSyntax("markdown_inline")
	}
	return nil
}

func sourceSyntax(ctx context.Context, doc *document.Document) (syntaxTree, error) {
	if doc.Format == document.C || doc.Format == document.CPP {
		return cFamilySyntax(ctx, sourceTerminator(doc), string(doc.Format))
	}
	if doc.Format != document.PowerShell {
		return parseSyntax(ctx, sourceTerminator(doc), string(doc.Format))
	}
	original, err := parseSyntaxTree(ctx, doc.Source, "powershell")
	if err != nil || !original.tree.RootNode().HasErrorOrMissing() {
		return original, err
	}
	defer original.tree.Release()
	commentsOnly, err := onlyPowerShellComments(ctx, original, doc.Source)
	if err != nil {
		return syntaxTree{}, err
	}
	if !commentsOnly {
		return syntaxTree{}, invalidSyntax("powershell")
	}
	// This grammar requires a statement even in a comment-only script. An empty
	// statement after the source supplies it without changing any original span.
	return parseSyntax(ctx, append(bytes.Clone(doc.Source), '\n', ';', '\n'), "powershell")
}

func onlyPowerShellComments(ctx context.Context, syntax syntaxTree, source []byte) (bool, error) {
	previous := 0
	onlyComments := true
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if node.Type(syntax.lang) != "comment" || node.HasErrorOrMissing() {
			return false, nil
		}
		span := syntaxSpan(node, 0)
		if !span.Valid(len(source)) || span.Start < previous {
			onlyComments = false
			return true, nil
		}
		if !sourceWhitespace(source, previous, span.Start) {
			onlyComments = false
		}
		previous = span.End
		return true, nil
	})
	return onlyComments && sourceWhitespace(source, previous, len(source)), err
}

func sourceWhitespace(source []byte, start, end int) bool {
	gap := string(source[start:end])
	if start == 0 {
		gap = strings.TrimPrefix(gap, "\ufeff")
	}
	return strings.TrimSpace(gap) == ""
}
