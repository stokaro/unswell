package main

import (
	"flag"

	"github.com/stokaro/unswell/research/annotation/training"
)

func trainingFlags(flags *flag.FlagSet, options *training.Options) {
	flags.StringVar(&options.Kind, "kind", "", "One corpus target kind: sentence, paragraph, or fragment")
	flags.StringVar(&options.MissingFeatures, "missing-features", "reject", "Missing feature policy: reject or exclude")
	flags.StringVar(&options.Calibration, "calibration", "none", "Separate calibration: none or isotonic")
	flags.BoolVar(&options.AllowSimulation, "allow-simulation", false, "Allow explicitly simulated tutorial fitting")
	flags.Float64Var(&options.Fit.L2, "l2", 1, "Logistic weight regularization; the intercept is unpenalized")
	flags.Float64Var(&options.Fit.Tolerance, "tolerance", 1e-8, "Maximum gradient infinity norm at convergence")
	flags.IntVar(&options.Fit.MaxIterations, "max-iterations", 100, "Maximum Newton iterations")
	flags.Int64Var(&options.Fit.MaxOperations, "max-operations", 1000000000, "Maximum logical fitting operations")
}
