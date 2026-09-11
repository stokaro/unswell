package corpus

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

//go:embed manifest.schema.json
var manifestSchema []byte

// LoadManifest decodes strict JSON and validates acquisition declarations.
// It cannot establish that a provenance or permission assertion is true.
func LoadManifest(ctx context.Context, data []byte) (Manifest, error) {
	var manifest Manifest
	if err := jsoninput.Decode(ctx, data, MaxManifestBytes, &manifest, inputLimits()); err != nil {
		return Manifest{}, err
	}
	if err := jsoninput.Schema(data, manifestSchema, "urn:unswell:corpus:manifest:v1"); err != nil {
		return Manifest{}, fmt.Errorf("corpus manifest schema: %w", err)
	}
	if err := manifest.validate(ctx); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func inputLimits() jsoninput.Limits {
	return jsoninput.Limits{Array: MaxUnits, Object: MaxSources,
		Arrays: map[string]int{"keys": 100000, "segments": 1024, "role_regions": 1000}}
}

func (m Manifest) validate(ctx context.Context) error {
	if err := m.validateHeader(); err != nil {
		return err
	}
	if err := m.validateSources(ctx); err != nil {
		return err
	}
	_, err := requiredFiles(m)
	return err
}

func (m Manifest) validateHeader() error {
	if m.Version != Version || !text(m.ID) || !text(m.Seed) || len(m.Seed) > 128 {
		return fmt.Errorf("invalid corpus version, ID, or seed")
	}
	w := m.Weights
	if err := w.validate(); err != nil {
		return err
	}
	if err := extract.ValidatePolicy(m.Policy); err != nil {
		return err
	}
	if err := choices(m.UnitKinds, []string{"sentence", "paragraph", "fragment"}, 3); err != nil {
		return err
	}
	if len(m.Sources) == 0 || len(m.Sources) > MaxSources {
		return fmt.Errorf("source count must be 1 through %d", MaxSources)
	}
	return nil
}

func (w Weights) validate() error {
	for _, value := range []int{w.Training, w.Development, w.Calibration, w.FinalTest} {
		if value <= 0 || value > 10000 {
			return fmt.Errorf("partition weights must be positive and at most 10000")
		}
	}
	if w.Training+w.Development+w.Calibration+w.FinalTest != 10000 {
		return fmt.Errorf("partition weights must sum to 10000")
	}
	return nil
}

func (m Manifest) validateSources(ctx context.Context) error {
	ids, paths := make(map[string]bool), make(map[string]bool)
	total, metadata := 0, 0
	for _, source := range m.Sources {
		if err := ctx.Err(); err != nil {
			return err
		}
		if ids[source.ID] || paths[source.Path] {
			return fmt.Errorf("duplicate source ID or path %q", source.ID)
		}
		ids[source.ID], paths[source.Path] = true, true
		if err := source.validate(); err != nil {
			return fmt.Errorf("source %s: %w", source.ID, err)
		}
		total += source.Bytes
		if total > MaxTotalBytes {
			return fmt.Errorf("sources exceed %d bytes", MaxTotalBytes)
		}
		data, err := json.Marshal(source)
		if err != nil {
			return err
		}
		metadata += len(data)
		if metadata > MaxManifestBytes {
			return fmt.Errorf("source metadata exceeds manifest size limit")
		}
	}
	return nil
}

func (s Source) validate() error {
	if err := s.validateIdentity(); err != nil {
		return err
	}
	if err := s.validateMetadata(); err != nil {
		return err
	}
	if err := s.validateRights(); err != nil {
		return err
	}
	if err := s.validateOrigin(); err != nil {
		return err
	}
	if err := s.validateSnapshot(); err != nil {
		return err
	}
	return s.validateRoles()
}

func (s Source) validateIdentity() error {
	if err := validateLocation(s.Path, s.SHA256, s.Bytes); err != nil {
		return err
	}
	for _, value := range []string{s.ID, s.Repository, s.Document, s.Reference, s.Topic, s.Purpose} {
		if !text(value) {
			return fmt.Errorf("source identity, context, and references must be nonempty bounded text")
		}
	}
	if !slices.Contains(document.Formats(), s.Format) || s.ProseLanguage != "en" {
		return fmt.Errorf("a supported source format and declared English prose are required")
	}
	if s.Partition != "" && !slices.Contains(partitions(), s.Partition) {
		return fmt.Errorf("unknown partition %q", s.Partition)
	}
	if !slices.Contains(roles(), s.Role) {
		return fmt.Errorf("unknown input role %q", s.Role)
	}
	return nil
}

func (s Source) validateMetadata() error {
	for _, group := range [][]string{s.Authors, s.Templates, s.Related, s.GenerationTasks} {
		if err := uniqueText(group, 100); err != nil {
			return err
		}
	}
	if (s.AuthorLanguage == "") != (s.AuthorLanguageBasis == "") {
		return fmt.Errorf("author language requires independent metadata and its basis")
	}
	for _, value := range []string{s.AuthorLanguage, s.AuthorLanguageBasis, s.Origin.GenerationRecord} {
		if value != "" && !text(value) {
			return fmt.Errorf("optional metadata must be bounded text when present")
		}
	}
	return nil
}

func (s Source) validateRights() error {
	if !text(s.Rights.License) || !text(s.Rights.Evidence) || len(s.Notices) == 0 || len(s.Notices) > 20 {
		return fmt.Errorf("reviewed permission evidence and retained notices are required")
	}
	uses := []string{"annotation", "training", "evaluation", "redistribute_source", "redistribute_annotations"}
	if err := choices(s.Rights.AllowedUses, uses, len(uses)); err != nil {
		return err
	}
	if !slices.Contains(s.Rights.AllowedUses, "annotation") {
		return fmt.Errorf("annotation permission is required for candidate preparation")
	}
	seen := make(map[string]bool)
	for _, notice := range s.Notices {
		if err := validateLocation(notice.Path, notice.SHA256, notice.Bytes); err != nil {
			return err
		}
		if seen[notice.Path] {
			return fmt.Errorf("duplicate notice %s", notice.Path)
		}
		seen[notice.Path] = true
	}
	return nil
}

func (s Source) validateOrigin() error {
	// A whole-file origin assertion must not become an automatic unit label.
	labels := []string{"human", "generated", "human_ai_edited", "generated_human_edited", "mixed", "unknown"}
	validScope := slices.Contains([]string{"document", "repository"}, s.Origin.Scope)
	if !slices.Contains(labels, s.Origin.Label) || !validScope || !text(s.Origin.Evidence) {
		return fmt.Errorf("source provenance requires a known label, evidence, and document/repository scope")
	}
	if s.Origin.Label != "human" && s.Origin.Label != "unknown" && !text(s.Origin.GenerationRecord) {
		return fmt.Errorf("generated or edited source provenance requires a generation record")
	}
	return nil
}

func (s Source) validateSnapshot() error {
	snapshot := s.Snapshot
	if snapshot == nil {
		return nil
	}
	if !slices.Contains(cohorts(), snapshot.Cohort) || !slices.Contains(dateConfidences(), snapshot.Confidence) ||
		!text(snapshot.Evidence) {
		return fmt.Errorf("snapshot requires a known cohort, a known date confidence, and evidence")
	}
	if err := snapshot.validateDate(); err != nil {
		return err
	}
	// An undated source cannot represent any period; a controlled output must
	// carry the generation provenance the origin contract already requires.
	if strings.HasPrefix(snapshot.Cohort, "historical") && snapshot.Confidence == "unknown" {
		return fmt.Errorf("historical cohort membership requires a dated snapshot")
	}
	if snapshot.Cohort == "controlled" && slices.Contains([]string{"human", "unknown"}, s.Origin.Label) {
		return fmt.Errorf("controlled cohort membership requires generated or edited origin")
	}
	return nil
}

func (snapshot Snapshot) validateDate() error {
	if snapshot.Date == "" {
		if snapshot.Confidence != "unknown" {
			return fmt.Errorf("a corroborated or vcs_only snapshot requires a date")
		}
		return nil
	}
	if _, err := time.Parse(time.DateOnly, snapshot.Date); err != nil || len(snapshot.Date) != len(time.DateOnly) {
		return fmt.Errorf("snapshot date must use the YYYY-MM-DD form")
	}
	return nil
}

func (s Source) validateRoles() error {
	if len(s.Roles) > 1000 {
		return fmt.Errorf("too many role regions")
	}
	end := 0
	for _, region := range s.Roles {
		if !region.Span.Valid(s.Bytes) || region.Span.Start < end || !slices.Contains(roles(), region.Role) {
			return fmt.Errorf("role regions must be ordered, disjoint, bounded, and use known roles")
		}
		end = region.Span.End
	}
	return nil
}

// ValidPath accepts relative paths without traversal or common platform aliases.
func ValidPath(name string) bool {
	if !pathForm(name) {
		return false
	}
	for part := range strings.SplitSeq(name, "/") {
		if !portablePart(part) {
			return false
		}
	}
	return true
}

func pathForm(name string) bool {
	return name != "" && len(name) <= 1024 && path.Clean(name) == name && name != "." && utf8.ValidString(name) &&
		!strings.HasPrefix(name, "/") && !strings.ContainsAny(name, "\\:\x00") && !strings.ContainsFunc(name, unicode.IsControl)
}

func portablePart(part string) bool {
	if part == ".." || strings.TrimRight(part, " .") != part || strings.ContainsAny(part, "<>\"|?*") {
		return false
	}
	base, _, _ := strings.Cut(strings.ToUpper(part), ".")
	if slices.Contains([]string{"CON", "PRN", "AUX", "NUL"}, base) {
		return false
	}
	return len(base) != 4 || (!strings.HasPrefix(base, "COM") && !strings.HasPrefix(base, "LPT")) || base[3] < '1' || base[3] > '9'
}

func validateLocation(name, hash string, size int) error {
	if !ValidPath(name) || !validHash(hash) || size <= 0 || size > MaxSourceBytes {
		return fmt.Errorf("invalid source path, SHA-256, or byte count")
	}
	return nil
}

func text(value string) bool {
	return strings.TrimSpace(value) != "" && len(value) <= 4096 && utf8.ValidString(value) && !strings.ContainsRune(value, 0)
}

func validHash(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}

func uniqueText(values []string, maximum int) error {
	if len(values) > maximum {
		return fmt.Errorf("too many metadata values")
	}
	seen := make(map[string]bool)
	for _, value := range values {
		if !text(value) || seen[value] {
			return fmt.Errorf("invalid or duplicate metadata value %q", value)
		}
		seen[value] = true
	}
	return nil
}

func choices(values, allowed []string, maximum int) error {
	if len(values) == 0 {
		return fmt.Errorf("at least one value is required")
	}
	if err := uniqueText(values, maximum); err != nil {
		return err
	}
	for _, value := range values {
		if !slices.Contains(allowed, value) {
			return fmt.Errorf("unknown value %q", value)
		}
	}
	return nil
}

func partitions() []string { return []string{"training", "development", "calibration", "final_test"} }

func roles() []string {
	return []string{"documentation", "readme", "api_reference", "doc_comment", "comment", "release_note",
		"string", "error_message", "log_message", "ui_text", "other_string", "unknown"}
}

// cohorts lists the protocol's cohorts. The dated historical periods are the
// placebo pseudo-boundaries of December 31, 2012 and 2016 and the H1
// sensitivity boundary of December 31, 2018; "historical" is H0.
func cohorts() []string {
	return []string{"historical", "historical-2012", "historical-2016", "historical-2018", "controlled", "natural", "contemporary"}
}

func dateConfidences() []string { return []string{"corroborated", "vcs_only", "unknown"} }

func digest(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return hashBytes(data), nil
}

func hashBytes(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
