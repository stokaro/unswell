package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

// runScreen builds the exploratory screening of one partition. It tests
// every rule of the role stratum on the generate/neutral responses of the
// given runs against the H0 units. Each test carries the bootstrap p-value,
// the Benjamini-Hochberg step, the card state, and the sample-size record.
// The plan supplies the partition of every source.
type screenOptions struct {
	classes, plan             string
	records, tasks, findings  []string
	partition, unitKind, role string
	fdr, mid                  float64
}

func runScreen(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := screenFlags(args[1:])
	if err != nil {
		return err
	}
	plan, err := loadDatasetPlan(ctx, options.plan)
	if err != nil {
		return err
	}
	runs, err := loadScreenRuns(ctx, options.records, options.tasks)
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
	settings := patterns.ScreenOptions{Partition: options.partition, UnitKind: options.unitKind, Role: options.role,
		FDR: options.fdr, MID: options.mid}
	screening, err := patterns.AnalyzeScreen(ctx, plan, runs, inputs, classes, settings)
	if err != nil {
		return err
	}
	return writeResult(ctx, "screen", screening, output)
}

func loadDatasetPlan(ctx context.Context, path string) (corpus.DatasetPlan, error) {
	data, err := loadLocalArtifact(ctx, path, corpus.MaxArtifactBytes, "dataset plan")
	if err != nil {
		return corpus.DatasetPlan{}, err
	}
	return corpus.LoadDatasetPlan(ctx, data)
}

// loadScreenRuns pairs each records file with the task set given in the
// same position.
func loadScreenRuns(ctx context.Context, records, tasks []string) ([]patterns.ScreenRun, error) {
	runs := make([]patterns.ScreenRun, 0, len(records))
	for i := range records {
		generation, err := loadGeneration(ctx, records[i])
		if err != nil {
			return nil, err
		}
		taskSet, _, err := loadTasks(ctx, tasks[i])
		if err != nil {
			return nil, err
		}
		runs = append(runs, patterns.ScreenRun{Records: generation, Tasks: taskSet})
	}
	return runs, nil
}

func screenFlags(args []string) (screenOptions, error) {
	var options screenOptions
	flags := flag.NewFlagSet("corpus screen", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.classes, "classes", "", "Committed rule-class file")
	flags.StringVar(&options.plan, "plan", "", "Dataset plan that assigns every source to a partition")
	flags.StringVar(&options.partition, "partition", patterns.DefaultPartition, "Partition to screen")
	flags.StringVar(&options.unitKind, "unit-kind", patterns.DefaultUnitKind, "Unit kind of the prevalence")
	flags.StringVar(&options.role, "role", patterns.DefaultRole, "Role stratum of the tasks and the H0 units")
	flags.Float64Var(&options.fdr, "fdr", patterns.DefaultFDR, "False discovery rate of the Benjamini-Hochberg step")
	flags.Float64Var(&options.mid, "mid", patterns.DefaultMID, "Minimum useful difference in unit prevalence")
	flags.Func("records", "Generation record of a run that enters selection; repeat per run", appendFlag(&options.records))
	flags.Func("tasks", "Task set of the run given in the same position; repeat per run", appendFlag(&options.tasks))
	flags.Func("findings", "Finding artifact; repeat for every shard", appendFlag(&options.findings))
	if err := flags.Parse(args); err != nil {
		return screenOptions{}, err
	}
	if flags.NArg() != 0 || options.classes == "" || options.plan == "" || len(options.records) == 0 ||
		len(options.records) != len(options.tasks) || len(options.findings) == 0 || len(options.findings) > patterns.MaxInputs {
		return screenOptions{}, fmt.Errorf("screen requires --classes, --plan, paired --records and --tasks, and 1 through %d --findings",
			patterns.MaxInputs)
	}
	return options, nil
}

// appendFlag collects every value of a repeatable flag.
func appendFlag(target *[]string) func(string) error {
	return func(value string) error {
		*target = append(*target, value)
		return nil
	}
}
