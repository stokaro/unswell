package training

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// MaxPredictionBytes bounds a saved model plus its reserved-target predictions.
const MaxPredictionBytes = 32 << 20

// LoadPredictionPlan rejects ambiguous or unsupported run plans. Model and corpus
// bindings are checked again by Predict before measuring any inputs.
func LoadPredictionPlan(ctx context.Context, data []byte) (PredictionPlan, error) {
	var plan PredictionPlan
	if err := jsoninput.Decode(ctx, data, 16<<10, &plan, jsoninput.Limits{Array: 1, Object: 16}); err != nil {
		return PredictionPlan{}, err
	}
	if err := validatePlanFields(plan); err != nil {
		return PredictionPlan{}, err
	}
	return plan, nil
}

func validatePlanFields(p PredictionPlan) error {
	if p.Version != PredictionVersion || len(p.ID) == 0 || len(p.ID) > 128 || strings.TrimSpace(p.ID) != p.ID {
		return fmt.Errorf("prediction plan requires a supported version and a bounded ID")
	}
	if !planDigestsValid(p) {
		return fmt.Errorf("prediction plan requires protocol, model, and corpus SHA-256 identities")
	}
	if !slices.Contains([]string{"development", "final_test"}, p.Partition) ||
		!slices.Contains([]string{"prepared_piece", "source_document"}, p.Context) ||
		!slices.Contains([]string{"logistic", "forest", "isotonic"}, p.Response) {
		return fmt.Errorf("prediction plan has an unsupported partition, context, or response")
	}
	if !validResponse(p.Threshold) {
		return fmt.Errorf("prediction plan requires an explicit finite threshold within [0,1]")
	}
	return nil
}

func validatePredictionPlan(p PredictionPlan, a Artifact) error {
	if err := validatePlanFields(p); err != nil {
		return err
	}
	if p.ModelSHA256 != a.SHA256 || p.CorpusSHA256 != a.CorpusSHA256 {
		return fmt.Errorf("prediction plan does not match the frozen model and corpus")
	}
	context, err := predictionContext(a.Identity.FeatureSource)
	if err != nil {
		return err
	}
	if p.Response != "isotonic" && p.Response != a.Options.Estimator {
		return fmt.Errorf("prediction response does not match the fitted estimator")
	}
	if p.Context != context || (p.Response == "isotonic" && a.Calibration == nil) {
		return fmt.Errorf("prediction plan has incompatible context or unavailable calibration")
	}
	return nil
}

func predictionContext(source string) (string, error) {
	switch source {
	case "rule_activations":
		return "source_document", nil
	case "", "lexical_ngrams":
		return "prepared_piece", nil
	default:
		return "", fmt.Errorf("unsupported prediction feature source")
	}
}

// LoadPredictions restores a saved numerical run without rerunning models or
// extraction. It checks its model, plan, rows, and digest; it does not attest origin.
func LoadPredictions(ctx context.Context, data []byte) (Predictions, error) {
	var result Predictions
	limits := jsoninput.Limits{Array: corpus.MaxUnits, Object: 256,
		Arrays: map[string]int{"scores": model.MaxCalibrationSamples, "responses": model.MaxCalibrationSamples}}
	if err := jsoninput.Decode(ctx, data, MaxPredictionBytes, &result, limits); err != nil {
		return Predictions{}, err
	}
	want := result.SHA256
	result.SHA256 = ""
	rebuilt, err := finishPredictions(ctx, result)
	if err != nil {
		return Predictions{}, err
	}
	if rebuilt.SHA256 != want || rebuilt.Version != PredictionVersion || rebuilt.Status != "experimental_predictions" ||
		rebuilt.ProbabilityStatus != "unavailable_unqualified_model" {
		return Predictions{}, fmt.Errorf("prediction artifact digest, version, or status mismatch")
	}
	if err := validatePredictionInputs(ctx, rebuilt); err != nil {
		return Predictions{}, err
	}
	return rebuilt, nil
}

func validatePredictionInputs(ctx context.Context, result Predictions) error {
	data, err := json.Marshal(result.Model)
	if err != nil {
		return err
	}
	fitted, err := Load(ctx, data)
	if err != nil {
		return err
	}
	if err := validatePredictionPlan(result.Plan, fitted); err != nil {
		return err
	}
	hash, err := hashJSON(result.Plan)
	if err != nil {
		return err
	}
	if result.PlanSHA256 != hash || result.Verification.ArtifactSHA256 != result.Plan.CorpusSHA256 ||
		result.Verification.Status != "source_and_candidates_reproduced" || len(result.Rows) == 0 {
		return fmt.Errorf("prediction artifact has inconsistent plan, verification, or targets")
	}
	if err := validatePredictionRows(ctx, result.Rows, result.Plan); err != nil {
		return err
	}
	return validatePredictionEstimator(result.Rows, fitted.Options.Estimator)
}

func validatePredictionRows(ctx context.Context, rows []Prediction, plan PredictionPlan) error {
	for i, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if i > 0 && row.UnitID <= rows[i-1].UnitID {
			return fmt.Errorf("prediction targets must be sorted and unique")
		}
		if err := validatePredictionRow(row, plan); err != nil {
			return err
		}
	}
	return nil
}

func validatePredictionRow(row Prediction, plan PredictionPlan) error {
	if row.UnitID == "" || row.SourceID == "" || row.GroupID == "" {
		return fmt.Errorf("prediction requires target, source, and group IDs")
	}
	if row.Status == "available" {
		if !validAvailablePrediction(row) {
			return fmt.Errorf("available prediction requires valid numerical outputs")
		}
		if *row.Positive != (*row.Response >= *plan.Threshold) ||
			!consistentRawResponse(row, plan.Response) {
			return fmt.Errorf("prediction response or threshold decision is inconsistent")
		}
		return nil
	}
	return validateAbsentPrediction(row, plan)
}

func validateAbsentPrediction(row Prediction, plan PredictionPlan) error {
	if row.Status != "inapplicable" || row.Response != nil || row.Positive != nil {
		return fmt.Errorf("unavailable prediction requires null response and decision")
	}
	if validCalibrationAbsence(row, plan) {
		return nil
	}
	if row.LinearScore != nil || row.LogisticResponse != nil || row.ForestResponse != nil {
		return fmt.Errorf("unmeasured target cannot have numerical outputs")
	}
	if validTargetAbsence(row) {
		return nil
	}
	if validFeatureAbsence(row) {
		return nil
	}
	return fmt.Errorf("prediction has unsupported absence reason or binding")
}

func finitePointer(value *float64) bool {
	return value != nil && !math.IsNaN(*value) && !math.IsInf(*value, 0)
}

func validResponse(value *float64) bool { return finitePointer(value) && *value >= 0 && *value <= 1 }

func finishPredictions(ctx context.Context, result Predictions) (Predictions, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return Predictions{}, err
	}
	if len(encoded)+128 > MaxPredictionBytes {
		return Predictions{}, fmt.Errorf("predictions exceed artifact byte limit")
	}
	result.SHA256 = fmt.Sprintf("%x", sha256.Sum256(encoded))
	if err := ctx.Err(); err != nil {
		return Predictions{}, err
	}
	return result, nil
}

func hashJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func planDigestsValid(p PredictionPlan) bool {
	return validDigest(p.ProtocolSHA256) && validDigest(p.ModelSHA256) && validDigest(p.CorpusSHA256)
}

func validAvailablePrediction(row Prediction) bool {
	return validDigest(row.FeatureInputHash) && row.Reason == "" && validRawPrediction(row) &&
		validResponse(row.Response) && row.Positive != nil
}

func validCalibrationAbsence(row Prediction, plan PredictionPlan) bool {
	return row.Reason == "calibration/out_of_range" && plan.Response == "isotonic" &&
		validDigest(row.FeatureInputHash) && validRawPrediction(row)
}

func validTargetAbsence(row Prediction) bool {
	return row.Reason == "target/no_complete_block_match" && row.FeatureInputHash == ""
}

func validFeatureAbsence(row Prediction) bool {
	return strings.HasPrefix(row.Reason, "feature/") && len(row.Reason) > len("feature/") && validDigest(row.FeatureInputHash)
}
