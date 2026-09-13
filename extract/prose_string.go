package extract

import (
	"fmt"
	"slices"
	"strings"

	ts "github.com/stokaro/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/pathglob"
)

func (r *sourceReader) proseString(node *ts.Node, mapped document.MappedText) error {
	for _, selector := range r.markdownStrings {
		if r.markdownStringMatches(node, selector) {
			return r.markdownString(mapped, selector.value.ID)
		}
	}
	appendBlock(r.doc, mapped, "string")
	return nil
}

func (r *sourceReader) markdownStringMatches(node *ts.Node, selector compiledException) bool {
	value := selector.value
	return pathglob.Matches(r.doc.Name, selector.paths) &&
		(len(value.Formats) == 0 || slices.Contains(value.Formats, r.doc.Format)) &&
		(len(value.Symbols) == 0 || r.withinSymbol(node, value.Symbols))
}

func (r *sourceReader) markdownString(mapped document.MappedText, id string) error {
	if strings.ContainsRune(mapped.Text, 0) {
		return fmt.Errorf("markdown_strings %q requires a static string without protected interpolation or controls", id)
	}
	options := r.options
	options.MaxBlocks = max(1, options.MaxBlocks-len(r.doc.Blocks))
	inner, err := Parse(r.ctx, document.Source{Name: r.doc.Name, Format: document.Markdown, Bytes: []byte(mapped.Text)}, options)
	if err != nil {
		return fmt.Errorf("markdown_strings %q at byte %d: %w", id, document.Bounds(mapped.Map).Start, err)
	}
	if len(r.doc.Blocks)+len(inner.Blocks) > r.options.MaxBlocks {
		return fmt.Errorf("source exceeds %d prose blocks", r.options.MaxBlocks)
	}
	selected := make([]bool, len(mapped.Text))
	if err := r.embeddedBlocks(inner, mapped, document.Markdown, selected); err != nil {
		return err
	}
	return r.embeddedControls(inner, mapped, selected)
}
