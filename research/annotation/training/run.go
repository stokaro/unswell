package training

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// Run reproduces source targets, selects permitted resolved rows, and fits the
// shared Go model. Training and calibration use separate frozen partitions.
// It returns no partial artifact on error and never qualifies probabilities.
func Run(ctx context.Context, candidates corpus.Artifact, round *annotation.Round, files map[string][]byte,
	options Options,
) (Artifact, error) {
	if err := validateOptions(ctx, candidates, options); err != nil {
		return Artifact{}, err
	}
	joined, err := corpus.Join(ctx, candidates, round, files, options.Features)
	if err != nil {
		return Artifact{}, err
	}
	if joined.Decisions.Basis == "simulation" && !options.AllowSimulation {
		return Artifact{}, fmt.Errorf("tutorial training requires explicit allow_simulation")
	}
	options.Features = slices.Clone(joined.Features.Requested)
	selected, err := selectRows(ctx, candidates.Plan, joined, options)
	if err != nil {
		return Artifact{}, err
	}
	fit, err := model.FitLogistic(ctx, selected.training, options.Fit.numerical())
	if err != nil {
		return Artifact{}, fmt.Errorf("training partition: %w", err)
	}
	calibration, err := fitCalibration(ctx, fit.Model, selected.calibration, options.Calibration)
	if err != nil {
		return Artifact{}, err
	}
	result := Artifact{Version: Version, Status: "experimental_numerical_fit", HumanCorpus: "not_qualified",
		ProbabilityStatus: "unavailable_unqualified_model", Basis: joined.Decisions.Basis,
		CorpusSHA256: candidates.SHA256, ManifestSHA256: candidates.Plan.ManifestSHA256,
		RoundSHA256: joined.Decisions.RoundSHA256, JoinedSHA256: joined.SHA256, Options: options,
		Identity: selected.identity, Partitions: selected.partitions, Logistic: logisticResult(fit), Calibration: calibration}
	return finish(ctx, result)
}

func validateOptions(ctx context.Context, candidates corpus.Artifact, options Options) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !slices.Contains([]string{"sentence", "paragraph", "fragment"}, options.Kind) ||
		!slices.Contains(candidates.Plan.Manifest.UnitKinds, options.Kind) {
		return fmt.Errorf("training kind must select one kind present in the frozen corpus plan")
	}
	if options.MissingFeatures != "reject" && options.MissingFeatures != "exclude" {
		return fmt.Errorf("missing_features must be reject or exclude")
	}
	if options.Calibration != "none" && options.Calibration != "isotonic" {
		return fmt.Errorf("calibration must be none or isotonic")
	}
	return options.Fit.numerical().Validate()
}

func logisticResult(fit model.FitResult) Logistic {
	p := fit.Model.Parameters()
	return Logistic{Algorithm: fit.Algorithm, InputSHA256: fit.InputSHA256, Means: p.Means, Scales: p.Scales,
		Weights: p.Weights, Intercept: p.Intercept, Iterations: fit.Iterations, Operations: fit.Operations,
		Loss: fit.Loss, GradientNorm: fit.GradientNorm}
}

func fitCalibration(ctx context.Context, classifier *model.Logistic, examples []model.Example, method string) (*Calibration, error) {
	if method == "none" {
		return nil, nil
	}
	samples := make([]model.CalibrationSample, 0, len(examples))
	for _, example := range examples {
		value, err := classifier.Evaluate(ctx, example.Values)
		if err != nil {
			return nil, err
		}
		samples = append(samples, model.CalibrationSample{Score: value.LinearScore, Label: example.Label})
	}
	fit, err := model.FitIsotonic(ctx, samples)
	if err != nil {
		return nil, fmt.Errorf("calibration partition: %w", err)
	}
	p := fit.Model.Parameters()
	return &Calibration{Algorithm: fit.Algorithm, InputSHA256: fit.InputSHA256, ScoreKind: "linear_score",
		Scores: p.Scores, Responses: p.Responses, Samples: fit.Samples, DistinctScores: fit.DistinctScores,
		Pools: fit.Pools, MeanSquaredError: fit.MeanSquaredError}, nil
}

func finish(ctx context.Context, result Artifact) (Artifact, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return Artifact{}, err
	}
	if len(encoded)+128 > MaxArtifactBytes {
		return Artifact{}, fmt.Errorf("training artifact exceeds %d bytes", MaxArtifactBytes)
	}
	result.SHA256 = fmt.Sprintf("%x", sha256.Sum256(encoded))
	if err := ctx.Err(); err != nil {
		return Artifact{}, err
	}
	return result, nil
}
