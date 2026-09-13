package extract

import (
	"fmt"
	"sort"
	"strings"

	ts "github.com/stokaro/gotreesitter"

	"github.com/stokaro/unswell/document"
)

type structureKey struct {
	span document.Span
	kind string
}

// Structural labels come from the same syntax tree used for extraction. They
// contain no line numbers or sibling ordinals and never become analyzed prose.
func (r *sourceReader) assignSourceContexts() error {
	r.structureLabels = make(map[structureKey]string)
	for i := range r.doc.Blocks {
		if err := r.ctx.Err(); err != nil {
			return err
		}
		labels, err := r.sourceContext(r.doc.Blocks[i].Span.Start)
		if err != nil {
			return err
		}
		r.doc.Blocks[i].Context = labels
	}
	return nil
}

func (r *sourceReader) sourceContext(offset int) ([]string, error) {
	var labels []string
	node := r.syntax.tree.RootNode()
	for depth := 0; node != nil; depth++ {
		if err := r.ctx.Err(); err != nil {
			return nil, err
		}
		if depth > 128 {
			return nil, fmt.Errorf("structural context nesting exceeds 128")
		}
		if label := r.ownerLabel(node); label != "" {
			labels = append(labels, label)
		}
		index := sort.Search(node.ChildCount(), func(i int) bool { return int(node.Child(i).EndByte()) > offset })
		if index == node.ChildCount() || int(node.Child(index).StartByte()) > offset {
			break
		}
		child := node.Child(index)
		if sourceComment(child.Type(r.syntax.lang)) {
			return r.followingContext(labels, node, index+1)
		}
		node = child
	}
	return labels, nil
}

func (r *sourceReader) followingContext(labels []string, node *ts.Node, start int) ([]string, error) {
	label, err := r.followingOwner(node, start)
	if err != nil {
		return nil, err
	}
	if label != "" {
		labels = append(labels, "following:"+label)
	}
	return labels, nil
}

func sourceComment(kind string) bool {
	return kind == "comment" || kind == "line_comment" || kind == "block_comment"
}

func (r *sourceReader) followingOwner(parent *ts.Node, start int) (string, error) {
	for i := start; i < parent.ChildCount(); i++ {
		if err := r.ctx.Err(); err != nil {
			return "", err
		}
		node := parent.Child(i)
		if sourceComment(node.Type(r.syntax.lang)) {
			continue
		}
		return r.ownerLabel(node), nil
	}
	return "", nil
}

func (r *sourceReader) ownerLabel(node *ts.Node) string {
	kind := node.Type(r.syntax.lang)
	if !symbolOwner(kind) && kind != "function_statement" {
		return ""
	}
	key := structureKey{span: syntaxSpan(node, 0), kind: kind}
	if label, exists := r.structureLabels[key]; exists {
		return label
	}
	label := r.uncachedOwnerLabel(node, kind)
	r.structureLabels[key] = label
	return label
}

func (r *sourceReader) uncachedOwnerLabel(node *ts.Node, kind string) string {
	for _, field := range []string{"name", "left", "key", "declarator", "pattern"} {
		child := node.ChildByFieldName(field, r.syntax.lang)
		name := r.contextText(child)
		if name == "" {
			continue
		}
		var label strings.Builder
		label.WriteString(kind + ":" + field + ":" + name)
		for _, qualifier := range []string{"receiver", "parameters", "type_parameters"} {
			if text := r.contextText(node.ChildByFieldName(qualifier, r.syntax.lang)); text != "" {
				label.WriteString(":" + qualifier + ":" + text)
			}
		}
		return label.String()
	}
	return r.powerShellOwnerLabel(node)
}

func (r *sourceReader) contextText(node *ts.Node) string {
	if node == nil {
		return ""
	}
	span := syntaxSpan(node, 0)
	if !span.Valid(len(r.doc.Source)) {
		return ""
	}
	return compactContext(r.doc.Source, span)
}

func (r *sourceReader) powerShellOwnerLabel(node *ts.Node) string {
	if r.doc.Format != document.PowerShell {
		return ""
	}
	// The pinned grammar exposes assignment targets and function names as child
	// node kinds instead of named fields. Keep that distinction explicit.
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		kind := child.Type(r.syntax.lang)
		if kind == "left_assignment_expression" || kind == "function_name" {
			return node.Type(r.syntax.lang) + ":" + kind + ":" + r.contextText(child)
		}
	}
	return ""
}

func compactContext(source []byte, span document.Span) string {
	return strings.Join(strings.Fields(string(source[span.Start:span.End])), " ")
}
