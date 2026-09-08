package baseline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

type index struct {
	documents map[string]Document
	entries   map[string]Entry
}

// Fingerprint returns the versioned hash of a validated structural identity.
func Fingerprint(identity Identity) (string, error) {
	if err := validateIdentity(identity); err != nil {
		return "", err
	}
	data, err := json.Marshal(struct {
		Version  string   `json:"version"`
		Identity Identity `json:"identity"`
	}{FingerprintVersion, identity})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func validateIdentity(identity Identity) error {
	if !validPath(identity.Path) {
		return fmt.Errorf("baseline identity requires a normalized project-relative path")
	}
	if err := validateIdentityKind(identity); err != nil {
		return err
	}
	if !validHash(identity.StructureHash) || !validHash(identity.ContentHash) || !validHash(identity.EvidenceHash) {
		return fmt.Errorf("baseline identity requires structural, content, and evidence SHA-256 hashes")
	}
	return nil
}

func validateIdentityKind(identity Identity) error {
	switch identity.Kind {
	case "finding":
		if !validName(identity.RuleID) || !strings.Contains(identity.RuleID, ".") || !validName(identity.RuleVersion) {
			return fmt.Errorf("baseline finding requires a rule ID and behavior version")
		}
	case "sentence", "paragraph":
		if identity.RuleID != "" || identity.RuleVersion != "" {
			return fmt.Errorf("baseline scored units cannot carry a rule identity")
		}
	default:
		return fmt.Errorf("unknown baseline identity kind %q", identity.Kind)
	}
	return nil
}

func validHash(value string) bool {
	return len(value) == 64 && strings.Trim(value, "0123456789abcdef") == ""
}

func validName(value string) bool {
	return len(value) > 0 && len(value) <= 128 && !strings.ContainsFunc(value, func(r rune) bool { return r <= ' ' || r > '~' })
}

func validPath(value string) bool {
	return value != "" && len(value) <= 4096 && value != "." && value != ".." &&
		!strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "../") &&
		!strings.ContainsAny(value, "\\:\x00\r\n") && path.Clean(value) == value
}

func validateCompatibility(value Compatibility) error {
	if !validHash(value.PolicyHash) || !validHash(value.RulesHash) || !validHash(value.NLPHash) || !validHash(value.ModelHash) {
		return fmt.Errorf("baseline compatibility requires policy, rules, NLP, and model SHA-256 hashes")
	}
	if !validName(value.FeatureContract) || !validName(value.ScoringContract) {
		return fmt.Errorf("baseline compatibility requires feature and scoring contracts")
	}
	return nil
}

func documentIndex(ctx context.Context, documents []Document) (map[string]Document, error) {
	if len(documents) == 0 || len(documents) > MaxDocuments {
		return nil, fmt.Errorf("baseline requires 1 to %d completely analyzed documents", MaxDocuments)
	}
	result := make(map[string]Document, len(documents))
	for _, doc := range documents {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !validDocument(doc) {
			return nil, fmt.Errorf("invalid baseline document %q", doc.Path)
		}
		if _, exists := result[doc.Path]; exists {
			return nil, fmt.Errorf("duplicate baseline document %q", doc.Path)
		}
		result[doc.Path] = doc
	}
	return result, nil
}

func validDocument(doc Document) bool {
	return validPath(doc.Path) && validName(doc.Format) && validHash(doc.SourceHash) &&
		validHash(doc.PolicyHash) && validHash(doc.SuppressionHash)
}

func fileIndex(ctx context.Context, file File) (index, error) {
	if err := ctx.Err(); err != nil {
		return index{}, err
	}
	if file.Version != Version || file.FingerprintVersion != FingerprintVersion {
		return index{}, fmt.Errorf("incompatible baseline or fingerprint version")
	}
	if err := validateCompatibility(file.Compatibility); err != nil {
		return index{}, err
	}
	documents, err := documentIndex(ctx, file.Documents)
	if err != nil {
		return index{}, err
	}
	if file.Entries == nil || len(file.Entries) > MaxEntries {
		return index{}, fmt.Errorf("baseline entries must be an array with at most %d entries", MaxEntries)
	}
	result := index{documents: documents, entries: make(map[string]Entry, len(file.Entries))}
	for _, entry := range file.Entries {
		if err := result.add(ctx, entry); err != nil {
			return index{}, err
		}
	}
	return result, nil
}

func (i index) add(ctx context.Context, entry Entry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fingerprint, err := Fingerprint(entry.Identity)
	if err != nil {
		return err
	}
	if entry.Fingerprint != fingerprint || entry.Status != "accepted" {
		return fmt.Errorf("baseline entry has an invalid fingerprint or acceptance status")
	}
	if _, exists := i.documents[entry.Identity.Path]; !exists {
		return fmt.Errorf("baseline entry references an unobserved document %q", entry.Identity.Path)
	}
	if _, exists := i.entries[fingerprint]; exists {
		return fmt.Errorf("ambiguous baseline identity %s in %s", fingerprint, entry.Identity.Path)
	}
	i.entries[fingerprint] = entry
	return nil
}

func snapshotFile(ctx context.Context, snapshot Snapshot) (File, index, error) {
	if err := ctx.Err(); err != nil {
		return File{}, index{}, err
	}
	if !snapshot.Complete {
		return File{}, index{}, fmt.Errorf("incomplete analysis cannot create or compare accepted debt")
	}
	if len(snapshot.Candidates) > MaxEntries {
		return File{}, index{}, fmt.Errorf("baseline exceeds %d candidates", MaxEntries)
	}
	file := File{Version: Version, FingerprintVersion: FingerprintVersion, Compatibility: snapshot.Compatibility,
		Documents: snapshot.Documents, Entries: make([]Entry, 0, len(snapshot.Candidates))}
	for _, candidate := range snapshot.Candidates {
		if err := ctx.Err(); err != nil {
			return File{}, index{}, err
		}
		fingerprint, err := Fingerprint(candidate)
		if err != nil {
			return File{}, index{}, err
		}
		file.Entries = append(file.Entries, Entry{Identity: candidate, Fingerprint: fingerprint, Status: "accepted"})
	}
	indexed, err := fileIndex(ctx, file)
	return file, indexed, err
}
