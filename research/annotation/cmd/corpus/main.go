// Command corpus plans source groups and prepares or verifies unlabeled candidates.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
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
	if len(args) == 0 || !slices.Contains([]string{"plan", "extract", "verify"}, args[0]) {
		return fmt.Errorf("usage: corpus {plan|extract|verify} [--root source-directory] < artifact.json")
	}
	root, err := sourceRoot(args)
	if err != nil {
		return err
	}
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(input, corpus.MaxArtifactBytes+1))
	})
	if err != nil {
		return err
	}
	result, err := operation(ctx, args[0], root, data)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if len(encoded) >= corpus.MaxArtifactBytes {
		return fmt.Errorf("output exceeds artifact size limit")
	}
	count, err := commandio.Await(ctx, func() (int, error) { return output.Write(append(encoded, '\n')) })
	if err == nil && count != len(encoded)+1 {
		return io.ErrShortWrite
	}
	return err
}

func operation(ctx context.Context, name, root string, data []byte) (any, error) {
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
	files, err := loadFiles(ctx, root, plan)
	if err != nil {
		return nil, err
	}
	if name == "verify" {
		return corpus.Verify(ctx, artifact, files)
	}
	return corpus.Build(ctx, plan, files)
}

func sourceRoot(args []string) (string, error) {
	flags := flag.NewFlagSet("corpus", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "", "Local directory containing exactly pinned source and notice files")
	if err := flags.Parse(args[1:]); err != nil {
		return "", err
	}
	if flags.NArg() != 0 || (args[0] == "plan") != (*root == "") {
		return "", fmt.Errorf("only extract and verify require --root")
	}
	return *root, nil
}
