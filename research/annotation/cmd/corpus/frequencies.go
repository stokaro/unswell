package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

type frequenciesOptions struct {
	candidates    []string
	baseline      string
	targets       []string
	minCount      int
	minComponents int
	top           int
}

// runFrequencies counts word n-grams, sentence openers, and part-of-speech
// templates per cohort and role over candidate artifacts, and contrasts every
// stratum with the baseline cohort of the same role. It proposes candidate
// constructions; it decides nothing. Two passes read the artifacts: counts
// first, then the component support of the kept keys.
func runFrequencies(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := frequenciesFlags(args[1:])
	if err != nil {
		return err
	}
	frequencies, err := patterns.NewFrequencies(patterns.FrequencyOptions{Baseline: options.baseline,
		Targets: options.targets, MinCount: options.minCount, MinComponents: options.minComponents, Top: options.top})
	if err != nil {
		return err
	}
	if err := eachCandidateArtifact(ctx, options.candidates, frequencies.Add); err != nil {
		return err
	}
	frequencies.Select()
	if err := eachCandidateArtifact(ctx, options.candidates, frequencies.Support); err != nil {
		return err
	}
	return writeResult(ctx, "frequencies", frequencies.Tables(), output)
}

// eachCandidateArtifact loads the candidate artifacts one at a time, so the
// pass holds one artifact and the counts in memory.
func eachCandidateArtifact(ctx context.Context, paths []string, visit func(context.Context, corpus.Artifact) error) error {
	for _, path := range paths {
		data, err := commandio.Await(ctx, func() ([]byte, error) {
			return readLocalArtifact(path, corpus.MaxArtifactBytes, "candidate artifact")
		})
		if err != nil {
			return err
		}
		artifact, err := corpus.LoadArtifact(ctx, data)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := visit(ctx, artifact); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}

func frequenciesFlags(args []string) (frequenciesOptions, error) {
	var options frequenciesOptions
	flags := flag.NewFlagSet("corpus frequencies", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.baseline, "baseline", patterns.DefaultBaseline, "Cohort the contrasts compare against")
	flags.IntVar(&options.minCount, "min-count", 0, "Count a key needs in one stratum to enter the tables")
	flags.IntVar(&options.minComponents, "min-components", 0, "Components a key needs in the target stratum to enter a contrast")
	flags.IntVar(&options.top, "top", 0, "Contrasts kept per measure and target stratum")
	flags.Func("target", "Cohort that gets contrasts; repeat for several, default every cohort but the baseline",
		func(value string) error {
			options.targets = append(options.targets, value)
			return nil
		})
	flags.Func("candidates", "Candidate artifact; repeat for every shard", func(value string) error {
		options.candidates = append(options.candidates, value)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return frequenciesOptions{}, err
	}
	if flags.NArg() != 0 || len(options.candidates) == 0 || len(options.candidates) > patterns.MaxInputs ||
		options.minCount < 0 || options.minComponents < 0 || options.top < 0 {
		return frequenciesOptions{}, fmt.Errorf("frequencies requires 1 through %d --candidates and non-negative bounds",
			patterns.MaxInputs)
	}
	return options, nil
}
