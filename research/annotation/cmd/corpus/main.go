// Command corpus prepares source groups, binds annotations, and fits research models.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"

	"github.com/stokaro/unswell/probability"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/training"
)

func main() { os.Exit(mainCode()) }

func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout); err != nil {
		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return 130
		}
		_, _ = commandio.Await(ctx, func() (int, error) { return fmt.Fprintln(os.Stderr, err) })
		if ctx.Err() != nil {
			return 130
		}
		return 2
	}
	return 0
}

func run(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: corpus" +
			" {acquire|plan|extract|verify|join|measure|analyze|train|predict|evaluate|compare|reference-bank|pack|figures" +
			"|dataset|duplicates|first-appearance|dedupe|tasks|requests|generations|paired} [options] < artifact.json")
	}
	if command := dedicated(args[0]); command != nil {
		return command(ctx, args, input, output)
	}
	if !slices.Contains([]string{"plan", "extract", "verify", "join", "measure", "train"}, args[0]) {
		return fmt.Errorf("unknown corpus command %q", args[0])
	}
	options, err := commandOptions(args)
	if err != nil {
		return err
	}
	data, err := readInput(ctx, args[0], input)
	if err != nil {
		return err
	}
	result, err := operation(ctx, args[0], options, data)
	if err != nil {
		return err
	}
	return writeResult(ctx, args[0], result, output)
}

// dedicated returns the commands that own their own flags and input handling.
func dedicated(name string) func(context.Context, []string, io.Reader, io.Writer) error {
	commands := map[string]func(context.Context, []string, io.Reader, io.Writer) error{
		"pack": runPack, "reference-bank": runCompressionBank, "compare": runComparison, "predict": runEvaluation,
		"evaluate": runEvaluation, "figures": runFigures, "dataset": runDataset, "duplicates": runDuplicates,
		"analyze": runAnalyze, "acquire": runAcquire, "first-appearance": runFirstAppearance, "dedupe": runDedupe,
		"tasks": runTasks, "requests": runRequests, "generations": runGenerations, "paired": runPaired,
	}
	return commands[name]
}

// outputLimit bounds each command's serialized result.
func outputLimit(name string) int {
	switch name {
	case "pack":
		return probability.MaxBytes
	case "reference-bank":
		return corpus.MaxCompressionBankBytes
	case "train":
		return training.MaxArtifactBytes
	case "predict", "evaluate", "compare":
		return training.MaxPredictionBytes
	case "figures":
		return training.MaxPredictionBytes
	case "duplicates":
		return corpus.MaxManifestBytes
	}
	return corpus.MaxArtifactBytes
}

func writeResult(ctx context.Context, name string, result any, output io.Writer) error {
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if len(encoded) >= outputLimit(name) {
		return fmt.Errorf("output exceeds artifact size limit")
	}
	count, err := commandio.Await(ctx, func() (int, error) { return output.Write(append(encoded, '\n')) })
	if err == nil && count != len(encoded)+1 {
		return io.ErrShortWrite
	}
	return err
}

func readInput(ctx context.Context, name string, input io.Reader) ([]byte, error) {
	maximum := corpus.MaxArtifactBytes
	if name == "plan" || name == "extract" {
		maximum = corpus.MaxManifestBytes
	}
	if name == "pack" {
		maximum = training.MaxArtifactBytes
	}
	return commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(input, int64(maximum)+1))
	})
}

func operation(ctx context.Context, name string, options options, data []byte) (any, error) {
	if name == "plan" {
		manifest, err := corpus.LoadManifest(ctx, data)
		if err != nil {
			return nil, err
		}
		return corpus.MakePlan(ctx, manifest)
	}
	var plan corpus.Plan
	var artifact corpus.Artifact
	var err error
	if name == "extract" {
		plan, err = corpus.LoadPlan(ctx, data)
	} else {
		artifact, err = corpus.LoadArtifact(ctx, data)
		plan = artifact.Plan
	}
	if err != nil {
		return nil, err
	}
	files, err := loadFiles(ctx, options.root, plan)
	if err != nil {
		return nil, err
	}
	switch name {
	case "verify":
		return corpus.Verify(ctx, artifact, files)
	case "measure":
		return measureOperation(ctx, options, artifact, files)
	case "join", "train":
		return annotatedOperation(ctx, name, options, artifact, files)
	}
	return corpus.Build(ctx, plan, files)
}

func measureOperation(ctx context.Context, options options, artifact corpus.Artifact,
	files map[string][]byte,
) (any, error) {
	policy, err := commandio.Await(ctx, func() ([]byte, error) { return readPolicy(options.policy) })
	if err != nil {
		return nil, err
	}
	return corpus.MeasureFindings(ctx, artifact, files, policy)
}

func annotatedOperation(ctx context.Context, name string, options options, artifact corpus.Artifact,
	files map[string][]byte,
) (any, error) {
	if options.labels != "" {
		return provenanceOperation(ctx, name, options, artifact, files)
	}
	round, err := loadRound(ctx, options.round)
	if err != nil {
		return nil, err
	}
	if options.lexical {
		return training.RunLexical(ctx, artifact, round, files, options.train, options.lexicalOptions)
	}
	if options.ruleConfig != "" {
		return ruleOperation(ctx, name, options, artifact, round, files)
	}
	if name == "train" {
		return training.Run(ctx, artifact, round, files, options.train)
	}
	return corpus.Join(ctx, artifact, round, files, options.features)
}

// provenanceOperation labels the artifact from its declared cohorts and
// origins and takes the same join or train path a round would, including the
// lexical and rule-activation baselines.
func provenanceOperation(ctx context.Context, name string, options options, artifact corpus.Artifact,
	files map[string][]byte,
) (any, error) {
	decisions, err := labelDecisions(ctx, options.labels, artifact)
	if err != nil {
		return nil, err
	}
	if options.lexical {
		return training.RunLexicalDecisions(ctx, artifact, decisions, files, options.train, options.lexicalOptions)
	}
	if options.ruleConfig != "" {
		configuration, err := commandio.Await(ctx, func() ([]byte, error) { return readRuleConfig(options.ruleConfig) })
		if err != nil {
			return nil, err
		}
		if name == "train" {
			return training.RunRulesDecisions(ctx, artifact, decisions, files, options.train, configuration)
		}
		return corpus.JoinRulesDecisions(ctx, artifact, decisions, files, options.features, configuration)
	}
	if name == "train" {
		return training.RunDecisions(ctx, artifact, decisions, files, options.train)
	}
	return corpus.JoinDecisions(ctx, artifact, decisions, files, options.features)
}
