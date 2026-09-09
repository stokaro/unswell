package training

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// Predict restores a frozen model and measures reserved corpus targets without
// reading an annotation round or fitting any parameters. Configuration is required
// only for the rule baseline and must match its recorded bytes and effective policy.
// Errors, including cancellation and corrupt inputs, return no partial artifact.
func Predict(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	fitted Artifact, plan PredictionPlan, configuration []byte,
) (Predictions, error) {
	fitted, plan, err := restorePredictionInputs(ctx, candidates, fitted, plan, configuration)
	if err != nil {
		return Predictions{}, err
	}
	selector, bindings, verification, err := predictionMeasurements(ctx, candidates, files, fitted, configuration)
	if err != nil {
		return Predictions{}, err
	}
	classifier, calibration, err := restoreModels(fitted)
	if err != nil {
		return Predictions{}, err
	}
	predictor := targetPredictor{fitted: fitted, plan: plan, classifier: classifier, calibration: calibration,
		measure: selector.measure, targets: make(map[string]corpus.Candidate, len(candidates.Units))}
	for _, candidate := range candidates.Units {
		predictor.targets[candidate.Unit.ID] = candidate
	}
	rows, err := predictor.predict(ctx, bindings)
	if err != nil {
		return Predictions{}, err
	}
	if len(rows) == 0 {
		return Predictions{}, fmt.Errorf("prediction partition has no targets of the fitted kind")
	}
	planHash, err := hashJSON(plan)
	if err != nil {
		return Predictions{}, err
	}
	result := Predictions{Version: PredictionVersion, Status: "experimental_predictions",
		ProbabilityStatus: "unavailable_unqualified_model", Plan: plan, PlanSHA256: planHash,
		Model: fitted, Verification: verification, Rows: rows}
	return finishPredictions(ctx, result)
}

func predictionMeasurements(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	fitted Artifact, configuration []byte,
) (rowSelector, []corpus.FeatureBinding, corpus.Verification, error) {
	// Only frozen rubric/profile identity enters this adapter, with no decisions.
	identity := annotation.DecisionSet{Rubric: fitted.Identity.Rubric, ProfileSHA256: fitted.Identity.ProfileSHA256}
	if fitted.Identity.FeatureSource == "rule_activations" {
		measured, err := corpus.MeasureRules(ctx, candidates, files, fitted.Options.Features, configuration)
		if err != nil {
			return rowSelector{}, nil, corpus.Verification{}, err
		}
		joined := corpus.RuleJoinedArtifact{Features: measured.Features, Bindings: measured.Bindings,
			Rules: measured.Rules, Context: "source_document", Decisions: identity}
		selector, bindings, err := ruleSelector(joined, fitted.Options, fmt.Sprintf("%x", sha256.Sum256(configuration)))
		return selector, bindings, measured.Verification, err
	}
	if len(configuration) != 0 {
		return rowSelector{}, nil, corpus.Verification{}, fmt.Errorf("prepared prediction does not accept rule configuration")
	}
	measured, err := corpus.Measure(ctx, candidates, files, fitted.Options.Features)
	if err != nil {
		return rowSelector{}, nil, corpus.Verification{}, err
	}
	joined := corpus.JoinedArtifact{Features: measured.Features, Bindings: measured.Bindings, Decisions: identity}
	selector, err := preparedSelector(joined, fitted.Options)
	return selector, measured.Bindings, measured.Verification, err
}

type targetPredictor struct {
	fitted      Artifact
	plan        PredictionPlan
	classifier  *model.Logistic
	calibration *model.Isotonic
	measure     func(corpus.FeatureBinding) (measurement, bool)
	targets     map[string]corpus.Candidate
}

func (p targetPredictor) predict(ctx context.Context, bindings []corpus.FeatureBinding) ([]Prediction, error) {
	rows := []Prediction{}
	slices.SortFunc(bindings, func(a, b corpus.FeatureBinding) int { return strings.Compare(a.UnitID, b.UnitID) })
	for _, binding := range bindings {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		candidate, exists := p.targets[binding.UnitID]
		if !exists {
			return nil, fmt.Errorf("measurement has no corpus target %s", binding.UnitID)
		}
		if binding.Partition != p.plan.Partition || candidate.Unit.Kind != p.fitted.Options.Kind {
			continue
		}
		if !slices.Contains(candidate.Unit.Rights.AllowedUses, "evaluation") {
			return nil, fmt.Errorf("unit %s lacks a declared evaluation permission", binding.UnitID)
		}
		row, err := p.one(ctx, binding)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (p targetPredictor) one(ctx context.Context, binding corpus.FeatureBinding) (Prediction, error) {
	row := Prediction{UnitID: binding.UnitID, SourceID: binding.SourceID, GroupID: binding.GroupID,
		FeatureInputHash: binding.FeatureInputHash, Status: "inapplicable"}
	unit, exists := p.measure(binding)
	if !exists {
		return Prediction{}, fmt.Errorf("missing measurement for %s", binding.UnitID)
	}
	if unit.reason != "" {
		row.Reason = "target/" + unit.reason
		return row, nil
	}
	if !reflect.DeepEqual(unit.identity, p.fitted.Identity) {
		return Prediction{}, fmt.Errorf("prediction has incompatible feature, policy, or NLP identity")
	}
	values, reason, err := numericValues(unit.values, p.fitted.Identity.Columns)
	if err != nil {
		return Prediction{}, err
	}
	if reason != "" {
		row.Reason = "feature/" + reason
		return row, nil
	}
	value, err := p.classifier.Evaluate(ctx, values)
	if err != nil {
		return Prediction{}, err
	}
	row.LinearScore, row.LogisticResponse = &value.LinearScore, &value.Response
	response := value.Response
	if p.plan.Response == "isotonic" {
		response, err = p.calibration.Evaluate(ctx, value.LinearScore)
		if errors.Is(err, model.ErrCalibrationRange) {
			row.Reason = "calibration/out_of_range"
			return row, nil
		}
		if err != nil {
			return Prediction{}, err
		}
	}
	positive := response >= *p.plan.Threshold
	row.Status, row.Response, row.Positive = "available", &response, &positive
	return row, nil
}

func restorePredictionInputs(ctx context.Context, candidates corpus.Artifact, fitted Artifact,
	plan PredictionPlan, configuration []byte,
) (Artifact, PredictionPlan, error) {
	encoded, err := json.Marshal(fitted)
	if err != nil {
		return Artifact{}, PredictionPlan{}, err
	}
	fitted, err = Load(ctx, encoded)
	if err != nil {
		return Artifact{}, PredictionPlan{}, err
	}
	planData, err := json.Marshal(plan)
	if err != nil {
		return Artifact{}, PredictionPlan{}, err
	}
	plan, err = LoadPredictionPlan(ctx, planData)
	if err != nil {
		return Artifact{}, PredictionPlan{}, err
	}
	if err := validatePredictionPlan(plan, fitted); err != nil {
		return Artifact{}, PredictionPlan{}, err
	}
	if !samePredictionCorpus(candidates, fitted, plan) {
		return Artifact{}, PredictionPlan{}, fmt.Errorf("prediction requires the frozen training corpus and split manifest")
	}
	if fitted.Identity.FeatureSource == "rule_activations" &&
		fmt.Sprintf("%x", sha256.Sum256(configuration)) != fitted.Identity.RuleConfigSHA256 {
		return Artifact{}, PredictionPlan{}, fmt.Errorf("prediction rule configuration digest mismatch")
	}
	if err := validateFittedTargets(fitted, candidates); err != nil {
		return Artifact{}, PredictionPlan{}, err
	}
	return fitted, plan, nil
}

func samePredictionCorpus(candidates corpus.Artifact, fitted Artifact, plan PredictionPlan) bool {
	return candidates.SHA256 == plan.CorpusSHA256 && candidates.Plan.ManifestSHA256 == fitted.ManifestSHA256
}
