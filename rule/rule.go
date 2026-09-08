// Package rule defines explicit rule registration, evidence, and configuration.
package rule

import (
	"context"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

// Parameters contains the typed matcher controls supported by the alpha catalog.
// A descriptor explicitly lists the fields it accepts. External Go rules can
// receive additional typed configuration through their own constructors.
type Parameters struct {
	Phrases               []string `json:"phrases"                yaml:"phrases"`
	Positions             []string `json:"positions"              yaml:"positions"`
	MinWords              int      `json:"min_words"              yaml:"min_words"`
	Similarity            float64  `json:"similarity"             yaml:"similarity"`
	Window                string   `json:"window"                 yaml:"window"`
	WindowSentences       int      `json:"window_sentences"       yaml:"window_sentences"`
	AllowedOccurrences    int      `json:"allowed_occurrences"    yaml:"allowed_occurrences"`
	SaturationOccurrences int      `json:"saturation_occurrences" yaml:"saturation_occurrences"`
	Onset                 int      `json:"onset"                  yaml:"onset"`
	Saturation            int      `json:"saturation"             yaml:"saturation"`
	OpenerWords           int      `json:"opener_words"           yaml:"opener_words"`
	ProtectNegation       bool     `json:"protect_negation"       yaml:"protect_negation"`
	ProtectNumbers        bool     `json:"protect_numbers"        yaml:"protect_numbers"`
	ProtectIdentifiers    bool     `json:"protect_identifiers"    yaml:"protect_identifiers"`
	MaxAnswerWords        int      `json:"max_answer_words,omitempty" yaml:"max_answer_words"`
	MinNgramWords         int      `json:"min_ngram_words,omitempty" yaml:"min_ngram_words"`
	MaxNgramWords         int      `json:"max_ngram_words,omitempty" yaml:"max_ngram_words"`
	WindowBlocks          int      `json:"window_blocks,omitempty" yaml:"window_blocks"`
	Verbs                 []string `json:"verbs,omitempty" yaml:"verbs"`
	Nouns                 []string `json:"nouns,omitempty" yaml:"nouns"`
	MinSentences          int      `json:"min_sentences,omitempty" yaml:"min_sentences"`
	SentenceWords         int      `json:"sentence_words,omitempty" yaml:"sentence_words"`
	MinLongSentences      int      `json:"min_long_sentences,omitempty" yaml:"min_long_sentences"`
	AllowedDepth          int      `json:"allowed_depth,omitempty" yaml:"allowed_depth"`
	SaturationDepth       int      `json:"saturation_depth,omitempty" yaml:"saturation_depth"`
	MaxItemWords          int      `json:"max_item_words,omitempty" yaml:"max_item_words"`
	MaxListItems          int      `json:"max_list_items,omitempty" yaml:"max_list_items"`
}

// Score is independent of presentation severity. Units are index points.
type Score struct {
	Weight int `json:"weight" yaml:"weight"`
	Cap    int `json:"cap"    yaml:"cap"`
}

// Settings controls execution, presentation, unconditional policy, and scoring.
type Settings struct {
	Enabled    bool       `json:"enabled"    yaml:"enabled"`
	Severity   string     `json:"severity"   yaml:"severity"`
	Gate       string     `json:"gate"       yaml:"gate"`
	Score      Score      `json:"score"      yaml:"score"`
	Parameters Parameters `json:"parameters" yaml:"parameters"`
}

// Example is an executable catalog fixture. An empty Format selects plain prose.
type Example struct {
	Text   string          `json:"text"`
	Match  bool            `json:"match"`
	Config string          `json:"config,omitempty"`
	Format document.Format `json:"format,omitempty"`
}

// Descriptor documents a versioned, independently useful editorial signal.
type Descriptor struct {
	ID             string           `json:"id"`
	Version        string           `json:"version"`
	Summary        string           `json:"summary"`
	Description    string           `json:"description"`
	Limitations    string           `json:"limitations"`
	Scope          string           `json:"scope"`
	Contexts       []string         `json:"contexts"`
	Requires       []nlp.Capability `json:"requires"`
	Group          string           `json:"group"`
	Status         string           `json:"status"`
	Defaults       Settings         `json:"defaults"`
	Parameters     []string         `json:"parameters"`
	Examples       []Example        `json:"examples"`
	Origin         *Origin          `json:"origin,omitempty"`
	TermExemptions bool             `json:"term_exemptions,omitempty"`
	// SharedFeatures requests the common block measurements in View.Features.
	// POS values still require an explicit POS capability in Requires.
	SharedFeatures bool `json:"shared_features,omitempty"`
	// RequiresStructure requests grammar-derived block context, including heading
	// ancestry. It does not restore prose excluded by the extraction policy.
	RequiresStructure bool `json:"requires_structure,omitempty"`
	// DependencyScheme requires an exact provider label scheme. Requires must
	// include nlp.Dependencies. Empty permits scheme-independent graph analysis.
	DependencyScheme string `json:"dependency_scheme,omitempty"`
}

// Origin identifies a declarative ruleset and its author-supplied provenance.
// Hash covers the complete validated definition; license and provenance are
// declarations, not endorsements or automatically verified permissions.
type Origin struct {
	Namespace  string `json:"namespace"`
	Version    string `json:"version"`
	License    string `json:"license"`
	Provenance string `json:"provenance"`
	Hash       string `json:"sha256"`
}

// Metric is measurable evidence, with a named unit and activation thresholds.
type Metric struct {
	Name       string  `json:"name"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Onset      float64 `json:"onset"`
	Saturation float64 `json:"saturation"`
}

// Occurrence links a finding to each affected sentence and structural block.
type Occurrence struct {
	BlockID    int             `json:"block_id"`
	SentenceID int             `json:"sentence_id"`
	Spans      []document.Span `json:"spans"`
}

// Evidence is emitted once for a cluster, with all of its occurrences.
type Evidence struct {
	Kind        string       `json:"kind"`
	Message     string       `json:"message"`
	Suggestion  string       `json:"suggestion"`
	Occurrences []Occurrence `json:"occurrences"`
	Metrics     []Metric     `json:"metrics"`
	Activation  int          `json:"activation"` // Fixed point: 0 through 1000.
}

// View is read-only for the duration of Evaluate. Do not retain its document or
// term pointers. Features is independently owned and immutable and may be retained.
type View struct {
	Document       *document.Document
	Features       *feature.Set
	Parameters     Parameters
	MaxCandidates  int
	TermExemptions *TermMatches
}

// TokenRange identifies a half-open token range in one sentence and block.
type TokenRange struct{ BlockID, SentenceID, Start, End int }

// Exempts reports whether an entire candidate lies inside one approved term.
// Rules call this before counting or aggregating evidence, only when supported.
func (v View) Exempts(sentence document.Sentence, start, end int) bool {
	if start < 0 || end <= start || end > len(sentence.Tokens) {
		return false
	}
	return v.TermExemptions.contains(sentence, start, end)
}

// Emitter validates evidence. An emitter error makes the entire analysis incomplete.
type Emitter interface{ Emit(Evidence) error }

// Rule is trusted Go code. Implementations must support concurrent evaluation.
type Rule interface {
	Descriptor() Descriptor
	Evaluate(context.Context, View, Emitter) error
}
