package extract

import (
	"fmt"
	"slices"

	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/mapping"
)

type markdownScope struct {
	span document.Span
	kind string
}

func (r *markdownReader) headingContext(node *ts.Node, mapped document.MappedText) error {
	level := markdownHeadingLevel(node, r.syntax.lang)
	if level == 0 {
		return fmt.Errorf("markdown heading has no structural level")
	}
	label, err := mapping.Canonical(mapped, r.doc.Source)
	if err != nil {
		return err
	}
	scope, headings := r.headingState(node)
	headings[level-1] = fmt.Sprintf("heading-%d:%s", level, label)
	clear(headings[level:])
	r.headings[scope] = headings
	return nil
}

// Use grammar heading levels within their quote/list containers. The pinned
// grammar's section wrappers do not nest ATX and setext headings consistently.
func (r *markdownReader) headingState(node *ts.Node) (markdownScope, [6]string) {
	var scopes []markdownScope
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		kind := parent.Type(r.syntax.lang)
		if kind == "block_quote" || kind == "list_item" {
			scopes = append(scopes, markdownScope{span: r.input.span(parent), kind: kind})
		}
	}
	current := markdownScope{}
	headings := r.headings[current]
	for _, scope := range slices.Backward(scopes) {
		current = scope
		if previous, exists := r.headings[scope]; exists {
			headings = previous
		} else {
			r.headings[scope] = headings
		}
	}
	return current, headings
}

func (r *markdownReader) sectionContext(node *ts.Node) []string {
	scope, headings := r.headingState(node)
	var labels []string
	if scope.kind != "" {
		labels = append(labels, "container:"+scope.kind)
	}
	for _, label := range headings {
		if label != "" {
			labels = append(labels, label)
		}
	}
	return labels
}

func markdownHeadingLevel(node *ts.Node, lang *ts.Language) int {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		kind := parent.Type(lang)
		if kind != "atx_heading" && kind != "setext_heading" {
			continue
		}
		for i := 0; i < parent.ChildCount(); i++ {
			if level := headingMarkerLevel(parent.Child(i).Type(lang)); level > 0 {
				return level
			}
		}
		return 0
	}
	return 0
}

func headingMarkerLevel(marker string) int {
	switch marker {
	case "atx_h1_marker", "setext_h1_underline":
		return 1
	case "atx_h2_marker", "setext_h2_underline":
		return 2
	case "atx_h3_marker":
		return 3
	case "atx_h4_marker":
		return 4
	case "atx_h5_marker":
		return 5
	case "atx_h6_marker":
		return 6
	default:
		return 0
	}
}
