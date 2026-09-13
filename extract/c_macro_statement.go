package extract

import (
	"bytes"
	"context"
	"slices"

	ts "github.com/stokaro/gotreesitter"
)

// cFamilySyntax parses C and C++ source. A function-like macro used as a
// statement with a compound body and no semicolon, such as
// list_for_each(pos, head) { ... }, leaves the grammar one missing ";" that
// nothing else explains. The call is blanked in an owned copy of the source,
// so the body parses as a plain block at the same offsets. That happens only
// when the call holds no comment or string and no other error exists. Every
// other incomplete tree stays an error.
func cFamilySyntax(ctx context.Context, source []byte, name string) (syntaxTree, error) {
	original, err := parseSyntaxTree(ctx, source, name)
	if err != nil || !original.tree.RootNode().HasErrorOrMissing() {
		return original, err
	}
	defer original.tree.Release()
	masked := bytes.Clone(source)
	blanked, err := blankMacroStatements(ctx, original, masked)
	if err != nil {
		return syntaxTree{}, err
	}
	if !blanked {
		return syntaxTree{}, invalidSyntax(name)
	}
	return parseSyntax(ctx, masked, name)
}

// blankMacroStatements blanks every macro statement of the tree in the
// buffer and reports whether it blanked at least one with no other error
// left. An error or missing node outside a macro statement makes it report
// false without touching the buffer further.
func blankMacroStatements(ctx context.Context, syntax syntaxTree, masked []byte) (bool, error) {
	count, unexplained := 0, false
	err := walkSyntax(ctx, syntax.tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if unexplained || !node.HasErrorOrMissing() {
			return true, nil
		}
		if call := macroStatementCall(syntax, node); call != nil {
			if !prosefree(ctx, syntax, call) {
				unexplained = true
				return true, nil
			}
			blank(masked, int(call.StartByte()), int(call.EndByte()))
			count++
			return true, nil
		}
		if node.IsError() || node.IsMissing() {
			unexplained = true
			return true, nil
		}
		return false, nil
	})
	return err == nil && count > 0 && !unexplained, err
}

// macroStatementCall returns the call of an expression statement that ends
// in a missing ";" and precedes a compound statement, or nil.
func macroStatementCall(syntax syntaxTree, node *ts.Node) *ts.Node {
	if node.Type(syntax.lang) != "expression_statement" || node.ChildCount() != 2 {
		return nil
	}
	call, terminator := node.Child(0), node.Child(1)
	if call.Type(syntax.lang) != "call_expression" || call.HasErrorOrMissing() || !terminator.IsMissing() {
		return nil
	}
	next := node.NextSibling()
	for next != nil && !next.IsNamed() {
		next = next.NextSibling()
	}
	if next == nil || next.Type(syntax.lang) != "compound_statement" {
		return nil
	}
	return call
}

// prosefree reports whether a subtree holds no comment or string, so blanking
// it removes no prose.
func prosefree(ctx context.Context, syntax syntaxTree, node *ts.Node) bool {
	free := true
	err := walkSyntax(ctx, node, 0, func(child *ts.Node) (bool, error) {
		kind := child.Type(syntax.lang)
		if kind == "comment" || slices.Contains([]string{"string_literal", "raw_string_literal", "char_literal",
			"concatenated_string", "system_lib_string", "user_defined_literal"}, kind) {
			free = false
			return true, nil
		}
		return false, nil
	})
	return err == nil && free
}

// blank overwrites a span with spaces, keeping line terminators in place.
func blank(buffer []byte, start, end int) {
	for i := start; i < end && i < len(buffer); i++ {
		if buffer[i] != '\n' && buffer[i] != '\r' {
			buffer[i] = ' '
		}
	}
}
