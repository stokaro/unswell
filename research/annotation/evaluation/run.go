package evaluation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/training"
)

// Version identifies saved binary-event metric and exclusion semantics. Version
// 2 added the risk-coverage curve, version 3 recall at fixed false-positive
// limits and prevalence sensitivity, and version 4 the per-stratum breakdown.
const Version = "unswell-research-evaluation-v4"

// Result binds a numerical summary to exact saved predictions and an independent
// annotation round. No inference or training runs while computing this report.
type Result struct {
	Version           string                  `json:"version"`
	SHA256            string                  `json:"sha256,omitempty"`
	Status            string                  `json:"status"`
	HumanCorpus       string                  `json:"human_corpus"`
	ProbabilityStatus string                  `json:"probability_status"`
	Basis             string                  `json:"basis"`
	TrainingBasis     string                  `json:"training_basis"`
	PredictionsSHA256 string                  `json:"predictions_sha256"`
	PlanSHA256        string                  `json:"plan_sha256"`
	Plan              training.PredictionPlan `json:"plan"`
	Identity          training.Identity       `json:"identity"`
	DecisionsSHA256   string                  `json:"decisions_sha256"`
	RoundSHA256       string                  `json:"round_sha256"`
	Candidates        int                     `json:"candidates"`
	Excluded          map[string]int          `json:"excluded_labels"`
	TrainingConstant  float64                 `json:"training_constant"`
	Summary           Summary                 `json:"summary"`
	Strata            []Stratum               `json:"strata"`
}

// Run joins independent editorial decisions to saved predictions. Corpus and
// target mappings must agree exactly; missing or uncertain labels remain counted
// exclusions. Simulated rounds require explicit opt-in and remain unqualified.
func Run(ctx context.Context, data []byte, candidates corpus.Artifact,
	round *annotation.Round, allowSimulation bool,
) (Result, error) {
	predictions, err := training.LoadPredictions(ctx, data)
	if err != nil {
		return Result{}, err
	}
	if err := validateTargets(ctx, candidates, predictions, round); err != nil {
		return Result{}, err
	}
	decisions, err := evaluationDecisions(ctx, round, predictions.Model, allowSimulation)
	if err != nil {
		return Result{}, err
	}
	rows, excluded, err := labeledRows(ctx, predictions, decisions)
	if err != nil {
		return Result{}, err
	}
	constant, err := trainingConstant(predictions.Model)
	if err != nil {
		return Result{}, err
	}
	summary, err := Summarize(ctx, rows, constant)
	if err != nil {
		return Result{}, err
	}
	strata, err := Stratify(ctx, rows, candidateAttributes(candidates), constant)
	if err != nil {
		return Result{}, err
	}
	result := Result{Version: Version, Status: "experimental_metrics", HumanCorpus: "not_qualified",
		ProbabilityStatus: "unavailable_unqualified_model", Basis: decisions.Basis, TrainingBasis: predictions.Model.Basis,
		PredictionsSHA256: predictions.SHA256, PlanSHA256: predictions.PlanSHA256,
		Plan: predictions.Plan, Identity: predictions.Model.Identity,
		DecisionsSHA256: decisions.SHA256, RoundSHA256: decisions.RoundSHA256, Candidates: len(predictions.Rows),
		Excluded: excluded, TrainingConstant: constant, Summary: summary, Strata: strata}
	return finish(ctx, result)
}

// candidateAttributes reads the facts the corpus already records about each
// unit; nothing here is inferred from text.
func candidateAttributes(candidates corpus.Artifact) map[string]Attributes {
	attributes := make(map[string]Attributes, len(candidates.Units))
	for _, candidate := range candidates.Units {
		attributes[candidate.Unit.ID] = Attributes{Words: candidate.Words, Role: candidate.Unit.Role,
			Language: string(candidate.Unit.Source.Language), ProseLanguage: candidate.Unit.Source.ProseLanguage,
			Origin: candidate.Unit.Origin.Label}
	}
	return attributes
}

func finish(ctx context.Context, result Result) (Result, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return Result{}, err
	}
	if len(encoded) > training.MaxPredictionBytes {
		return Result{}, fmt.Errorf("evaluation exceeds artifact byte limit")
	}
	result.SHA256 = fmt.Sprintf("%x", sha256.Sum256(encoded))
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	return result, nil
}

func validateTargets(ctx context.Context, candidates corpus.Artifact, predictions training.Predictions, round *annotation.Round) error {
	data, err := json.Marshal(candidates)
	if err != nil {
		return err
	}
	if _, err := corpus.LoadArtifact(ctx, data); err != nil {
		return err
	}
	if candidates.SHA256 != predictions.Plan.CorpusSHA256 {
		return fmt.Errorf("evaluation corpus differs from the prediction corpus")
	}
	units := make([]annotation.Unit, 0, len(candidates.Units))
	targets := make(map[string]corpus.Candidate)
	for _, candidate := range candidates.Units {
		units = append(units, candidate.Unit)
		if candidate.Partition == predictions.Plan.Partition && candidate.Unit.Kind == predictions.Model.Options.Kind {
			targets[candidate.Unit.ID] = candidate
		}
	}
	if err := matchPredictionTargets(targets, predictions.Rows); err != nil {
		return err
	}
	return round.MatchTargets(ctx, units)
}

func matchPredictionTargets(targets map[string]corpus.Candidate, rows []training.Prediction) error {
	if len(targets) != len(rows) {
		return fmt.Errorf("prediction rows must cover every selected corpus target")
	}
	for _, row := range rows {
		target, exists := targets[row.UnitID]
		if !exists || row.SourceID != target.SourceID || row.GroupID != target.GroupID {
			return fmt.Errorf("prediction has an incompatible target, source, or group")
		}
	}
	return nil
}

func labeledRows(ctx context.Context, predictions training.Predictions, decisions annotation.DecisionSet,
) ([]Observation, map[string]int, error) {
	byID := make(map[string]annotation.EditorialDecision, len(decisions.Units))
	for _, decision := range decisions.Units {
		byID[decision.UnitID] = decision
	}
	rows, excluded := []Observation{}, make(map[string]int)
	for _, row := range predictions.Rows {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		decision, exists := byID[row.UnitID]
		if !exists {
			excluded["unannotated"]++
			continue
		}
		if decision.Status != "resolved" {
			excluded["decision/"+decision.Reason]++
			continue
		}
		if !slices.Contains(decision.Target.AllowedUses, "evaluation") {
			return nil, nil, fmt.Errorf("unit %s lacks a declared evaluation permission", row.UnitID)
		}
		if decision.Label == nil || !slices.Contains([]string{"acceptable", "needs_revision"}, *decision.Label) {
			return nil, nil, fmt.Errorf("evaluation requires a resolved binary editorial label")
		}
		label := 0
		if *decision.Label == "needs_revision" {
			label = 1
		}
		rows = append(rows, Observation{row.UnitID, row.GroupID, label, row.Response, row.Positive})
	}
	return rows, excluded, nil
}

func trainingConstant(fitted training.Artifact) (float64, error) {
	for _, partition := range fitted.Partitions {
		if partition.Name != "training" {
			continue
		}
		positive, negative := partition.Classes["needs_revision"], partition.Classes["acceptable"]
		if positive < 0 || negative < 0 || positive+negative == 0 || positive+negative != len(partition.Rows) {
			return 0, fmt.Errorf("training artifact requires consistent fitted class counts")
		}
		return float64(positive) / float64(positive+negative), nil
	}
	return 0, fmt.Errorf("training artifact has no training partition")
}

func evaluationDecisions(ctx context.Context, round *annotation.Round, fitted training.Artifact,
	allowSimulation bool,
) (annotation.DecisionSet, error) {
	decisions, err := round.Decisions(ctx)
	if err != nil {
		return annotation.DecisionSet{}, err
	}
	if decisions.Rubric != fitted.Identity.Rubric || decisions.ProfileSHA256 != fitted.Identity.ProfileSHA256 {
		return annotation.DecisionSet{}, fmt.Errorf("evaluation rubric or profile differs from the frozen model")
	}
	if decisions.Basis == "simulation" && !allowSimulation {
		return annotation.DecisionSet{}, fmt.Errorf("tutorial evaluation requires explicit allow_simulation")
	}
	return decisions, nil
}
