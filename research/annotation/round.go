// Package annotation defines the research annotation protocol, not a product model.
package annotation

import "github.com/stokaro/unswell/document"

// Version identifies the annotation artifact contract.
const Version = "unswell-annotation-v1"

// Rubric identifies the published editorial decision rubric.
const Rubric = "unswell-editorial-v1"

// Round preserves administrative metadata, original judgments, and adjudication.
// Only Packet may be distributed to blinded annotators.
type Round struct {
	data        roundData
	inputSHA256 string
}

type roundData struct {
	packetSHA256  string
	Version       string         `json:"version"`
	ID            string         `json:"round_id"`
	Purpose       string         `json:"purpose"`
	Rubric        string         `json:"rubric"`
	Profile       Profile        `json:"profile"`
	Units         []Unit         `json:"units"`
	Actors        []Actor        `json:"actors"`
	Judgments     []Judgment     `json:"judgments"`
	Adjudications []Adjudication `json:"adjudications"`
}

// Profile freezes the editorial policy shown to annotators.
type Profile struct {
	ID           string `json:"id"`
	Instructions string `json:"instructions"`
	SHA256       string `json:"sha256"`
}

// Unit contains extracted prose and references to the original source segments.
// Corpus acquisition must verify the source hash and extraction mapping separately.
type Unit struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	Role       string     `json:"role"`
	Text       string     `json:"text"`
	Context    string     `json:"context"`
	Source     Source     `json:"source"`
	Extraction Extraction `json:"extraction"`
	Origin     Origin     `json:"origin"`
	Rights     Rights     `json:"rights"`
}

// Source records acquisition identity without claiming that metadata proves authorship.
type Source struct {
	DocumentID    string          `json:"document_id"`
	RepositoryID  string          `json:"repository_id"`
	TemplateID    string          `json:"template_id"`
	AuthorGroup   string          `json:"author_group"`
	RelatedGroup  string          `json:"related_group"`
	Reference     string          `json:"reference"`
	SHA256        string          `json:"sha256"`
	Bytes         int             `json:"bytes"`
	Language      document.Format `json:"language"`
	ProseLanguage string          `json:"prose_language"`
	Segments      []document.Span `json:"segments"`
}

// Extraction identifies the existing extractor and the selected context policy.
type Extraction struct {
	Identity      string `json:"identity"`
	PolicySHA256  string `json:"policy_sha256"`
	ContextPolicy string `json:"context_policy"`
}

// Origin is a separately curated claim with evidence at a declared scope.
type Origin struct {
	Label            string `json:"label"`
	Scope            string `json:"scope"`
	Evidence         string `json:"evidence"`
	GenerationRecord string `json:"generation_record"`
}

// Rights records reviewed permissions; a syntactically valid record is not legal approval.
type Rights struct {
	License     string   `json:"license"`
	Evidence    string   `json:"evidence"`
	AllowedUses []string `json:"allowed_uses"`
}

// Actor is an opaque identifier and a curator's declaration of participation.
type Actor struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Role        string `json:"role"`
	Declaration string `json:"declaration"`
}

// Judgment is one immutable, independent response before adjudication.
// No record means a missing answer; uncertain is an actual nominal label.
type Judgment struct {
	PacketSHA256 string   `json:"packet_sha256"`
	UnitID       string   `json:"unit_id"`
	ActorID      string   `json:"actor_id"`
	Label        string   `json:"label"`
	Categories   []string `json:"categories"`
	Context      string   `json:"context"`
	Rationale    string   `json:"rationale"`
	RecordedAt   string   `json:"recorded_at"`
}

// Adjudication records a later decision without replacing the original judgments.
type Adjudication struct {
	PacketSHA256 string   `json:"packet_sha256"`
	UnitID       string   `json:"unit_id"`
	Reviewers    []string `json:"reviewers"`
	Label        string   `json:"label"`
	Categories   []string `json:"categories"`
	Rationale    string   `json:"rationale"`
	RecordedAt   string   `json:"recorded_at"`
}

// Categories returns the rubric's reason IDs in a stable order.
func Categories() []string {
	return []string{"empty_framing", "needless_repetition", "wordiness", "unjustified_intensifiers",
		"vague_claims", "formulaic_transitions", "needless_complexity"}
}
