package training

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/probability"
)

// PackOptions carry the decisions a training artifact cannot make. The minimum
// word count is an applicability limit chosen from validation data; acceptance
// is a maintainer statement about a qualified corpus and a published evaluation.
type PackOptions struct {
	ID         string
	MinWords   int
	Task       string
	Accepted   bool
	Evaluation string
}

// BuildPack converts one fitted artifact into a product pack. It copies the
// numerical parameters and the measurement contract the artifact already
// records, and refuses to invent anything else. A pack it produces is
// experimental unless the artifact declares a qualified corpus and the caller
// supplies the evaluation reference that acceptance requires.
func BuildPack(ctx context.Context, artifact Artifact, options PackOptions) (probability.File, error) {
	if err := ctx.Err(); err != nil {
		return probability.File{}, err
	}
	if err := validatePackInput(artifact, options); err != nil {
		return probability.File{}, err
	}
	identity := artifact.Identity
	file := probability.File{
		Version: probability.Version, ID: options.ID, DeclaredStatus: "experimental",
		HumanCorpus: artifact.HumanCorpus, Task: options.Task, Rubric: identity.Rubric, Kind: identity.Kind,
		Contract: probability.Contract{
			FeatureContract: identity.FeatureContract, UnitContract: identity.UnitContract,
			Columns: slices.Clone(identity.Columns), ColumnsSHA256: identity.ColumnsSHA256,
			NLP: identity.NLP, Capabilities: slices.Clone(identity.Capabilities),
			PreparationHash: identity.PreparationHash, IncludeQuotes: identity.IncludeQuotes,
			IncludeStructure: identity.IncludeStructure,
		},
		Limits:    probability.Limits{MinWords: options.MinWords},
		Estimator: "logistic",
		Logistic: &probability.Logistic{Means: slices.Clone(artifact.Logistic.Means),
			Scales: slices.Clone(artifact.Logistic.Scales), Weights: slices.Clone(artifact.Logistic.Weights),
			Intercept: artifact.Logistic.Intercept},
		// The pack records the mapping kind its contract defines. The fitting
		// implementation stays with the artifact that produced these knots.
		Calibration: probability.Calibration{Algorithm: "isotonic",
			Scores: slices.Clone(artifact.Calibration.Scores), Responses: slices.Clone(artifact.Calibration.Responses)},
	}
	if options.Accepted {
		file.DeclaredStatus, file.Evaluation = "accepted", options.Evaluation
	}
	digest, err := probability.Digest(file)
	if err != nil {
		return probability.File{}, err
	}
	file.SHA256 = digest
	if err := ctx.Err(); err != nil {
		return probability.File{}, err
	}
	return file, nil
}

func validatePackInput(artifact Artifact, options PackOptions) error {
	if !slices.Contains([]string{probability.Task, probability.TaskOrigin}, options.Task) {
		return fmt.Errorf("a pack estimates %s or %s", probability.Task, probability.TaskOrigin)
	}
	if artifact.Options.Estimator != "logistic" || artifact.Logistic == nil {
		return fmt.Errorf("only a logistic artifact becomes a pack; this one used %q", artifact.Options.Estimator)
	}
	if artifact.Calibration == nil || artifact.Calibration.Algorithm != model.IsotonicAlgorithm {
		return fmt.Errorf("a pack needs the artifact's separate isotonic calibration")
	}
	if options.MinWords < 1 {
		return fmt.Errorf("a pack needs a minimum word count chosen from validation data")
	}
	if !options.Accepted {
		return nil
	}
	if artifact.HumanCorpus != "qualified" {
		return fmt.Errorf("an accepted pack needs a qualified corpus; this artifact records %q", artifact.HumanCorpus)
	}
	if options.Evaluation == "" {
		return fmt.Errorf("an accepted pack must name its published evaluation")
	}
	return nil
}
