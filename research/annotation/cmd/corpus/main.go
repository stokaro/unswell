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
	if len(args) == 0 || !slices.Contains([]string{"plan", "extract", "verify", "join", "train"}, args[0]) {
		return fmt.Errorf("usage: corpus {plan|extract|verify|join|train} [options] < artifact.json")
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

func writeResult(ctx context.Context, name string, result any, output io.Writer) error {
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	maximum := corpus.MaxArtifactBytes
	if name == "train" {
		maximum = training.MaxArtifactBytes
	}
	if len(encoded) >= maximum {
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
	if name == "verify" {
		return corpus.Verify(ctx, artifact, files)
	}
	if name == "join" || name == "train" {
		return annotatedOperation(ctx, name, options, artifact, files)
	}
	return corpus.Build(ctx, plan, files)
}

func annotatedOperation(ctx context.Context, name string, options options, artifact corpus.Artifact,
	files map[string][]byte,
) (any, error) {
	round, err := loadRound(ctx, options.round)
	if err != nil {
		return nil, err
	}
	if name == "train" {
		return training.Run(ctx, artifact, round, files, options.train)
	}
	return corpus.Join(ctx, artifact, round, files, options.features)
}
