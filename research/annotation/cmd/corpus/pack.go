package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/stokaro/unswell/probability"
	"github.com/stokaro/unswell/research/annotation/training"
)

type packOptions struct {
	id         string
	task       string
	evaluation string
	minWords   int
	accepted   bool
}

func packFlags(args []string) (packOptions, error) {
	var options packOptions
	flags := flag.NewFlagSet("corpus pack", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.id, "id", "", "Pack identifier recorded in results")
	flags.StringVar(&options.task, "task", probability.Task, "Estimation target: editorial_needs_revision or origin_endpoint")
	flags.StringVar(&options.evaluation, "evaluation", "", "Published evaluation reference; required to declare acceptance")
	flags.IntVar(&options.minWords, "min-words", 0, "Applicability floor chosen from validation data")
	flags.BoolVar(&options.accepted, "accepted", false, "Declare acceptance; needs a qualified corpus and an evaluation")
	if err := flags.Parse(args[1:]); err != nil {
		return packOptions{}, err
	}
	if flags.NArg() != 0 || options.id == "" || options.minWords < 1 {
		return packOptions{}, fmt.Errorf("pack requires --id, --min-words, and no positional arguments")
	}
	return options, nil
}

// runPack converts a fitted training artifact into a product pack. It adds only
// the decisions the artifact cannot hold: an identifier, an applicability floor,
// the estimation target, and any acceptance the maintainer is prepared to state.
func runPack(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	options, err := packFlags(args)
	if err != nil {
		return err
	}
	data, err := readInput(ctx, "pack", input)
	if err != nil {
		return err
	}
	artifact, err := training.Load(ctx, data)
	if err != nil {
		return err
	}
	file, err := training.BuildPack(ctx, artifact, training.PackOptions{ID: options.id, MinWords: options.minWords,
		Task: options.task, Accepted: options.accepted, Evaluation: options.evaluation})
	if err != nil {
		return err
	}
	// A pack must load through the same contract the engine applies.
	encoded, err := json.Marshal(file)
	if err != nil {
		return err
	}
	if _, err := probability.Load(ctx, encoded); err != nil {
		return fmt.Errorf("the generated pack does not satisfy the engine contract: %w", err)
	}
	return writeResult(ctx, "pack", file, output)
}
