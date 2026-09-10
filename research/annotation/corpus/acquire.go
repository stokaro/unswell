package corpus

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// AcquisitionVersion identifies the acquisition record and its selection rules.
const AcquisitionVersion = "unswell-corpus-acquisition-v1"

// Acquisition limits keep one repository within a shard's limits.
const (
	MaxAcquisitionBytes = 1 << 20
	MaxAcquiredSources  = 2000
)

// Acquisition freezes what a curator asserts about one repository snapshot
// before any file is read: identity, rights, provenance, dating evidence, and
// the selection rules. The tooling that produced it did the network work;
// the library only applies the rules to bytes it is given.
type Acquisition struct {
	Version    string            `json:"version"`
	Manifest   AcquisitionHeader `json:"manifest"`
	Repository RepositoryRecord  `json:"repository"`
	Selection  SelectionRules    `json:"selection"`
}

// AcquisitionHeader repeats the dataset header the shard must carry.
type AcquisitionHeader struct {
	ID        string         `json:"id"`
	Seed      string         `json:"seed"`
	Weights   Weights        `json:"weights"`
	Policy    extract.Policy `json:"extraction_policy"`
	UnitKinds []string       `json:"unit_kinds"`
}

// RepositoryRecord names the snapshot and the assertions every source repeats.
type RepositoryRecord struct {
	Name      string            `json:"name"`
	Reference string            `json:"reference"`
	Commit    string            `json:"commit"`
	Topic     string            `json:"topic"`
	Purpose   string            `json:"purpose"`
	Origin    annotation.Origin `json:"origin"`
	Rights    annotation.Rights `json:"rights"`
	Notices   []string          `json:"notices"`
	Snapshot  Snapshot          `json:"snapshot"`
	Ecosystem string            `json:"ecosystem"`
}

// SelectionRules are the predefined inclusion and exclusion rules. They are
// fixed before extraction and never depend on what a rule finds.
type SelectionRules struct {
	MaxSources        int      `json:"max_sources"`
	ShardSources      int      `json:"shard_sources"`
	ShardBytes        int      `json:"shard_bytes"`
	MaxSourceBytes    int      `json:"max_source_bytes"`
	DocumentRoots     []string `json:"document_roots"`
	SourceExtensions  []string `json:"source_extensions"`
	ExcludedSegments  []string `json:"excluded_segments"`
	ExcludedBasenames []string `json:"excluded_basenames"`
	GeneratedMarkers  []string `json:"generated_markers"`
	TranslationHints  []string `json:"translation_hints"`
}

// AcquisitionResult is the shard manifests plus the record of every file the
// rules left out, so a reviewer can see what the selection did. A repository
// with more selected sources than ShardSources spans several manifests whose
// IDs carry a numbered suffix; they share the repository key, so the dataset
// plan keeps them in one group.
type AcquisitionResult struct {
	Version   string                 `json:"version"`
	Manifests []Manifest             `json:"manifests"`
	Selected  int                    `json:"selected"`
	Excluded  []AcquisitionExclusion `json:"excluded"`
	Subsample *Subsample             `json:"subsample,omitempty"`
}

// AcquisitionExclusion names one file and the rule that excluded it.
type AcquisitionExclusion struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// Subsample records a seeded reduction to the source cap.
type Subsample struct {
	Seed    string   `json:"seed"`
	Before  int      `json:"before"`
	After   int      `json:"after"`
	Dropped []string `json:"dropped"`
}

// LoadAcquisition decodes strict JSON and validates the record.
func LoadAcquisition(ctx context.Context, data []byte) (Acquisition, error) {
	var record Acquisition
	if err := jsoninput.Decode(ctx, data, MaxAcquisitionBytes, &record, jsoninput.Limits{Array: 1000, Object: 64}); err != nil {
		return Acquisition{}, err
	}
	if err := record.validate(); err != nil {
		return Acquisition{}, err
	}
	return record, nil
}

func (a Acquisition) validate() error {
	if a.Version != AcquisitionVersion {
		return fmt.Errorf("unsupported acquisition version")
	}
	if err := a.Manifest.validate(); err != nil {
		return err
	}
	if err := a.Repository.validate(); err != nil {
		return err
	}
	return a.Selection.validate()
}

func (h AcquisitionHeader) validate() error {
	if !text(h.ID) || !text(h.Seed) {
		return fmt.Errorf("invalid manifest ID or seed")
	}
	if err := h.Weights.validate(); err != nil {
		return err
	}
	if err := extract.ValidatePolicy(h.Policy); err != nil {
		return err
	}
	return choices(h.UnitKinds, []string{"sentence", "paragraph", "fragment"}, 3)
}

func (r RepositoryRecord) validate() error {
	for _, value := range []string{r.Name, r.Reference, r.Commit, r.Topic, r.Purpose, r.Ecosystem} {
		if !text(value) {
			return fmt.Errorf("repository name, reference, commit, topic, purpose, and ecosystem are required")
		}
	}
	if len(r.Notices) == 0 || len(r.Notices) > 20 {
		return fmt.Errorf("one through 20 notice paths are required")
	}
	for _, notice := range r.Notices {
		if !ValidPath(notice) {
			return fmt.Errorf("invalid notice path %q", notice)
		}
	}
	return nil
}

func (s SelectionRules) validate() error {
	if err := s.validateCaps(); err != nil {
		return err
	}
	for _, group := range [][]string{s.DocumentRoots, s.SourceExtensions, s.ExcludedSegments, s.ExcludedBasenames,
		s.GeneratedMarkers, s.TranslationHints} {
		if err := uniqueText(group, 200); err != nil {
			return err
		}
	}
	return nil
}

func (s SelectionRules) validateCaps() error {
	caps := []struct {
		value, limit int
		name         string
	}{
		{s.MaxSources, MaxAcquiredSources, "max_sources"},
		{s.MaxSourceBytes, MaxSourceBytes, "max_source_bytes"},
		{s.ShardSources, MaxSources, "shard_sources"},
		{s.ShardBytes, MaxTotalBytes, "shard_bytes"},
	}
	for _, cap := range caps {
		if cap.value <= 0 || cap.value > cap.limit {
			return fmt.Errorf("%s must be positive and at most %d", cap.name, cap.limit)
		}
	}
	return nil
}

// Acquire applies the record's rules to the files of a checkout and returns
// the shard manifest with every exclusion. Files are keyed by slash paths
// relative to the checkout root. Notices must be present and are never
// selected as sources.
func Acquire(ctx context.Context, record Acquisition, files map[string][]byte) (AcquisitionResult, error) {
	if err := record.validate(); err != nil {
		return AcquisitionResult{}, err
	}
	notices, err := acquisitionNotices(record, files)
	if err != nil {
		return AcquisitionResult{}, err
	}
	result := AcquisitionResult{Version: AcquisitionVersion, Excluded: []AcquisitionExclusion{}}
	candidates, err := selectSources(ctx, record, files, notices, &result)
	if err != nil {
		return AcquisitionResult{}, err
	}
	candidates, result.Subsample = subsampleSources(record, candidates)
	for _, dropped := range subsampleDropped(result.Subsample) {
		result.Excluded = append(result.Excluded, AcquisitionExclusion{Path: dropped, Reason: "seeded_subsample"})
	}
	slices.SortFunc(result.Excluded, func(a, b AcquisitionExclusion) int { return strings.Compare(a.Path, b.Path) })
	if len(candidates) == 0 {
		return AcquisitionResult{}, fmt.Errorf("no file of %s passed the selection rules", record.Repository.Name)
	}
	result.Selected = len(candidates)
	result.Manifests = shardManifests(record, candidates)
	for _, manifest := range result.Manifests {
		if err := manifest.validate(ctx); err != nil {
			return AcquisitionResult{}, err
		}
	}
	return result, nil
}

// shardManifests cuts the sorted sources into manifests that stay within
// ShardSources files and ShardBytes of source text, which keeps the unit
// count of one extraction within the artifact limit. One manifest keeps the
// bare ID; several get a numbered suffix.
func shardManifests(record Acquisition, sources []Source) []Manifest {
	var chunks [][]Source
	var current []Source
	bytes := 0
	for _, source := range sources {
		full := len(current) >= record.Selection.ShardSources || bytes+source.Bytes > record.Selection.ShardBytes
		if full && len(current) > 0 {
			chunks = append(chunks, current)
			current, bytes = nil, 0
		}
		current = append(current, source)
		bytes += source.Bytes
	}
	if len(current) > 0 {
		chunks = append(chunks, current)
	}
	manifests := make([]Manifest, 0, len(chunks))
	for i, chunk := range chunks {
		id := record.Manifest.ID
		if len(chunks) > 1 {
			id = fmt.Sprintf("%s-%03d", id, i+1)
		}
		manifests = append(manifests, Manifest{Version: Version, ID: id, Seed: record.Manifest.Seed,
			Weights: record.Manifest.Weights, Policy: record.Manifest.Policy, UnitKinds: record.Manifest.UnitKinds,
			Sources: chunk})
	}
	return manifests
}

// selectSources applies the rules to every file in path order and records
// each exclusion on the result.
func selectSources(ctx context.Context, record Acquisition, files map[string][]byte, notices []Notice,
	result *AcquisitionResult,
) ([]Source, error) {
	paths := make([]string, 0, len(files))
	for name := range files {
		paths = append(paths, name)
	}
	sort.Strings(paths)
	var candidates []Source
	for _, name := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		source, reason := selectSource(record, name, files[name], notices)
		if reason != "" {
			result.Excluded = append(result.Excluded, AcquisitionExclusion{Path: name, Reason: reason})
			continue
		}
		candidates = append(candidates, source)
	}
	return candidates, nil
}

func acquisitionNotices(record Acquisition, files map[string][]byte) ([]Notice, error) {
	notices := make([]Notice, 0, len(record.Repository.Notices))
	for _, name := range record.Repository.Notices {
		data, found := files[name]
		if !found || len(data) == 0 || len(data) > MaxSourceBytes {
			return nil, fmt.Errorf("notice %s is missing or outside the byte limit", name)
		}
		notices = append(notices, Notice{Path: name, SHA256: hashBytes(data), Bytes: len(data)})
	}
	return notices, nil
}

// selectSource decides one file's role and format or names the rule that
// excludes it. The rules never look at rule outcomes.
func selectSource(record Acquisition, name string, data []byte, notices []Notice) (Source, string) {
	rules := record.Selection
	if slices.ContainsFunc(notices, func(n Notice) bool { return n.Path == name }) {
		return Source{}, "notice"
	}
	if !ValidPath(name) {
		return Source{}, "unportable_path"
	}
	if len(data) == 0 || len(data) > rules.MaxSourceBytes {
		return Source{}, "outside_byte_limit"
	}
	if !utf8.Valid(data) {
		return Source{}, "not_utf8"
	}
	base := path.Base(name)
	if reason := excludedPath(rules, name, base); reason != "" {
		return Source{}, reason
	}
	role, format, ok := roleAndFormat(rules, name, base, data)
	if !ok {
		return Source{}, "unselected_kind"
	}
	if hasMarker(data, rules.GeneratedMarkers) {
		return Source{}, "generated_marker"
	}
	// A file the frozen extraction policy cannot parse would fail the whole
	// shard later; it is excluded here with its reason instead.
	source := sourceFor(record, name, data, role, format, notices)
	if _, err := extract.Parse(context.Background(), document.Source{Name: name, Format: format, Bytes: data},
		extract.Options{MaxBytes: MaxSourceBytes, MaxBlocks: MaxUnits, Policy: record.Manifest.Policy}); err != nil {
		return Source{}, "parse_failure"
	}
	return source, ""
}

func excludedPath(rules SelectionRules, name, base string) string {
	for segment := range strings.SplitSeq(name, "/") {
		if slices.Contains(rules.ExcludedSegments, segment) {
			return "excluded_segment"
		}
		lowered := strings.ToLower(segment)
		if slices.Contains(rules.TranslationHints, lowered) {
			return "translation_hint"
		}
	}
	upper := strings.ToUpper(base)
	for _, excluded := range rules.ExcludedBasenames {
		if upper == strings.ToUpper(excluded) || strings.HasPrefix(upper, strings.ToUpper(excluded)+".") {
			return "excluded_basename"
		}
	}
	return ""
}

func roleAndFormat(rules SelectionRules, name, base string, data []byte) (string, document.Format, bool) {
	format, ok := extract.Detect(name, data)
	if !ok {
		return "", "", false
	}
	if format == document.Markdown || format == document.Plain {
		role := documentRole(rules, name, base)
		return role, format, role != ""
	}
	if slices.Contains(rules.SourceExtensions, strings.ToLower(path.Ext(base))) {
		return "comment", format, true
	}
	return "", "", false
}

// documentRole names a prose file's role, or nothing when the file lies
// outside every documentation root and the top level.
func documentRole(rules SelectionRules, name, base string) string {
	upper := strings.ToUpper(strings.TrimSuffix(base, path.Ext(base)))
	switch {
	case upper == "README":
		return "readme"
	case slices.Contains([]string{"CHANGELOG", "CHANGES", "HISTORY", "NEWS", "RELEASES"}, upper):
		return "release_note"
	case underRoot(rules.DocumentRoots, name) || !strings.Contains(name, "/"):
		return "documentation"
	}
	return ""
}

func underRoot(roots []string, name string) bool {
	for _, root := range roots {
		if strings.HasPrefix(name, root+"/") {
			return true
		}
	}
	return false
}

func hasMarker(data []byte, markers []string) bool {
	head := data
	if len(head) > 4096 {
		head = head[:4096]
	}
	text := string(head)
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func sourceFor(record Acquisition, name string, data []byte, role string, format document.Format, notices []Notice) Source {
	r := record.Repository
	snapshot := r.Snapshot
	return Source{ID: sourceID(r.Name, name), Path: name, SHA256: hashBytes(data), Bytes: len(data), Format: format,
		ProseLanguage: "en", Repository: r.Name, Document: r.Name + "/" + name, Authors: []string{}, Templates: []string{},
		Related: []string{}, GenerationTasks: []string{}, Reference: strings.TrimSuffix(r.Reference, "/") + "/" + name,
		Topic: r.Topic, Purpose: r.Purpose, Role: role, Roles: []RoleRegion{}, Origin: r.Origin, Rights: r.Rights,
		Notices: slices.Clone(notices), Snapshot: &snapshot}
}

func sourceID(repository, name string) string {
	sum := sha256.Sum256([]byte(repository + "\x00" + name))
	return fmt.Sprintf("%x", sum[:12])
}

// subsampleSources keeps at most the cap by a seeded, order-independent draw
// over whole files; the seed and the dropped paths are recorded.
func subsampleSources(record Acquisition, sources []Source) ([]Source, *Subsample) {
	limit := record.Selection.MaxSources
	if len(sources) <= limit {
		return sources, nil
	}
	seed := record.Manifest.Seed + "\x00" + record.Repository.Name
	type ranked struct {
		key    uint64
		source Source
	}
	items := make([]ranked, 0, len(sources))
	for _, source := range sources {
		sum := sha256.Sum256([]byte(seed + "\x00" + source.Path))
		items = append(items, ranked{key: binary.BigEndian.Uint64(sum[:8]), source: source})
	}
	sort.Slice(items, func(a, b int) bool {
		if items[a].key != items[b].key {
			return items[a].key < items[b].key
		}
		return items[a].source.Path < items[b].source.Path
	})
	summary := &Subsample{Seed: seed, Before: len(sources), After: limit, Dropped: []string{}}
	kept := make([]Source, 0, limit)
	for i, item := range items {
		if i < limit {
			kept = append(kept, item.source)
		} else {
			summary.Dropped = append(summary.Dropped, item.source.Path)
		}
	}
	slices.Sort(summary.Dropped)
	slices.SortFunc(kept, func(a, b Source) int { return strings.Compare(a.Path, b.Path) })
	return kept, summary
}

func subsampleDropped(record *Subsample) []string {
	if record == nil {
		return nil
	}
	return record.Dropped
}
