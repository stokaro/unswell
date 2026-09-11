// Package corpus prepares unlabeled research units from frozen source groups.
// Its callers supply source bytes; the package never reads paths or fetches URLs.
package corpus

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation"
)

// Version identifies the source manifest and candidate artifact contract.
const Version = "unswell-corpus-v1"

// Resource limits bound preparation without defining scientific sample adequacy.
const (
	MaxManifestBytes = 16 << 20
	MaxArtifactBytes = 128 << 20
	MaxSourceBytes   = 2 << 20
	MaxTotalBytes    = 64 << 20
	MaxSources       = 10000
	MaxUnits         = 10000
	// MaxRuleBlocks bounds the blocks a rule collection records. Every block
	// of a source carries activations. A corpus at the unit limit has three
	// to four blocks per candidate, so the bound is five per unit.
	// MaxRuleCollectionBytes bounds the collection itself. Forty activations
	// per block over such a corpus take most of the artifact. The decisions
	// and bindings fill the rest.
	MaxRuleBlocks          = 5 * MaxUnits
	MaxRuleCollectionBytes = MaxArtifactBytes / 4 * 3
)

// Manifest freezes acquisition assertions before source grouping and extraction.
type Manifest struct {
	Version   string         `json:"version"`
	ID        string         `json:"id"`
	Seed      string         `json:"seed"`
	Weights   Weights        `json:"weights"`
	Policy    extract.Policy `json:"extraction_policy"`
	UnitKinds []string       `json:"unit_kinds"`
	Sources   []Source       `json:"sources"`
}

// Weights gives each partition a positive share out of 10000 hash buckets.
type Weights struct {
	Training    int `json:"training"`
	Development int `json:"development"`
	Calibration int `json:"calibration"`
	FinalTest   int `json:"final_test"`
}

// Source records one exact file and the curator's metadata. Group IDs must be
// globally scoped; empty optional grouping lists mean unknown relationships.
type Source struct {
	ID                  string            `json:"id"`
	Path                string            `json:"path"`
	SHA256              string            `json:"sha256"`
	Bytes               int               `json:"bytes"`
	Format              document.Format   `json:"format"`
	ProseLanguage       string            `json:"prose_language"`
	Repository          string            `json:"repository"`
	Document            string            `json:"document"`
	Authors             []string          `json:"authors"`
	Templates           []string          `json:"templates"`
	Related             []string          `json:"related"`
	GenerationTasks     []string          `json:"generation_tasks"`
	Reference           string            `json:"reference"`
	Topic               string            `json:"topic"`
	Purpose             string            `json:"purpose"`
	AuthorLanguage      string            `json:"author_language"`
	AuthorLanguageBasis string            `json:"author_language_basis"`
	Role                string            `json:"role"`
	Roles               []RoleRegion      `json:"role_regions"`
	Origin              annotation.Origin `json:"origin"`
	Rights              annotation.Rights `json:"rights"`
	Notices             []Notice          `json:"notices"`
	Partition           string            `json:"partition"`
	Snapshot            *Snapshot         `json:"snapshot,omitempty"`
}

// Snapshot dates the exact bytes of a source and names its research cohort.
// A date never proves authorship; Confidence says what corroborates it. A
// source without a snapshot has no cohort and enters no temporal analysis.
type Snapshot struct {
	Date       string `json:"date"`
	Confidence string `json:"confidence"`
	Evidence   string `json:"evidence"`
	Cohort     string `json:"cohort"`
}

// Notice identifies a local permission/copyright record retained with the source.
type Notice struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

// RoleRegion gives a curator-reviewed role to complete units within an original
// source range. Partial intersections and overlapping regions are rejected.
type RoleRegion struct {
	Span document.Span `json:"span"`
	Role string        `json:"role"`
}

// Plan is a reproducible grouping result. Its manifest remains the authority;
// consumers recompute assignments before accepting this artifact.
type Plan struct {
	Version        string   `json:"version"`
	Algorithm      string   `json:"algorithm"`
	Manifest       Manifest `json:"manifest"`
	ManifestSHA256 string   `json:"manifest_sha256"`
	Groups         []Group  `json:"groups"`
}

// Group records an inseparable connected component and its partition.
type Group struct {
	ID        string   `json:"id"`
	Sources   []string `json:"sources"`
	Keys      []string `json:"keys"`
	Partition string   `json:"partition"`
	Pinned    bool     `json:"pinned"`
}

// Pipeline identifies source extraction, unit/context formation, and NLP models.
type Pipeline struct {
	Version      string       `json:"version"`
	Revision     string       `json:"revision"`
	Modified     bool         `json:"modified"`
	GoVersion    string       `json:"go_version"`
	Dependencies []Dependency `json:"dependencies"`
	NLP          nlp.Identity `json:"nlp"`
}

// Dependency pins the module content used by the compiled preparation tool.
type Dependency struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Sum     string `json:"sum"`
}

// Candidate is unlabeled prose with acquisition and split identities. Cohort
// repeats the source snapshot's cohort and is empty without a snapshot.
type Candidate struct {
	SourceID  string          `json:"source_id"`
	GroupID   string          `json:"group_id"`
	Partition string          `json:"partition"`
	Cohort    string          `json:"cohort,omitempty"`
	Words     int             `json:"words"`
	Unit      annotation.Unit `json:"unit"`
}

// SourceResult retains exclusions and empty-source evidence alongside candidates.
type SourceResult struct {
	ID          string               `json:"id"`
	Blocks      int                  `json:"blocks"`
	Units       int                  `json:"units"`
	Excluded    []document.Exclusion `json:"excluded"`
	EmptyReason string               `json:"empty_reason"`
}

// Artifact is a complete candidate preparation, never a human-labeled corpus.
// SHA256 covers compact Go JSON with that field omitted.
type Artifact struct {
	Version  string         `json:"version"`
	Status   string         `json:"status"`
	SHA256   string         `json:"sha256,omitempty"`
	Plan     Plan           `json:"plan"`
	Pipeline Pipeline       `json:"pipeline"`
	Sources  []SourceResult `json:"sources"`
	Units    []Candidate    `json:"units"`
}
