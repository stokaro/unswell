package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/training"
)

type options struct {
	root, round string
	features    []string
	train       training.Options
}

func commandOptions(args []string) (options, error) {
	var result options
	flags := flag.NewFlagSet("corpus", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&result.root, "root", "", "Local directory containing exactly pinned source and notice files")
	flags.StringVar(&result.round, "round", "", "Explicit local annotation round for join or train")
	flags.Func("feature", "Prepared feature ID; repeat to select a set for join or train", func(value string) error {
		result.features = append(result.features, value)
		return nil
	})
	if args[0] == "train" {
		trainingFlags(flags, &result.train)
	}
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 || (args[0] == "plan") != (result.root == "") {
		return options{}, fmt.Errorf("extract, verify, join, and train require --root; plan does not accept it")
	}
	result.train.Features = result.features
	return result, result.validate(args[0])
}

func (result options) validate(name string) error {
	if name == "join" || name == "train" {
		if result.round == "" || len(result.features) == 0 {
			return fmt.Errorf("join and train require --round and at least one --feature")
		}
	} else if result.round != "" || len(result.features) != 0 {
		return fmt.Errorf("only join and train accept --round and --feature")
	}
	if name == "train" && result.train.Kind == "" {
		return fmt.Errorf("train requires --kind")
	}
	return nil
}

func loadRound(ctx context.Context, path string) (*annotation.Round, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) { return readRound(path) })
	if err != nil {
		return nil, err
	}
	return annotation.Load(ctx, data)
}

func readRound(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > annotation.MaxBytes {
		return nil, fmt.Errorf("annotation round must be a regular file within its byte limit")
	}
	// #nosec G304 -- The operator explicitly selects this local round; reads require a regular file and a byte limit.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, annotation.MaxBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	return data, closeErr
}
