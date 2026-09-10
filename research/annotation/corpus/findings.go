package corpus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
	"github.com/stokaro/unswell/rule"
)

// FindingsVersion identifies label-free rule findings measured over candidates.
const FindingsVersion = "unswell-corpus-findings-v1"

// MaxFindingRecords bounds the bound finding records one measurement retains.
const MaxFindingRecords = 200000

// FindingRecord is one rule finding whose primary span lies inside a candidate.
type FindingRecord struct {
	ID          string        `json:"id"`
	RuleID      string        `json:"rule_id"`
	RuleVersion string        `json:"rule_version"`
	Severity    string        `json:"severity"`
	Gate        string        `json:"gate"`
	Scope       string        `json:"scope"`
	Span        document.Span `json:"span"`
}

// UnitFindings lists the findings located within one candidate. Nested
// candidates share a finding; an analysis chooses the unit kind it reports.
// Unmeasured marks a unit of a document the policy run could not finish.
type UnitFindings struct {
	UnitID     string          `json:"unit_id"`
	SourceID   string          `json:"source_id"`
	GroupID    string          `json:"group_id"`
	Partition  string          `json:"partition"`
	Cohort     string          `json:"cohort,omitempty"`
	Kind       string          `json:"kind"`
	Role       string          `json:"role"`
	Words      int             `json:"words"`
	Unmeasured bool            `json:"unmeasured,omitempty"`
	Findings   []FindingRecord `json:"findings"`
}

// DocumentFindings counts every retained finding of one source, whether or not
// a candidate contains it. Derived and suppressed findings are counted apart
// and never enter a candidate. A document whose policy run failed keeps the
// status "failed" with the error and zero counts; it is a coverage gap, never
// a document without findings.
type DocumentFindings struct {
	SourceID   string         `json:"source_id"`
	Path       string         `json:"path"`
	GroupID    string         `json:"group_id"`
	Partition  string         `json:"partition"`
	Cohort     string         `json:"cohort,omitempty"`
	Role       string         `json:"role"`
	Status     string         `json:"status"`
	Error      string         `json:"error,omitempty"`
	ProseWords int            `json:"prose_words"`
	Blocks     int            `json:"blocks"`
	Findings   int            `json:"findings"`
	Unbound    int            `json:"unbound"`
	Derived    int            `json:"derived"`
	Suppressed int            `json:"suppressed"`
	ByRule     map[string]int `json:"by_rule"`
}

// PolicyIdentity pins the configuration every source was measured under.
type PolicyIdentity struct {
	ConfigHash     string            `json:"config_hash"`
	RulesetHash    string            `json:"ruleset_hash"`
	ScoringProfile string            `json:"scoring_profile"`
	Rules          []rule.Descriptor `json:"rules"`
}

// FindingsArtifact holds rule findings over a reproduced corpus without any
// label. A count is a rule outcome under the pinned policy: not a quality
// judgment, not a false-positive rate, and not recall. SHA256 covers compact
// Go JSON with that field omitted.
type FindingsArtifact struct {
	Version      string             `json:"version"`
	Status       string             `json:"status"`
	SHA256       string             `json:"sha256,omitempty"`
	HumanCorpus  string             `json:"human_corpus"`
	Verification Verification       `json:"verification"`
	Policy       PolicyIdentity     `json:"policy"`
	Failed       int                `json:"failed_documents"`
	Documents    []DocumentFindings `json:"documents"`
	Units        []UnitFindings     `json:"units"`
}

// LoadFindings decodes a strict finding artifact and checks its digest, so an
// analysis cannot read an edited artifact as a measurement.
func LoadFindings(ctx context.Context, data []byte) (FindingsArtifact, error) {
	var artifact FindingsArtifact
	limits := jsoninput.Limits{Array: MaxFindingRecords, Object: MaxUnits}
	if err := jsoninput.Decode(ctx, data, MaxArtifactBytes, &artifact, limits); err != nil {
		return FindingsArtifact{}, err
	}
	if artifact.Version != FindingsVersion || artifact.HumanCorpus != "not_qualified" {
		return FindingsArtifact{}, fmt.Errorf("unsupported finding artifact version or corpus claim")
	}
	expected := artifact.SHA256
	artifact.SHA256 = ""
	encoded, err := json.Marshal(artifact)
	if err != nil {
		return FindingsArtifact{}, err
	}
	if hashBytes(encoded) != expected {
		return FindingsArtifact{}, fmt.Errorf("finding artifact digest does not match its content")
	}
	artifact.SHA256 = expected
	return artifact, nil
}

// MeasureFindings reproduces the corpus, runs the supplied policy over every
// source with the public engine, and locates each finding inside the
// candidates whose source range contains its primary span. The policy's
// extraction must equal the frozen corpus policy. It accepts no labels and
// enables no rule; errors return no partial artifact.
func MeasureFindings(ctx context.Context, artifact Artifact, files map[string][]byte,
	configuration []byte,
) (FindingsArtifact, error) {
	engine, frozen, result, err := prepareFindings(ctx, artifact, files, configuration)
	if err != nil {
		return FindingsArtifact{}, err
	}
	var bySource map[string][]boundedUnit
	result.Units, bySource = candidateFindings(artifact)
	groups := sourceGroups(artifact.Plan)
	records := 0
	for _, source := range artifact.Plan.Manifest.Sources {
		if err := ctx.Err(); err != nil {
			return FindingsArtifact{}, err
		}
		if err := checkFrozenExtraction(engine, frozen, source); err != nil {
			return FindingsArtifact{}, err
		}
		doc, err := measurePolicySource(ctx, engine, &result, source, files[source.Path], groups[source.ID], bySource[source.ID])
		if err != nil {
			return FindingsArtifact{}, err
		}
		result.Documents = append(result.Documents, doc)
		records += doc.Findings - doc.Unbound
		if records > MaxFindingRecords {
			return FindingsArtifact{}, fmt.Errorf("finding records exceed %d", MaxFindingRecords)
		}
	}
	return finishFindings(ctx, result)
}

// prepareFindings builds the engine, reproduces the corpus, compiles the
// frozen extraction policy, and takes the policy identity from a run on an
// empty document, so an artifact whose every source fails still names the
// policy it ran under.
func prepareFindings(ctx context.Context, artifact Artifact, files map[string][]byte,
	configuration []byte,
) (*unswell.Engine, *config.Plan, FindingsArtifact, error) {
	if len(configuration) == 0 {
		return nil, nil, FindingsArtifact{}, fmt.Errorf("finding measurement requires an explicit policy")
	}
	engine, err := unswell.New(unswell.Options{Config: configuration, Jobs: 1, AllowEmpty: true})
	if err != nil {
		return nil, nil, FindingsArtifact{}, err
	}
	verification, err := verifyMeasurements(ctx, artifact, files)
	if err != nil {
		return nil, nil, FindingsArtifact{}, err
	}
	frozen, err := frozenPlan(engine, artifact.Plan)
	if err != nil {
		return nil, nil, FindingsArtifact{}, err
	}
	result := FindingsArtifact{Version: FindingsVersion, Status: "verified_targets_with_policy_findings",
		HumanCorpus: "not_qualified", Verification: verification, Documents: []DocumentFindings{}}
	probe, err := engine.AnalyzeAll(ctx, []document.Source{{Name: "identity.md", Format: document.Markdown}})
	if err != nil {
		return nil, nil, FindingsArtifact{}, err
	}
	if err := result.Policy.adopt(probe.Manifest, engine.Catalog()); err != nil {
		return nil, nil, FindingsArtifact{}, err
	}
	return engine, frozen, result, nil
}

// measurePolicySource runs the policy over one source. An operational failure of
// the run is recorded on the document, not hidden and not fatal, so the
// document stays in the artifact as a coverage gap.
func measurePolicySource(ctx context.Context, engine *unswell.Engine, result *FindingsArtifact, source Source, data []byte,
	group Group, candidates []boundedUnit,
) (DocumentFindings, error) {
	run, err := analyzeFindingSource(ctx, engine, source, data)
	if err != nil {
		if ctx.Err() != nil {
			return DocumentFindings{}, ctx.Err()
		}
		result.Failed++
		return failedDocument(source, group, err, candidates, result.Units), nil
	}
	if err := result.Policy.adopt(run.Manifest, engine.Catalog()); err != nil {
		return DocumentFindings{}, err
	}
	return bindFindings(run, source, group, candidates, result.Units), nil
}

func failedDocument(source Source, group Group, cause error, candidates []boundedUnit, units []UnitFindings) DocumentFindings {
	doc := DocumentFindings{SourceID: source.ID, Path: source.Path, GroupID: group.ID, Partition: group.Partition,
		Role: source.Role, Status: "failed", Error: cause.Error(), ByRule: map[string]int{}}
	if source.Snapshot != nil {
		doc.Cohort = source.Snapshot.Cohort
	}
	for _, candidate := range candidates {
		units[candidate.index].Unmeasured = true
	}
	return doc
}

// frozenPlan compiles the corpus extraction policy so a measurement can prove
// its own policy extracts the same prose.
func frozenPlan(engine *unswell.Engine, plan Plan) (*config.Plan, error) {
	configuration, err := measurementConfig(plan)
	if err != nil {
		return nil, err
	}
	expected, _, err := config.CompileBundle(config.Bundle{Root: ".unswell.yaml",
		Files: map[string][]byte{".unswell.yaml": configuration}}, engine.Catalog())
	return expected, err
}

// checkFrozenExtraction rejects a policy whose extraction differs from the
// frozen corpus policy for one source.
func checkFrozenExtraction(engine *unswell.Engine, expected *config.Plan, source Source) error {
	policy, err := engine.PolicyForFile(source.Path)
	if err != nil {
		return err
	}
	frozen, err := expected.ForFile(source.Path)
	if err != nil {
		return err
	}
	if policy.Analysis.IncludeQuotes || !sameRuleExtraction(policy.Extraction, frozen.Extraction) {
		return fmt.Errorf("rule extraction differs from frozen corpus policy for %s", source.ID)
	}
	return nil
}

func analyzeFindingSource(ctx context.Context, engine *unswell.Engine, source Source, data []byte) (unswell.RunResult, error) {
	run, err := engine.AnalyzeAll(ctx, []document.Source{{Name: source.Path, Format: source.Format, Bytes: data}})
	if err != nil {
		return unswell.RunResult{}, err
	}
	if !run.Manifest.Complete || len(run.Documents) != 1 || len(run.Errors) != 0 {
		message := "incomplete policy run"
		if len(run.Errors) > 0 {
			message = run.Errors[0].Message
		}
		return unswell.RunResult{}, fmt.Errorf("%s: %s", source.ID, message)
	}
	return run, nil
}

func (identity *PolicyIdentity) adopt(manifest unswell.Manifest, rules []rule.Descriptor) error {
	if identity.ConfigHash == "" {
		*identity = PolicyIdentity{ConfigHash: manifest.ConfigHash, RulesetHash: manifest.RulesetHash,
			ScoringProfile: manifest.ScoringProfile, Rules: rules}
		return nil
	}
	if identity.ConfigHash != manifest.ConfigHash || identity.RulesetHash != manifest.RulesetHash {
		return fmt.Errorf("policy identity changed between sources")
	}
	return nil
}

type boundedUnit struct {
	index  int
	bounds document.Span
}

// candidateFindings prepares one empty record per candidate and indexes the
// candidates of each source with the bounds of their original segments.
func candidateFindings(artifact Artifact) ([]UnitFindings, map[string][]boundedUnit) {
	units := make([]UnitFindings, 0, len(artifact.Units))
	bySource := make(map[string][]boundedUnit)
	for i, candidate := range artifact.Units {
		units = append(units, UnitFindings{UnitID: candidate.Unit.ID, SourceID: candidate.SourceID,
			GroupID: candidate.GroupID, Partition: candidate.Partition, Cohort: candidate.Cohort,
			Kind: candidate.Unit.Kind, Role: candidate.Unit.Role, Words: candidate.Words, Findings: []FindingRecord{}})
		bySource[candidate.SourceID] = append(bySource[candidate.SourceID],
			boundedUnit{index: i, bounds: document.Bounds(candidate.Unit.Source.Segments)})
	}
	return units, bySource
}

// bindFindings attaches each retained finding to every candidate of its source
// whose bounds contain the finding's primary span, and counts the rest.
func bindFindings(run unswell.RunResult, source Source, group Group, candidates []boundedUnit,
	units []UnitFindings,
) DocumentFindings {
	doc := DocumentFindings{SourceID: source.ID, Path: source.Path, GroupID: group.ID, Partition: group.Partition,
		Role: source.Role, Status: "measured", ProseWords: run.Documents[0].ProseWords, Blocks: run.Documents[0].Blocks,
		ByRule: map[string]int{}}
	if source.Snapshot != nil {
		doc.Cohort = source.Snapshot.Cohort
	}
	for _, finding := range run.Findings {
		switch {
		case finding.Derived:
			doc.Derived++
			continue
		case finding.Suppressed:
			doc.Suppressed++
			continue
		}
		doc.Findings++
		doc.ByRule[finding.RuleID]++
		record := FindingRecord{ID: finding.ID, RuleID: finding.RuleID, RuleVersion: finding.RuleVersion,
			Severity: finding.Severity, Gate: finding.Gate, Scope: finding.Scope, Span: finding.Primary.Span}
		bound := false
		for _, candidate := range candidates {
			if candidate.bounds.Start <= record.Span.Start && record.Span.End <= candidate.bounds.End {
				units[candidate.index].Findings = append(units[candidate.index].Findings, record)
				bound = true
			}
		}
		if !bound {
			doc.Unbound++
		}
	}
	return doc
}

func finishFindings(ctx context.Context, result FindingsArtifact) (FindingsArtifact, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return FindingsArtifact{}, err
	}
	if len(encoded)+128 > MaxArtifactBytes {
		return FindingsArtifact{}, fmt.Errorf("findings artifact exceeds %d bytes", MaxArtifactBytes)
	}
	result.SHA256 = hashBytes(encoded)
	if err := ctx.Err(); err != nil {
		return FindingsArtifact{}, err
	}
	return result, nil
}
