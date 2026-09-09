package annotation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

// MaxDecisionBytes bounds compact serialized decision exports.
const MaxDecisionBytes = 64 << 20

// DecisionSet binds editorial selections to the exact validated round input.
// A selection does not establish human participation or qualify a training corpus.
type DecisionSet struct {
	Version       string              `json:"version"`
	SHA256        string              `json:"sha256,omitempty"`
	RoundID       string              `json:"round_id"`
	RoundSHA256   string              `json:"round_sha256"`
	PacketSHA256  string              `json:"packet_sha256"`
	Purpose       string              `json:"purpose"`
	Basis         string              `json:"basis"`
	Rubric        string              `json:"rubric"`
	ProfileSHA256 string              `json:"profile_sha256"`
	HumanCorpus   string              `json:"human_corpus"`
	PrimaryRaters []string            `json:"primary_raters"`
	Units         []EditorialDecision `json:"units"`
}

// DecisionTarget binds a label to source ranges and hashes without copying prose.
// AllowedUses records the round's assertions, not independently verified rights.
type DecisionTarget struct {
	Kind          string          `json:"kind"`
	Role          string          `json:"role"`
	TextSHA256    string          `json:"text_sha256"`
	ContextSHA256 string          `json:"context_sha256"`
	SourceSHA256  string          `json:"source_sha256"`
	SourceBytes   int             `json:"source_bytes"`
	SourceFormat  document.Format `json:"source_format"`
	ProseLanguage string          `json:"prose_language"`
	Segments      []document.Span `json:"segments"`
	Extraction    Extraction      `json:"extraction"`
	AllowedUses   []string        `json:"allowed_uses"`
}

// EditorialDecision retains missing answers and uncertainty separately from labels.
// Original judgments and adjudication remain in the round bound by RoundSHA256.
type EditorialDecision struct {
	UnitID             string         `json:"unit_id"`
	Target             DecisionTarget `json:"target"`
	Status             string         `json:"status"`
	Reason             string         `json:"reason"`
	Basis              string         `json:"basis"`
	Label              *string        `json:"label"`
	Categories         []string       `json:"categories"`
	PrimaryJudgments   int            `json:"primary_judgments"`
	AuxiliaryJudgments int            `json:"auxiliary_judgments"`
	MissingRaters      []string       `json:"missing_raters"`
}

// Decisions selects labels only after every assigned primary rater has answered.
// Adjudication takes precedence; otherwise labels and category sets must agree.
// Returned data is detached. Errors and cancellation return no partial export.
func (r *Round) Decisions(ctx context.Context) (DecisionSet, error) {
	if r == nil || r.data.packetSHA256 == "" || r.inputSHA256 == "" {
		return DecisionSet{}, fmt.Errorf("load a validated annotation round before exporting decisions")
	}
	responses, err := r.data.responses(ctx)
	if err != nil {
		return DecisionSet{}, err
	}
	basis := "declared_human"
	if r.data.Purpose == "tutorial" {
		basis = "simulation"
	}
	result := DecisionSet{Version: "unswell-editorial-decisions-v1", RoundID: r.data.ID, RoundSHA256: r.inputSHA256,
		PacketSHA256: r.data.packetSHA256, Purpose: r.data.Purpose, Basis: basis, Rubric: r.data.Rubric,
		ProfileSHA256: r.data.Profile.SHA256, HumanCorpus: "not_qualified", PrimaryRaters: sortedKeys(responses.raters),
		Units: make([]EditorialDecision, 0, len(r.data.Units))}
	adjudications := make(map[string]Adjudication, len(r.data.Adjudications))
	for _, decision := range r.data.Adjudications {
		adjudications[decision.UnitID] = decision
	}
	for _, unit := range r.data.Units {
		if err := ctx.Err(); err != nil {
			return DecisionSet{}, err
		}
		decision, exists := adjudications[unit.ID]
		selected := editorialDecision(unit, responses.byUnit[unit.ID], result.PrimaryRaters, decision, exists)
		selected.AuxiliaryJudgments = responses.auxiliary[unit.ID]
		result.Units = append(result.Units, selected)
	}
	slices.SortFunc(result.Units, func(a, b EditorialDecision) int { return strings.Compare(a.UnitID, b.UnitID) })
	return finishDecisions(ctx, result)
}

func finishDecisions(ctx context.Context, result DecisionSet) (DecisionSet, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return DecisionSet{}, err
	}
	if len(data)+128 > MaxDecisionBytes {
		return DecisionSet{}, fmt.Errorf("decision export exceeds %d bytes", MaxDecisionBytes)
	}
	result.SHA256 = fmt.Sprintf("%x", sha256.Sum256(data))
	if err := ctx.Err(); err != nil {
		return DecisionSet{}, err
	}
	return result, nil
}
