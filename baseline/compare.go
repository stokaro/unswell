package baseline

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// Create accepts every candidate in a complete snapshot, without retaining aliases.
func Create(ctx context.Context, snapshot Snapshot) (File, error) {
	file, _, err := snapshotFile(ctx, snapshot)
	if err != nil {
		return File{}, err
	}
	return ordered(file), nil
}

// Compare matches exact debt without changing the baseline or the snapshot.
// Incompatibility and ambiguity are errors; stale and unobserved debt are explicit.
func Compare(ctx context.Context, file File, snapshot Snapshot) (Comparison, error) {
	previous, err := fileIndex(ctx, file)
	if err != nil {
		return Comparison{}, err
	}
	_, current, err := snapshotFile(ctx, snapshot)
	if err != nil {
		return Comparison{}, err
	}
	if file.Compatibility != snapshot.Compatibility {
		return Comparison{}, fmt.Errorf("baseline analysis compatibility changed; explicit update requires complete coverage")
	}
	if err := compatibleDocuments(previous.documents, snapshot.Documents); err != nil {
		return Comparison{}, err
	}
	return compareIndexes(ctx, previous, current)
}

func compatibleDocuments(previous map[string]Document, current []Document) error {
	for _, doc := range current {
		old, exists := previous[doc.Path]
		if exists && (old.Format != doc.Format || old.PolicyHash != doc.PolicyHash || old.SuppressionHash != doc.SuppressionHash) {
			return fmt.Errorf("baseline policy, format, or suppressions changed for %s; explicit update required", doc.Path)
		}
	}
	return nil
}

func compareIndexes(ctx context.Context, previous, current index) (Comparison, error) {
	result := Comparison{Matches: []Match{}, Stale: []Entry{}, Unobserved: []Entry{}}
	for fingerprint := range current.entries {
		if err := ctx.Err(); err != nil {
			return Comparison{}, err
		}
		state := "new"
		if _, exists := previous.entries[fingerprint]; exists {
			state = "existing"
		}
		result.Matches = append(result.Matches, Match{Fingerprint: fingerprint, State: state})
	}
	for fingerprint, entry := range previous.entries {
		if err := ctx.Err(); err != nil {
			return Comparison{}, err
		}
		if _, observed := current.documents[entry.Identity.Path]; !observed {
			result.Unobserved = append(result.Unobserved, entry)
		} else if _, exists := current.entries[fingerprint]; !exists {
			result.Stale = append(result.Stale, entry)
		}
	}
	slices.SortFunc(result.Matches, func(a, b Match) int { return strings.Compare(a.Fingerprint, b.Fingerprint) })
	slices.SortFunc(result.Stale, compareEntries)
	slices.SortFunc(result.Unobserved, compareEntries)
	return result, nil
}

// Update explicitly replaces debt for observed documents and preserves unobserved
// paths. Changing global compatibility requires observing every previous path.
func Update(ctx context.Context, file File, snapshot Snapshot) (File, error) {
	previous, err := fileIndex(ctx, file)
	if err != nil {
		return File{}, err
	}
	updated, current, err := snapshotFile(ctx, snapshot)
	if err != nil {
		return File{}, err
	}
	updated.Documents = slices.Clone(updated.Documents)
	for _, doc := range file.Documents {
		if err := ctx.Err(); err != nil {
			return File{}, err
		}
		if _, observed := current.documents[doc.Path]; observed {
			continue
		}
		if file.Compatibility != snapshot.Compatibility {
			return File{}, fmt.Errorf("baseline compatibility update requires observing %s", doc.Path)
		}
		updated.Documents = append(updated.Documents, doc)
	}
	comparison, err := compareIndexes(ctx, previous, current)
	if err != nil {
		return File{}, err
	}
	updated.Entries = append(updated.Entries, comparison.Unobserved...)
	if _, err := fileIndex(ctx, updated); err != nil {
		return File{}, err
	}
	return ordered(updated), nil
}

func ordered(file File) File {
	file.Documents = slices.Clone(file.Documents)
	file.Entries = slices.Clone(file.Entries)
	slices.SortFunc(file.Documents, func(a, b Document) int { return strings.Compare(a.Path, b.Path) })
	slices.SortFunc(file.Entries, compareEntries)
	return file
}

func compareEntries(a, b Entry) int {
	if a.Identity.Path != b.Identity.Path {
		return strings.Compare(a.Identity.Path, b.Identity.Path)
	}
	return strings.Compare(a.Fingerprint, b.Fingerprint)
}
