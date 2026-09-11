package evaluation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/training"
)

// Trial identifies one frozen input and its training basis and response meaning.
type Trial struct {
	PredictionsSHA256 string                  `json:"predictions_sha256"`
	PlanSHA256        string                  `json:"plan_sha256"`
	Plan              training.PredictionPlan `json:"plan"`
	Identity          training.Identity       `json:"identity"`
	TrainingBasis     string                  `json:"training_basis"`
	TrainingConstant  float64                 `json:"training_constant"`
	Unavailable       map[string]int          `json:"unavailable_reasons"`
}

// ComparisonResult binds numerical comparisons to frozen inputs and labels.
// It does not select a winner, fit a model, or qualify a product probability.
type ComparisonResult struct {
	Version           string         `json:"version"`
	SHA256            string         `json:"sha256,omitempty"`
	Status            string         `json:"status"`
	HumanCorpus       string         `json:"human_corpus"`
	ProbabilityStatus string         `json:"probability_status"`
	Basis             string         `json:"basis"`
	Execution         Execution      `json:"execution"`
	Plan              ComparisonPlan `json:"plan"`
	PlanSHA256        string         `json:"plan_sha256"`
	Candidate         Trial          `json:"candidate"`
	Comparator        Trial          `json:"comparator"`
	DecisionsSHA256   string         `json:"decisions_sha256"`
	RoundSHA256       string         `json:"round_sha256"`
	Candidates        int            `json:"candidates"`
	Excluded          map[string]int `json:"excluded_labels"`
	Summary           PairedSummary  `json:"summary"`
}

// RunComparison loads two frozen prediction artifacts, validates their common
// scope and one independent annotation round, and summarizes paired outcomes.
// Sources and external model files are unnecessary; no inference runs here.
func RunComparison(ctx context.Context, data, comparator []byte, plan ComparisonPlan,
	candidates corpus.Artifact, round *annotation.Round, allowSimulation bool,
) (ComparisonResult, error) {
	a, err := training.LoadPredictions(ctx, data)
	if err != nil {
		return ComparisonResult{}, err
	}
	b, err := training.LoadPredictions(ctx, comparator)
	if err != nil {
		return ComparisonResult{}, err
	}
	planHash, err := comparisonPlanIdentity(ctx, plan, a, b)
	if err != nil {
		return ComparisonResult{}, err
	}
	decisions, err := comparisonDecisions(ctx, candidates, round, a, b, allowSimulation)
	if err != nil {
		return ComparisonResult{}, err
	}
	rows, excluded, err := comparisonRows(ctx, a, b, decisions)
	if err != nil {
		return ComparisonResult{}, err
	}
	left, right, err := comparisonTrials(a, b)
	if err != nil {
		return ComparisonResult{}, err
	}
	summary, err := Compare(ctx, rows, [2]float64{left.TrainingConstant, right.TrainingConstant})
	if err != nil {
		return ComparisonResult{}, err
	}
	result := ComparisonResult{Version: ComparisonVersion, Status: "experimental_metrics", HumanCorpus: "not_qualified",
		ProbabilityStatus: "unavailable_unqualified_model", Basis: decisions.Basis, Execution: executionIdentity(),
		Plan: plan, PlanSHA256: planHash,
		Candidate: left, Comparator: right, DecisionsSHA256: decisions.SHA256, RoundSHA256: decisions.RoundSHA256,
		Candidates: len(a.Rows), Excluded: excluded, Summary: summary}
	return finishComparison(ctx, result)
}

func comparisonDecisions(ctx context.Context, candidates corpus.Artifact, round *annotation.Round,
	a, b training.Predictions, simulation bool,
) (annotation.DecisionSet, error) {
	for _, predictions := range []training.Predictions{a, b} {
		if err := validateTargets(ctx, candidates, predictions, round); err != nil {
			return annotation.DecisionSet{}, err
		}
	}
	decisions, err := round.Decisions(ctx)
	if err != nil {
		return annotation.DecisionSet{}, err
	}
	if err := checkDecisions(decisions, a.Model, simulation); err != nil {
		return annotation.DecisionSet{}, err
	}
	return decisions, nil
}

func comparisonRows(ctx context.Context, a, b training.Predictions, decisions annotation.DecisionSet,
) ([]PairedObservation, map[string]int, error) {
	left, excluded, err := labeledRows(ctx, a, decisions)
	if err != nil {
		return nil, nil, err
	}
	right, _, err := labeledRows(ctx, b, decisions)
	if err != nil {
		return nil, nil, err
	}
	byID := make(map[string]Observation, len(right))
	for _, row := range right {
		byID[row.UnitID] = row
	}
	paired := make([]PairedObservation, 0, len(left))
	for _, row := range left {
		paired = append(paired, PairedObservation{row, byID[row.UnitID]})
	}
	return paired, excluded, nil
}

func comparisonTrials(a, b training.Predictions) (Trial, Trial, error) {
	left, err := trainingConstant(a.Model)
	if err != nil {
		return Trial{}, Trial{}, err
	}
	right, err := trainingConstant(b.Model)
	if err != nil {
		return Trial{}, Trial{}, err
	}
	return Trial{a.SHA256, a.PlanSHA256, a.Plan, a.Model.Identity, a.Model.Basis, left, unavailableReasons(a)},
		Trial{b.SHA256, b.PlanSHA256, b.Plan, b.Model.Identity, b.Model.Basis, right, unavailableReasons(b)}, nil
}

func unavailableReasons(predictions training.Predictions) map[string]int {
	counts := make(map[string]int)
	for _, row := range predictions.Rows {
		if row.Response == nil {
			counts[row.Reason]++
		}
	}
	return counts
}

func finishComparison(ctx context.Context, result ComparisonResult) (ComparisonResult, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return ComparisonResult{}, err
	}
	if len(data) > training.MaxPredictionBytes {
		return ComparisonResult{}, fmt.Errorf("comparison exceeds artifact byte limit")
	}
	result.SHA256 = fmt.Sprintf("%x", sha256.Sum256(data))
	if err := ctx.Err(); err != nil {
		return ComparisonResult{}, err
	}
	return result, nil
}
