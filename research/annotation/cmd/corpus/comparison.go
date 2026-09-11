package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"slices"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/evaluation"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/training"
)

type comparisonOptions struct {
	evaluationOptions
	comparator string
}

func comparisonFlags(args []string) (comparisonOptions, error) {
	var options comparisonOptions
	flags := flag.NewFlagSet("corpus compare", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.plan, "plan", "", "Frozen comparison plan")
	flags.StringVar(&options.protocol, "protocol", "", "Protocol bytes matching both trials")
	flags.StringVar(&options.comparator, "comparator", "", "Saved comparator predictions")
	flags.StringVar(&options.corpus, "corpus", "", "Frozen corpus shared by both trials")
	flags.StringVar(&options.round, "round", "", "Independent evaluation annotation round")
	flags.StringVar(&options.labels, "labels", "", "Label source instead of a round: provenance or cohort, from the corpus")
	flags.BoolVar(&options.allowSimulation, "allow-simulation", false, "Allow explicitly simulated tutorial labels")
	if err := flags.Parse(args[1:]); err != nil {
		return comparisonOptions{}, err
	}
	if flags.NArg() != 0 {
		return comparisonOptions{}, fmt.Errorf("compare accepts no positional arguments")
	}
	if slices.Contains([]string{options.plan, options.protocol, options.comparator, options.corpus}, "") ||
		(options.round == "") == (options.labels == "") {
		return comparisonOptions{}, fmt.Errorf("compare requires --plan, --protocol, --comparator, --corpus, and either --round or --labels")
	}
	return options, validateLabels(options.labels)
}

func runComparison(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	options, err := comparisonFlags(args)
	if err != nil {
		return err
	}
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(input, training.MaxPredictionBytes+1))
	})
	if err != nil {
		return err
	}
	result, err := comparisonOperation(ctx, options, data)
	if err != nil {
		return err
	}
	return writeResult(ctx, "compare", result, output)
}

func comparisonOperation(ctx context.Context, options comparisonOptions, data []byte) (evaluation.ComparisonResult, error) {
	plan, err := comparisonPlan(ctx, options)
	if err != nil {
		return evaluation.ComparisonResult{}, err
	}
	other, err := loadLocalArtifact(ctx, options.comparator, training.MaxPredictionBytes, "comparator predictions")
	if err != nil {
		return evaluation.ComparisonResult{}, err
	}
	encoded, err := loadLocalArtifact(ctx, options.corpus, corpus.MaxArtifactBytes, "candidate artifact")
	if err != nil {
		return evaluation.ComparisonResult{}, err
	}
	candidates, err := corpus.LoadArtifact(ctx, encoded)
	if err != nil {
		return evaluation.ComparisonResult{}, err
	}
	if options.labels != "" {
		decisions, err := labelDecisions(ctx, options.labels, candidates)
		if err != nil {
			return evaluation.ComparisonResult{}, err
		}
		return evaluation.RunComparisonDecisions(ctx, data, other, plan, candidates, decisions, options.allowSimulation)
	}
	round, err := loadRound(ctx, options.round)
	if err != nil {
		return evaluation.ComparisonResult{}, err
	}
	return evaluation.RunComparison(ctx, data, other, plan, candidates, round, options.allowSimulation)
}

func comparisonPlan(ctx context.Context, options comparisonOptions) (evaluation.ComparisonPlan, error) {
	data, err := loadLocalArtifact(ctx, options.plan, 16<<10, "comparison plan")
	if err != nil {
		return evaluation.ComparisonPlan{}, err
	}
	plan, err := evaluation.LoadComparisonPlan(ctx, data)
	if err != nil {
		return evaluation.ComparisonPlan{}, err
	}
	protocol, err := loadLocalArtifact(ctx, options.protocol, 1<<20, "comparison protocol")
	if err != nil {
		return evaluation.ComparisonPlan{}, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(protocol)) != plan.ProtocolSHA256 {
		return evaluation.ComparisonPlan{}, fmt.Errorf("comparison protocol digest mismatch")
	}
	return plan, nil
}
