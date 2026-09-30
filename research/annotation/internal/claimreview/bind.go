package claimreview

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
)

const maximumClaims = 10000

// Bind verifies declared claims against an existing source-preserving document.
// It owns copies of source and claim data; every later stage uses the same originals.
func Bind(ctx context.Context, doc document.Document, specifications []Specification) (Inventory, error) {
	if err := ctx.Err(); err != nil {
		return Inventory{}, err
	}
	if err := validateSource(doc, specifications); err != nil {
		return Inventory{}, err
	}
	if err := validateDocument(doc); err != nil {
		return Inventory{}, err
	}
	claims := make([]Claim, 0, len(specifications))
	seen := make(map[Origin]bool)
	for _, specification := range specifications {
		if err := bindSpecification(ctx, doc, specification, seen); err != nil {
			return Inventory{}, err
		}
		id := digest(struct {
			Version string
			Path    string
			Format  document.Format
			Source  string
			Claim   Specification
		}{Version, doc.Name, doc.Format, doc.Hash, specification})
		claims = append(claims, Claim{ID: id, Specification: cloneSpecification(specification)})
	}
	doc.Source = slices.Clone(doc.Source)
	doc.Blocks = slices.Clone(doc.Blocks)
	doc.Excluded = slices.Clone(doc.Excluded)
	return Inventory{document: doc, claims: claims}, nil
}

func validateSource(doc document.Document, specifications []Specification) error {
	if len(doc.Source) > 2<<20 || !utf8.Valid(doc.Source) || doc.Name == "" ||
		doc.Hash != fmt.Sprintf("%x", sha256.Sum256(doc.Source)) {
		return fmt.Errorf("invalid source identity or size")
	}
	if len(specifications) > maximumClaims || len(doc.Blocks) > maximumClaims || len(doc.Excluded) > maximumClaims {
		return fmt.Errorf("claim or source inventory exceeds %d entries", maximumClaims)
	}
	count := 0
	for _, spec := range specifications {
		count += len(spec.Targets) + len(spec.Support)
	}
	if count > maximumClaims {
		return fmt.Errorf("source references exceed %d entries", maximumClaims)
	}
	return nil
}

func validateDocument(doc document.Document) error {
	seen := make(map[int]bool)
	for _, block := range doc.Blocks {
		if block.ID < 0 || seen[block.ID] || !block.Span.Valid(len(doc.Source)) {
			return fmt.Errorf("invalid or duplicate engine block")
		}
		seen[block.ID] = true
	}
	for _, excluded := range doc.Excluded {
		if !excluded.Span.Valid(len(doc.Source)) {
			return fmt.Errorf("invalid protected source range")
		}
	}
	return nil
}

func bindSpecification(ctx context.Context, doc document.Document, spec Specification, seen map[Origin]bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, text := range []string{spec.Origin.Run, spec.Origin.Candidate, spec.Origin.Key, spec.Category,
		spec.Diagnostic, spec.Reason, spec.Suggestion} {
		if !validText(text) {
			return fmt.Errorf("missing or excessive claim identity or explanation")
		}
	}
	if seen[spec.Origin] {
		return fmt.Errorf("duplicate original claim provenance")
	}
	seen[spec.Origin] = true
	if len(spec.Targets) == 0 || len(spec.Targets) > 64 || len(spec.Support) > 64 {
		return fmt.Errorf("a claim needs 1..64 targets and at most 64 support references")
	}
	for _, refs := range [][]Reference{spec.Targets, spec.Support} {
		if err := validateReferences(doc, refs); err != nil {
			return err
		}
	}
	return distinctRoles(spec.Targets, spec.Support)
}

func distinctRoles(targets, support []Reference) error {
	for _, target := range targets {
		if slices.ContainsFunc(support, func(ref Reference) bool { return overlaps(target.Span, ref.Span) }) {
			return fmt.Errorf("support-only context overlaps a revision target")
		}
	}
	return nil
}

func validateReferences(doc document.Document, refs []Reference) error {
	previous := -1
	for _, ref := range refs {
		if ref.Span.Start < previous {
			return fmt.Errorf("references must be ordered and disjoint")
		}
		if err := validateReference(doc, ref); err != nil {
			return err
		}
		previous = ref.Span.End
	}
	return nil
}

func validateReference(doc document.Document, ref Reference) error {
	if !ref.Span.Valid(len(doc.Source)) || !runeBoundary(doc.Source, ref.Span.Start) ||
		!runeBoundary(doc.Source, ref.Span.End) || string(doc.Source[ref.Span.Start:ref.Span.End]) != ref.Quote {
		return fmt.Errorf("reference is not the exact original UTF-8 range")
	}
	index := slices.IndexFunc(doc.Blocks, func(block document.Block) bool { return block.ID == ref.Block })
	if index < 0 || doc.Blocks[index].Excluded || !contains(doc.Blocks[index].Span, ref.Span) {
		return fmt.Errorf("reference has no matching eligible engine block")
	}
	for _, excluded := range doc.Excluded {
		if overlaps(excluded.Span, ref.Span) {
			return fmt.Errorf("reference intersects protected source")
		}
	}
	return nil
}

func runeBoundary(source []byte, offset int) bool {
	return offset == len(source) || utf8.RuneStart(source[offset])
}

func validText(text string) bool {
	return strings.TrimSpace(text) != "" && len(text) <= 2400 && utf8.ValidString(text)
}

func contains(outer, inner document.Span) bool {
	return outer.Start <= inner.Start && inner.End <= outer.End
}

func overlaps(a, b document.Span) bool { return a.Start < b.End && b.Start < a.End }

func digest(value any) string {
	data, _ := json.Marshal(value) // Only fixed structs, strings, and integer spans are passed here.
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func cloneSpecification(spec Specification) Specification {
	spec.Targets = slices.Clone(spec.Targets)
	spec.Support = slices.Clone(spec.Support)
	return spec
}

// Claims returns an owned copy in original input order, with stable source-bound IDs.
func (in Inventory) Claims() []Claim {
	result := slices.Clone(in.claims)
	for i := range result {
		result[i].Specification = cloneSpecification(result[i].Specification)
	}
	return result
}

// EditDigest binds an approval to the entire exact edit, including its identity.
func EditDigest(edit Edit) string { return digest(edit) }
