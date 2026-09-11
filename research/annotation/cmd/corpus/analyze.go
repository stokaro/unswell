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
	classes    string
	findings   []string
	baseline   string
	unitKind   string
	appearance string
	selection  string
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
	inputs, err := loadFindingInputs(ctx, options.findings)
	if err != nil {
		return err
	}
	settings, err := analysisSettings(ctx, options)
	if err != nil {
		return err
	}
	tables, err := patterns.Analyze(ctx, inputs, classes, settings)
	if err != nil {
		return err
	}
	return writeResult(ctx, "analyze", tables, output)
}

func loadFindingInputs(ctx context.Context, paths []string) ([]corpus.FindingsArtifact, error) {
	inputs := make([]corpus.FindingsArtifact, 0, len(paths))
	for _, path := range paths {
		data, err := commandio.Await(ctx, func() ([]byte, error) {
			return readLocalArtifact(path, corpus.MaxArtifactBytes, "finding artifact")
		})
		if err != nil {
			return nil, err
		}
		artifact, err := corpus.LoadFindings(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		inputs = append(inputs, artifact)
	}
	return inputs, nil
}

// analysisSettings reads the optional filter and selection named by the flags.
func analysisSettings(ctx context.Context, options analyzeOptions) (patterns.Options, error) {
	settings := patterns.Options{Baseline: options.baseline, UnitKind: options.unitKind}
	if options.appearance != "" {
		filter, err := loadAppearance(ctx, options.appearance)
		if err != nil {
			return patterns.Options{}, err
		}
		settings.FirstAppearance = &filter
	}
	if options.selection != "" {
		selection, err := loadSelection(ctx, options.selection)
		if err != nil {
			return patterns.Options{}, err
		}
		settings.Selection = &selection
	}
	return settings, nil
}

func loadAppearance(ctx context.Context, path string) (corpus.AppearanceFilter, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(path, corpus.MaxArtifactBytes, "first-appearance filter")
	})
	if err != nil {
		return corpus.AppearanceFilter{}, err
	}
	filter, err := corpus.LoadAppearanceFilter(ctx, data)
	if err != nil {
		return corpus.AppearanceFilter{}, fmt.Errorf("%s: %w", path, err)
	}
	return filter, nil
}

func loadSelection(ctx context.Context, path string) (corpus.UnitSelection, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(path, corpus.MaxArtifactBytes, "unit selection")
	})
	if err != nil {
		return corpus.UnitSelection{}, err
	}
	selection, err := corpus.LoadUnitSelection(ctx, data)
	if err != nil {
		return corpus.UnitSelection{}, fmt.Errorf("%s: %w", path, err)
	}
	return selection, nil
}

func analyzeFlags(args []string) (analyzeOptions, error) {
	var options analyzeOptions
	flags := flag.NewFlagSet("corpus analyze", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.classes, "classes", "", "Committed rule-class file")
	flags.StringVar(&options.baseline, "baseline", patterns.DefaultBaseline, "Cohort the contrasts compare against")
	flags.StringVar(&options.unitKind, "unit-kind", "", "Count units of this kind instead of documents")
	flags.StringVar(&options.appearance, "first-appearance", "", "Filter from corpus first-appearance; needs --unit-kind")
	flags.StringVar(&options.selection, "selection", "", "Selection from corpus dedupe; needs --unit-kind")
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
	if (options.appearance != "" || options.selection != "") && options.unitKind == "" {
		return analyzeOptions{}, fmt.Errorf("analyze --first-appearance and --selection require --unit-kind")
	}
	return options, nil
}
