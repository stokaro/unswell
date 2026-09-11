package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation/training"
)

func trainingFlags(flags *flag.FlagSet, options *training.Options) {
	options.Fit = &training.Fit{}
	options.Forest = &model.ForestOptions{}
	flags.StringVar(&options.Estimator, "estimator", "logistic", "Numerical estimator: logistic or forest")
	forestFlags(flags, options.Forest)
	flags.StringVar(&options.Kind, "kind", "", "One corpus target kind: sentence, paragraph, or fragment")
	flags.StringVar(&options.MissingFeatures, "missing-features", "reject",
		"Missing feature policy: reject, exclude, or zero (rule activations only: a rule that cannot fire counts as zero)")
	flags.StringVar(&options.Calibration, "calibration", "none", "Separate calibration: none or isotonic")
	flags.BoolVar(&options.AllowSimulation, "allow-simulation", false, "Allow explicitly simulated tutorial fitting")
	flags.Float64Var(&options.Fit.L2, "l2", 1, "Logistic weight regularization; the intercept is unpenalized")
	flags.Float64Var(&options.Fit.Tolerance, "tolerance", 1e-8, "Maximum gradient infinity norm at convergence")
	flags.IntVar(&options.Fit.MaxIterations, "max-iterations", 100, "Maximum Newton iterations")
	flags.Int64Var(&options.Fit.MaxOperations, "max-operations", 1000000000, "Maximum logical fitting operations")
}

func forestFlags(flags *flag.FlagSet, o *model.ForestOptions) {
	flags.Uint64Var(&o.Seed, "forest-seed", 1, "Seed for the private PCG generator")
	flags.IntVar(&o.Trees, "forest-trees", 32, "Number of trees")
	flags.IntVar(&o.MaxDepth, "forest-max-depth", 6, "Maximum split levels per tree")
	flags.IntVar(&o.MinLeaf, "forest-min-leaf", 2, "Minimum bootstrap samples per child")
	flags.IntVar(&o.FeaturesPerSplit, "forest-features-per-split", 0, "Candidate features; zero uses floor(sqrt(width))")
	flags.IntVar(&o.MaxNodes, "forest-max-nodes", 8191, "Total node budget; exhaustion returns an error")
	flags.BoolVar(&o.Bootstrap, "forest-bootstrap", true, "Sample training rows with replacement per tree")
}

func selectTrainingEstimator(flags *flag.FlagSet, o *training.Options) error {
	var incompatible string
	flags.Visit(func(f *flag.Flag) {
		forestFlag := strings.HasPrefix(f.Name, "forest-")
		logisticFlag := f.Name == "l2" || f.Name == "tolerance" || f.Name == "max-iterations"
		if (forestFlag && o.Estimator != "forest") || (logisticFlag && o.Estimator == "forest") {
			incompatible = f.Name
		}
	})
	if incompatible != "" {
		return fmt.Errorf("--%s is incompatible with estimator %s", incompatible, o.Estimator)
	}
	switch o.Estimator {
	case "logistic":
		o.Forest = nil
	case "forest":
		o.Forest.MaxOperations = o.Fit.MaxOperations
		o.Fit = nil
	default:
		return fmt.Errorf("estimator must be logistic or forest")
	}
	return nil
}
