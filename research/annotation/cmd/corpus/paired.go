package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

// runPaired builds the E2 paired tables of one generation run from its
// records, its tasks, the rule classes, and the finding artifacts of the
// originals and the responses.
type pairedOptions struct {
	records, tasks, classes string
	findings                []string
}

func runPaired(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := pairedFlags(args[1:])
	if err != nil {
		return err
	}
	records, err := loadGeneration(ctx, options.records)
	if err != nil {
		return err
	}
	tasks, _, err := loadTasks(ctx, options.tasks)
	if err != nil {
		return err
	}
	classes, err := loadClasses(ctx, options.classes)
	if err != nil {
		return err
	}
	inputs, err := loadFindingInputs(ctx, options.findings)
	if err != nil {
		return err
	}
	tables, err := patterns.AnalyzePaired(ctx, records, tasks, inputs, classes)
	if err != nil {
		return err
	}
	return writeResult(ctx, "paired", tables, output)
}

func loadGeneration(ctx context.Context, path string) (generation.Generation, error) {
	var records generation.Generation
	if _, err := loadJSON(ctx, path, corpus.MaxArtifactBytes, &records); err != nil {
		return generation.Generation{}, err
	}
	if records.Version != generation.GenerationVersion {
		return generation.Generation{}, fmt.Errorf("%s: unsupported generation record", path)
	}
	return records, nil
}

func loadClasses(ctx context.Context, path string) (corpus.RuleClasses, error) {
	data, err := readLocalArtifact(path, corpus.MaxRuleClassBytes, "rule classes")
	if err != nil {
		return corpus.RuleClasses{}, err
	}
	return corpus.LoadRuleClasses(ctx, data)
}

func pairedFlags(args []string) (pairedOptions, error) {
	var options pairedOptions
	flags := flag.NewFlagSet("corpus paired", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.records, "records", "", "Generation record of the run")
	flags.StringVar(&options.tasks, "tasks", "", "Sampled task set of the run")
	flags.StringVar(&options.classes, "classes", "", "Committed rule-class file")
	flags.Func("findings", "Finding artifact; repeat for every shard", func(value string) error {
		options.findings = append(options.findings, value)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return pairedOptions{}, err
	}
	if flags.NArg() != 0 || options.records == "" || options.tasks == "" || options.classes == "" || len(options.findings) == 0 ||
		len(options.findings) > patterns.MaxInputs {
		return pairedOptions{}, fmt.Errorf("paired requires --records, --tasks, --classes, and 1 through %d --findings", patterns.MaxInputs)
	}
	return options, nil
}
