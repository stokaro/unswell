package unswell

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

// Version is the release identity used in canonical results.
const Version = "0.1.0-alpha.1"

// BuildCommit identifies release builds. Set it with -ldflags at build time.
// Applications should not change this value while engines are running.
var BuildCommit = "development"

// SchemaVersion versions the saved result contract independently of CLI flags.
const SchemaVersion = "1.0.0-alpha.1"

// Location contains original byte spans and Unicode code point positions.
type Location struct {
	Path     string            `json:"path"`
	Span     document.Span     `json:"span"`
	Start    document.Position `json:"start"`
	End      document.Position `json:"end"`
	Segments []document.Span   `json:"segments"`
	Snippet  string            `json:"snippet,omitempty"`
}

// Finding is a rule activation enriched with policy, identity, and source locations.
type Finding struct {
	ID            string        `json:"id"`
	RuleID        string        `json:"rule_id"`
	RuleVersion   string        `json:"rule_version"`
	Severity      string        `json:"severity"`
	Gate          string        `json:"gate"`
	Group         string        `json:"group"`
	Scope         string        `json:"scope"`
	Message       string        `json:"message"`
	Primary       Location      `json:"primary"`
	Related       []Location    `json:"related"`
	Evidence      rule.Evidence `json:"evidence"`
	Fingerprint   string        `json:"fingerprint"`
	Suppressed    bool          `json:"suppressed"`
	BaselineState string        `json:"baseline_state"`
	Derived       bool          `json:"derived"`
}

// Contribution explains fixed-point scoring and the effects of caps and deduplication.
type Contribution struct {
	FindingID  string  `json:"finding_id"`
	RuleID     string  `json:"rule_id"`
	Group      string  `json:"group"`
	Weight     int     `json:"weight"`
	Activation int     `json:"activation"`
	Raw        float64 `json:"raw"`
	Effective  float64 `json:"effective"`
	RuleCap    int     `json:"rule_cap"`
	GroupCap   int     `json:"group_cap"`
	Reason     string  `json:"reason"`
}

// Assessment is a local, nondilutable index. Probability is unavailable in alpha.
type Assessment struct {
	Path               string         `json:"path"`
	Scope              string         `json:"scope"`
	UnitID             int            `json:"unit_id"`
	Span               document.Span  `json:"span"`
	Words              int            `json:"words"`
	SlopScore          float64        `json:"slop_score"`
	EffectiveSlopScore float64        `json:"effective_slop_score"`
	SlopProbability    *float64       `json:"slop_probability"`
	ProbabilityStatus  string         `json:"probability_status"`
	Status             string         `json:"status"`
	Contributions      []Contribution `json:"contributions"`
}

// DocumentResult summarizes one source without including its prose by default.
type DocumentResult struct {
	Name            string               `json:"name"`
	Format          document.Format      `json:"format"`
	SourceHash      string               `json:"source_hash"`
	Bytes           int                  `json:"bytes"`
	ProseWords      int                  `json:"prose_words"`
	Blocks          int                  `json:"blocks"`
	Sentences       int                  `json:"sentences"`
	Excluded        []document.Exclusion `json:"excluded"`
	Source          string               `json:"source,omitempty"`
	Maximum         float64              `json:"paragraph_maximum"`
	Median          float64              `json:"paragraph_median"`
	P90             float64              `json:"paragraph_p90"`
	FlaggedFraction float64              `json:"eligible_paragraph_flagged_fraction"`
}

// Manifest records the exact execution policy and feature versions, without time.
type Manifest struct {
	ToolVersion     string            `json:"tool_version"`
	ToolCommit      string            `json:"tool_commit"`
	ConfigHash      string            `json:"config_hash"`
	RulesetHash     string            `json:"ruleset_hash"`
	ScoringProfile  string            `json:"scoring_profile"`
	FeatureContract string            `json:"feature_contract"`
	NLP             nlp.Identity      `json:"nlp"`
	Rules           []rule.Descriptor `json:"rules"`
	SelectionMode   string            `json:"selection_mode"`
	Complete        bool              `json:"complete"`
	NoGate          bool              `json:"no_gate"`
	IncludeSource   bool              `json:"include_source"`
	SkippedRules    []string          `json:"skipped_rules"`
}

// GateReason identifies the exact local condition behind a policy failure.
type GateReason struct {
	Code      string `json:"code"`
	Path      string `json:"path"`
	FindingID string `json:"finding_id"`
	Message   string `json:"message"`
}

// GateDecision reports policy separately from operational completion.
type GateDecision struct {
	Passed  bool         `json:"passed"`
	Reasons []GateReason `json:"reasons"`
}

// RunError records an operational failure in a partial result.
type RunError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// RunResult is the immutable input shared by every reporter.
// Returned slices are owned by the caller and do not alias the engine.
type RunResult struct {
	SchemaVersion string           `json:"schema_version"`
	Status        string           `json:"status"`
	Manifest      Manifest         `json:"manifest"`
	Documents     []DocumentResult `json:"documents"`
	Findings      []Finding        `json:"findings"`
	Assessments   []Assessment     `json:"assessments"`
	Gate          GateDecision     `json:"gate"`
	Errors        []RunError       `json:"errors"`
}

// Result is the same report contract restricted to one source by Analyze.
type Result = RunResult
