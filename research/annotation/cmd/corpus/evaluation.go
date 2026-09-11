package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/evaluation"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/training"
)

type evaluationOptions struct {
	root, model, plan, protocol, rules, corpus, round string
	labels, generation                                string
	allowSimulation                                   bool
}

func runEvaluation(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	options, err := evaluationFlags(args)
	if err != nil {
		return err
	}
	maximum := corpus.MaxArtifactBytes
	if args[0] == "evaluate" {
		maximum = training.MaxPredictionBytes
	}
	data, err := commandio.Await(ctx, func() ([]byte, error) { return io.ReadAll(io.LimitReader(input, int64(maximum)+1)) })
	if err != nil {
		return err
	}
	var result any
	if args[0] == "predict" {
		result, err = predictOperation(ctx, options, data)
	} else {
		result, err = evaluateOperation(ctx, options, data)
	}
	if err != nil {
		return err
	}
	return writeResult(ctx, args[0], result, output)
}

func evaluationFlags(args []string) (evaluationOptions, error) {
	var options evaluationOptions
	flags := flag.NewFlagSet("corpus "+args[0], flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	if args[0] == "predict" {
		flags.StringVar(&options.root, "root", "", "Pinned local source root")
		flags.StringVar(&options.model, "model", "", "Frozen numerical training artifact")
		flags.StringVar(&options.plan, "plan", "", "Frozen prediction plan")
		flags.StringVar(&options.protocol, "protocol", "", "Protocol bytes matching the plan digest")
		flags.StringVar(&options.rules, "rule-config", "", "Exact inline configuration for the rule baseline")
	} else {
		flags.StringVar(&options.corpus, "corpus", "", "Frozen candidate artifact used by prediction")
		flags.StringVar(&options.round, "round", "", "Independent evaluation annotation round")
		flags.StringVar(&options.labels, "labels", "", "Label source instead of a round: provenance or cohort, from the corpus")
		flags.StringVar(&options.generation, "generation", "", "Generation records whose arms add operation, prompt, and family strata")
		flags.BoolVar(&options.allowSimulation, "allow-simulation", false, "Allow explicitly simulated tutorial labels")
	}
	if err := flags.Parse(args[1:]); err != nil {
		return evaluationOptions{}, err
	}
	if flags.NArg() != 0 {
		return evaluationOptions{}, fmt.Errorf("prediction and evaluation do not accept positional arguments")
	}
	return options, options.validate(args[0])
}

func (options evaluationOptions) validate(name string) error {
	if name == "predict" && (options.root == "" || options.model == "" || options.plan == "" || options.protocol == "") {
		return fmt.Errorf("predict requires --root, --model, --plan, and --protocol")
	}
	if name == "evaluate" {
		return options.validateLabelSource()
	}
	return nil
}

// validateLabelSource requires the frozen corpus and exactly one of a round
// and a label source.
func (options evaluationOptions) validateLabelSource() error {
	if options.corpus == "" || (options.round == "") == (options.labels == "") {
		return fmt.Errorf("evaluate requires --corpus and either --round or --labels")
	}
	return validateLabels(options.labels)
}

func predictOperation(ctx context.Context, options evaluationOptions, data []byte) (training.Predictions, error) {
	candidates, err := corpus.LoadArtifact(ctx, data)
	if err != nil {
		return training.Predictions{}, err
	}
	planData, err := loadLocalArtifact(ctx, options.plan, 16<<10, "prediction plan")
	if err != nil {
		return training.Predictions{}, err
	}
	plan, err := training.LoadPredictionPlan(ctx, planData)
	if err != nil {
		return training.Predictions{}, err
	}
	fitted, configuration, err := predictionResources(ctx, options, plan)
	if err != nil {
		return training.Predictions{}, err
	}
	files, err := loadFiles(ctx, options.root, candidates.Plan)
	if err != nil {
		return training.Predictions{}, err
	}
	return training.Predict(ctx, candidates, files, fitted, plan, configuration)
}

func predictionResources(ctx context.Context, options evaluationOptions, plan training.PredictionPlan) (training.Artifact, []byte, error) {
	protocol, err := loadLocalArtifact(ctx, options.protocol, 1<<20, "research protocol")
	if err != nil {
		return training.Artifact{}, nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(protocol)) != plan.ProtocolSHA256 {
		return training.Artifact{}, nil, fmt.Errorf("research protocol digest mismatch")
	}
	data, err := loadLocalArtifact(ctx, options.model, training.MaxArtifactBytes, "training artifact")
	if err != nil {
		return training.Artifact{}, nil, err
	}
	fitted, err := training.Load(ctx, data)
	if err != nil {
		return training.Artifact{}, nil, err
	}
	var configuration []byte
	if options.rules != "" {
		configuration, err = loadLocalArtifact(ctx, options.rules, maxRuleConfigBytes, "rule config")
	}
	return fitted, configuration, err
}

func evaluateOperation(ctx context.Context, options evaluationOptions, data []byte) (evaluation.Result, error) {
	encoded, err := loadLocalArtifact(ctx, options.corpus, corpus.MaxArtifactBytes, "candidate artifact")
	if err != nil {
		return evaluation.Result{}, err
	}
	candidates, err := corpus.LoadArtifact(ctx, encoded)
	if err != nil {
		return evaluation.Result{}, err
	}
	settings, err := evaluationSettings(ctx, options, candidates)
	if err != nil {
		return evaluation.Result{}, err
	}
	if options.labels != "" {
		decisions, err := labelDecisions(ctx, options.labels, candidates)
		if err != nil {
			return evaluation.Result{}, err
		}
		return evaluation.EvaluateDecisions(ctx, data, candidates, decisions, settings)
	}
	round, err := loadRound(ctx, options.round)
	if err != nil {
		return evaluation.Result{}, err
	}
	return evaluation.Evaluate(ctx, data, candidates, round, settings)
}

// evaluationSettings reads the optional generation records and maps their
// arms onto the corpus, so the strata can name each controlled source's arm.
func evaluationSettings(ctx context.Context, options evaluationOptions, candidates corpus.Artifact) (evaluation.Options, error) {
	settings := evaluation.Options{AllowSimulation: options.allowSimulation}
	if options.generation == "" {
		return settings, nil
	}
	records, err := loadGeneration(ctx, options.generation)
	if err != nil {
		return evaluation.Options{}, err
	}
	settings.Arms, err = evaluation.ArmsFromRecords(candidates, records)
	return settings, err
}

func loadLocalArtifact(ctx context.Context, path string, maximum int, description string) ([]byte, error) {
	return commandio.Await(ctx, func() ([]byte, error) { return readLocalArtifact(path, maximum, description) })
}
