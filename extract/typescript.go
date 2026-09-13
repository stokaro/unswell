package extract

import (
	"bytes"
	"context"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
)

// typeScriptSyntax supports type-only wildcard exports missing from the pinned
// grammar. Only a "type" token between export and * is blanked in an owned
// buffer. A complete reparse must confirm each modified export, without errors
// or missing nodes. Comments, literals, and original offsets stay intact.
func typeScriptSyntax(ctx context.Context, source []byte, name string) (syntaxTree, error) {
	original, err := parseSyntaxTree(ctx, source, name)
	if err != nil || !original.tree.RootNode().HasErrorOrMissing() {
		return original, err
	}
	defer original.tree.Release()
	modifiers, err := typeExportModifiers(ctx, original, source)
	if err != nil {
		return syntaxTree{}, err
	}
	if len(modifiers) == 0 {
		return syntaxTree{}, invalidSyntax(name)
	}
	masked := bytes.Clone(source)
	for _, span := range modifiers {
		blank(masked, span.Start, span.End)
	}
	parsed, err := parseSyntax(ctx, masked, name)
	if err != nil {
		return syntaxTree{}, err
	}
	if err := verifyTypeExports(ctx, parsed, modifiers); err != nil {
		parsed.tree.Release()
		return syntaxTree{}, err
	}
	return parsed, nil
}

// Keep token order even when error recovery splits a multiline declaration into
// expressions. Comments are extras; other leaves interrupt the candidate.
func typeExportModifiers(ctx context.Context, syntax syntaxTree, source []byte) ([]document.Span, error) {
	var modifiers []document.Span
	var previous, beforePrevious string
	var previousSpan document.Span
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if node.Type(syntax.lang) == "comment" {
			return true, nil
		}
		if node.ChildCount() > 0 {
			return false, nil
		}
		span := syntaxSpan(node, 0)
		if !span.Valid(len(source)) || node.IsMissing() {
			previous, beforePrevious = "", ""
			return true, nil
		}
		text := string(source[span.Start:span.End])
		if beforePrevious == "export" && previous == "type" && text == "*" {
			modifiers = append(modifiers, previousSpan)
		}
		beforePrevious, previous, previousSpan = previous, text, span
		return true, nil
	})
	return modifiers, err
}

func verifyTypeExports(ctx context.Context, syntax syntaxTree, modifiers []document.Span) error {
	index := 0
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if index >= len(modifiers) {
			return true, nil
		}
		gap := typeExportGap(syntax, node)
		if gap.Start <= modifiers[index].Start && modifiers[index].End <= gap.End {
			index++
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	if index != len(modifiers) {
		return invalidSyntax(syntax.lang.Name)
	}
	return nil
}

func typeExportGap(syntax syntaxTree, node *ts.Node) document.Span {
	if node.Type(syntax.lang) != "export_statement" {
		return document.Span{}
	}
	export := node.Child(0)
	if export == nil || export.Type(syntax.lang) != "export" {
		return document.Span{}
	}
	wildcard := export.NextSibling()
	for wildcard != nil && wildcard.Type(syntax.lang) == "comment" {
		wildcard = wildcard.NextSibling()
	}
	if wildcard == nil {
		return document.Span{}
	}
	switch wildcard.Type(syntax.lang) {
	case "*", "namespace_export":
		return document.Span{Start: int(export.EndByte()), End: int(wildcard.StartByte())}
	default:
		return document.Span{}
	}
}
