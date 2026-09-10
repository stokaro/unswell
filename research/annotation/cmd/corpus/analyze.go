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

type analyzeOptions struct {
	classes  string
	findings []string
	baseline string
}

// runAnalyze builds the E1 pattern tables from finding artifacts measured
// under one policy and the committed rule classes. It reads nothing from
// stdin; every input is an explicit local file.
func runAnalyze(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := analyzeFlags(args[1:])
	if err != nil {
		return err
	}
	classData, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(options.classes, corpus.MaxRuleClassBytes, "rule classes")
	})
	if err != nil {
		return err
	}
	classes, err := corpus.LoadRuleClasses(ctx, classData)
	if err != nil {
		return err
	}
	inputs := make([]corpus.FindingsArtifact, 0, len(options.findings))
	for _, path := range options.findings {
		data, err := commandio.Await(ctx, func() ([]byte, error) {
			return readLocalArtifact(path, corpus.MaxArtifactBytes, "finding artifact")
		})
		if err != nil {
			return err
		}
		artifact, err := corpus.LoadFindings(ctx, data)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		inputs = append(inputs, artifact)
	}
	tables, err := patterns.Analyze(ctx, inputs, classes, patterns.Options{Baseline: options.baseline})
	if err != nil {
		return err
	}
	return writeResult(ctx, "analyze", tables, output)
}

func analyzeFlags(args []string) (analyzeOptions, error) {
	var options analyzeOptions
	flags := flag.NewFlagSet("corpus analyze", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.classes, "classes", "", "Committed rule-class file")
	flags.StringVar(&options.baseline, "baseline", patterns.DefaultBaseline, "Cohort the contrasts compare against")
	flags.Func("findings", "Finding artifact; repeat for every shard", func(value string) error {
		options.findings = append(options.findings, value)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return analyzeOptions{}, err
	}
	if flags.NArg() != 0 || options.classes == "" || len(options.findings) == 0 || len(options.findings) > patterns.MaxInputs {
		return analyzeOptions{}, fmt.Errorf("analyze requires --classes and 1 through %d --findings", patterns.MaxInputs)
	}
	return options, nil
}
