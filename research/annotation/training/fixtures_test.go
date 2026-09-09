package training_test

import "github.com/stokaro/unswell/research/annotation/training"

func fittingOptions() training.Options {
	return training.Options{Estimator: "logistic", Kind: "paragraph", Features: []string{"prose-words"},
		MissingFeatures: "reject", Calibration: "isotonic",
		AllowSimulation: true, Fit: &training.Fit{L2: 1, Tolerance: 1e-8, MaxIterations: 100, MaxOperations: 10_000_000}}
}
