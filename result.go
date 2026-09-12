package unswell

import (
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

// Version is the release identity used in canonical results.
const Version = "0.1.0-alpha.2"

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
	ChangeState         string        `json:"change_state,omitempty"`
	ChangeFingerprint   string        `json:"change_fingerprint,omitempty"`
	ID                  string        `json:"id"`
	RuleID              string        `json:"rule_id"`
	RuleVersion         string        `json:"rule_version"`
	Severity            string        `json:"severity"`
	Gate                string        `json:"gate"`
	Group               string        `json:"group"`
	Scope               string        `json:"scope"`
	Message             string        `json:"message"`
	Primary             Location      `json:"primary"`
	Related             []Location    `json:"related"`
	Evidence            rule.Evidence `json:"evidence"`
	Fingerprint         string        `json:"fingerprint"`
	Suppressed          bool          `json:"suppressed"`
	SuppressionIDs      []string      `json:"suppression_ids,omitempty"`
	BaselineState       string        `json:"baseline_state"`
	BaselineFingerprint string        `json:"baseline_fingerprint,omitempty"`
	Derived             bool          `json:"derived"`
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

// Assessment is a local, nondilutable index. A revision probability is present
// only when a compatible pack qualifies the unit; ProbabilityStatus always
// explains an unavailable estimate, and an index is never a probability.
type Assessment struct {
	ChangeState            string         `json:"change_state,omitempty"`
	ChangeFingerprint      string         `json:"change_fingerprint,omitempty"`
	BaselineFingerprint    string         `json:"baseline_fingerprint,omitempty"`
	BaselineState          string         `json:"baseline_state,omitempty"`
	Path                   string         `json:"path"`
	Scope                  string         `json:"scope"`
	UnitID                 int            `json:"unit_id"`
	Span                   document.Span  `json:"span"`
	Words                  int            `json:"words"`
	SlopScore              float64        `json:"slop_score"`
	EffectiveSlopScore     float64        `json:"effective_slop_score"`
	SlopProbability        *float64       `json:"slop_probability"`
	ProbabilityStatus      string         `json:"probability_status"`
	ProbabilityDetail      string         `json:"probability_detail,omitempty"`
	OriginEstimate         *float64       `json:"origin_estimate,omitempty"`
	OriginStatus           string         `json:"origin_status,omitempty"`
	OriginDetail           string         `json:"origin_detail,omitempty"`
	Status                 string         `json:"status"`
	Contributions          []Contribution `json:"contributions"`
	EffectiveContributions []Contribution `json:"effective_contributions,omitempty"`
}

// DocumentResult summarizes one source without including its prose by default.
type DocumentResult struct {
	GateMode         string               `json:"gate_mode,omitempty"`
	Name             string               `json:"name"`
	Format           document.Format      `json:"format"`
	SourceHash       string               `json:"source_hash"`
	ConfigHash       string               `json:"config_hash,omitempty"`
	AppliedOverrides []string             `json:"applied_overrides,omitempty"`
	Bytes            int                  `json:"bytes"`
	ProseWords       int                  `json:"prose_words"`
	Blocks           int                  `json:"blocks"`
	Sentences        int                  `json:"sentences"`
	Excluded         []document.Exclusion `json:"excluded"`
	Source           string               `json:"source,omitempty"`
	Maximum          float64              `json:"paragraph_maximum"`
	Median           float64              `json:"paragraph_median"`
	P90              float64              `json:"paragraph_p90"`
	FlaggedFraction  float64              `json:"eligible_paragraph_flagged_fraction"`
}

// Manifest records the exact execution policy and feature versions, without time.
type Manifest struct {
	Git             *GitSelection             `json:"git,omitempty"`
	GateMode        string                    `json:"gate_mode,omitempty"`
	ToolVersion     string                    `json:"tool_version"`
	ToolCommit      string                    `json:"tool_commit"`
	ConfigHash      string                    `json:"config_hash"`
	ConfigIdentity  string                    `json:"config_identity,omitempty"`
	ConfigSources   []config.SourceIdentity   `json:"config_sources,omitempty"`
	ConfigOverrides []config.OverrideIdentity `json:"config_overrides,omitempty"`
	RulesetHash     string                    `json:"ruleset_hash"`
	ScoringProfile  string                    `json:"scoring_profile"`
	FeatureContract string                    `json:"feature_contract"`
	Probability     *ProbabilityModel         `json:"probability,omitempty"`
	Origin          *ProbabilityModel         `json:"origin,omitempty"`
	NLP             nlp.Identity              `json:"nlp"`
	Rules           []rule.Descriptor         `json:"rules"`
	SelectionMode   string                    `json:"selection_mode"`
	Complete        bool                      `json:"complete"`
	NoGate          bool                      `json:"no_gate"`
	IncludeSource   bool                      `json:"include_source"`
	SkippedRules    []string                  `json:"skipped_rules"`
	// AbstainedRules names each rule that abstained on at least one document,
	// sorted and unique. Such a rule did not cover every document of the run;
	// RunResult.Abstentions lists the documents and reasons.
	AbstainedRules []string `json:"abstained_rules,omitempty"`
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
	Unchanged []GateReason `json:"unchanged,omitempty"`
	Passed    bool         `json:"passed"`
	Reasons   []GateReason `json:"reasons"`
	Accepted  []GateReason `json:"accepted,omitempty"`
}

// RunError records an operational failure in a partial result.
type RunError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// RuleAbstention records one rule that declined to judge one document for a
// declared reason, such as an exhausted candidate budget. The document stays
// measured by every other rule and the run stays complete; the abstaining rule
// contributes no findings and no activation values for that document.
type RuleAbstention struct {
	Path        string `json:"path"`
	RuleID      string `json:"rule_id"`
	RuleVersion string `json:"rule_version"`
	Reason      string `json:"reason"`
	Detail      string `json:"detail,omitempty"`
}

// SuppressionTarget identifies one complete unit permitted by a source directive.
type SuppressionTarget struct {
	Scope  string        `json:"scope"`
	UnitID int           `json:"unit_id"`
	Span   document.Span `json:"span"`
}

// Suppression records an explicit permission and links it to retained raw findings.
// A partially used or unused permission is an error when reject_unused is true.
type Suppression struct {
	TrustState       string              `json:"trust_state,omitempty"`
	TrustFingerprint string              `json:"trust_fingerprint,omitempty"`
	ID               string              `json:"id"`
	Kind             string              `json:"kind"`
	RuleIDs          []string            `json:"rule_ids"`
	Reason           string              `json:"reason"`
	Directive        Location            `json:"directive"`
	End              *Location           `json:"end,omitempty"`
	Targets          []SuppressionTarget `json:"targets"`
	FindingIDs       []string            `json:"finding_ids"`
	UsedRules        []string            `json:"used_rules"`
	Status           string              `json:"status"`
}

// RunResult is the immutable input shared by every reporter.
// Returned slices are owned by the caller and do not alias the engine.
type RunResult struct {
	PreparedFeatures *PreparedFeatureCollection `json:"prepared_features,omitempty"`
	Features         *FeatureCollection         `json:"features,omitempty"`
	PolicyComparison *PolicyComparison          `json:"policy_comparison,omitempty"`
	Changes          *ChangeSelection           `json:"changes,omitempty"`
	BaselineSnapshot *baseline.Snapshot         `json:"baseline_snapshot,omitempty"`
	Baseline         *baseline.Comparison       `json:"baseline,omitempty"`
	SchemaVersion    string                     `json:"schema_version"`
	Status           string                     `json:"status"`
	Manifest         Manifest                   `json:"manifest"`
	Documents        []DocumentResult           `json:"documents"`
	Findings         []Finding                  `json:"findings"`
	Assessments      []Assessment               `json:"assessments"`
	Gate             GateDecision               `json:"gate"`
	Errors           []RunError                 `json:"errors"`
	Suppressions     []Suppression              `json:"suppressions,omitempty"`
	// Abstentions lists each rule that declined one document, sorted by path
	// and rule ID. Manifest.AbstainedRules names the affected rules once.
	Abstentions []RuleAbstention `json:"abstentions,omitempty"`
}

// Result is the same report contract restricted to one source by Analyze.
type Result = RunResult
