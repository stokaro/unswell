package extract

import (
	"fmt"
	"strings"

	ts "github.com/stokaro/gotreesitter"

	"github.com/stokaro/unswell/document"
)

// The pinned grammar can omit a list wrapper around adjacent list_item nodes.
// Group direct siblings using grammar markers, never source-line punctuation.
func (r *markdownReader) indexLists() error {
	id := 0
	return walkSyntax(r.ctx, r.syntax.tree.RootNode(), 0, func(parent *ts.Node) (bool, error) {
		return false, r.indexListChildren(parent, &id)
	})
}

func (r *markdownReader) indexListChildren(parent *ts.Node, id *int) error {
	var items []*ts.Node
	marker := ""
	flush := func() error {
		if len(items) == 0 {
			return nil
		}
		*id++
		err := r.indexListItems(items, *id)
		items = nil
		return err
	}
	for i := 0; i < parent.ChildCount(); i++ {
		child := parent.Child(i)
		next := r.listMarker(child)
		if next == "" || next != marker {
			if err := flush(); err != nil {
				return err
			}
		}
		if next != "" {
			items = append(items, child)
		}
		marker = next
	}
	return flush()
}

func (r *markdownReader) listMarker(node *ts.Node) string {
	if node.Type(r.syntax.lang) != "list_item" {
		return ""
	}
	for i := 0; i < node.ChildCount(); i++ {
		kind := node.Child(i).Type(r.syntax.lang)
		if strings.HasPrefix(kind, "list_marker_") {
			return kind
		}
	}
	return ""
}

func (r *markdownReader) indexListItems(items []*ts.Node, id int) error {
	info := document.ListContext{ID: id, Items: len(items)}
	for _, item := range items {
		if err := r.inspectListItem(item, &info); err != nil {
			return err
		}
	}

	for ordinal, item := range items {
		info.Item, info.Depth = ordinal+1, 1
		for parent := item.Parent(); parent != nil; parent = parent.Parent() {
			if parent.Type(r.syntax.lang) == "list_item" {
				info.Depth++
			}
		}
		r.lists[r.input.span(item)] = info
	}
	return nil
}

func (r *markdownReader) inspectListItem(item *ts.Node, info *document.ListContext) error {
	inlines := 0
	err := walkSyntax(r.ctx, item, 0, func(node *ts.Node) (bool, error) {
		kind := node.Type(r.syntax.lang)
		if kind == "list_item" && r.input.span(node) != r.input.span(item) {
			info.Complex = true
			return true, nil
		}
		inspectListKind(info, kind)
		if kind == "inline" {
			inlines++
		}
		return false, nil
	})
	info.Complex = info.Complex || inlines > 1
	return err
}

func (r *markdownReader) listContext(node *ts.Node) (*document.ListContext, error) {
	for parent := node.Parent(); parent != nil; parent = parent.Parent() {
		if parent.Type(r.syntax.lang) != "list_item" {
			continue
		}
		info, exists := r.lists[r.input.span(parent)]
		if !exists {
			return nil, fmt.Errorf("markdown list item has no grammar list owner")
		}
		return &info, nil
	}
	return nil, fmt.Errorf("markdown list item has no grammar item owner")
}

func inspectListKind(info *document.ListContext, kind string) {
	if kind == "list_marker_dot" || kind == "list_marker_parenthesis" {
		info.Ordered = true
	}
	if strings.HasPrefix(kind, "task_list_marker_") {
		info.Task = true
	}
	if kind == "atx_heading" || kind == "setext_heading" || markdownExclusion(kind, true) != "" {
		info.Complex = true
	}
}
