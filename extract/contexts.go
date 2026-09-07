package extract

import (
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
)

// DefaultContexts returns the global default set of supported prose contexts.
// Only contexts supported by a document's grammar apply to that document.
func DefaultContexts() []string {
	return []string{"comment", "string", "paragraph", "heading", "list-item", "table-cell"}
}

// AvailableContexts returns the contexts supported by a format, or nil for an unknown format.
func AvailableContexts(format document.Format) []string {
	if format == document.Plain {
		return []string{"paragraph"}
	}
	if format == document.Markdown {
		return []string{"paragraph", "heading", "list-item", "table-cell"}
	}
	if slices.Contains(document.Formats(), format) {
		return []string{"comment", "string"}
	}
	return nil
}

func validateContexts(policy Policy) error {
	if err := validateContextSet(policy.Contexts, DefaultContexts()); err != nil {
		return fmt.Errorf("extraction contexts: %w", err)
	}
	for format, override := range policy.Languages {
		available := AvailableContexts(format)
		if available == nil {
			return fmt.Errorf("extraction languages: unknown format %q", format)
		}
		if override.Contexts == nil {
			return fmt.Errorf("extraction language %q requires a contexts set", format)
		}
		if err := validateContextSet(override.Contexts, available); err != nil {
			return fmt.Errorf("extraction language %q: %w", format, err)
		}
	}
	return nil
}

func validateContextSet(contexts, allowed []string) error {
	seen := make(map[string]bool)
	for _, kind := range contexts {
		if !slices.Contains(allowed, kind) {
			return fmt.Errorf("unsupported prose context %q", kind)
		}
		if seen[kind] {
			return fmt.Errorf("duplicate prose context %q", kind)
		}
		seen[kind] = true
	}
	return nil
}

func contextEnabled(policy Policy, format document.Format, kind string) bool {
	contexts := policy.Contexts
	if override, ok := policy.Languages[format]; ok {
		contexts = override.Contexts
	}
	return contexts == nil || slices.Contains(contexts, kind)
}

func contextExclusion(doc *document.Document, policy Policy, kind string, span document.Span) bool {
	if contextEnabled(policy, doc.Format, kind) {
		return false
	}
	doc.Excluded = append(doc.Excluded, document.Exclusion{Span: span, Reason: "config:context-disabled:" + kind})
	return true
}

func filterContexts(doc *document.Document, policy Policy) {
	blocks := doc.Blocks[:0]
	for _, block := range doc.Blocks {
		if contextExclusion(doc, policy, block.Kind, block.Span) {
			continue
		}
		block.ID = len(blocks)
		blocks = append(blocks, block)
	}
	doc.Blocks = blocks
}
