package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

type appearanceOptions struct {
	kind    string
	earlier []string
	later   []string
}

// runFirstAppearance builds a first-appearance filter from the candidate
// artifacts of an earlier and a later cohort. Artifacts are read one at a
// time and dropped after indexing, so a whole dataset need not fit in memory.
func runFirstAppearance(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := appearanceFlags(args[1:])
	if err != nil {
		return err
	}
	builder, err := corpus.NewAppearance(options.kind)
	if err != nil {
		return err
	}
	for _, path := range options.earlier {
		if err := addAppearance(ctx, builder.AddEarlier, path); err != nil {
			return err
		}
	}
	for _, path := range options.later {
		if err := addAppearance(ctx, builder.AddLater, path); err != nil {
			return err
		}
	}
	filter, err := builder.Filter()
	if err != nil {
		return err
	}
	return writeResult(ctx, "first-appearance", filter, output)
}

func addAppearance(ctx context.Context, add func(context.Context, corpus.Artifact) error, path string) error {
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
	if err := add(ctx, artifact); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func appearanceFlags(args []string) (appearanceOptions, error) {
	var options appearanceOptions
	flags := flag.NewFlagSet("corpus first-appearance", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.kind, "unit-kind", "", "Unit kind the filter compares")
	flags.Func("earlier", "Candidate artifact of the earlier cohort; repeat for every shard", func(value string) error {
		options.earlier = append(options.earlier, value)
		return nil
	})
	flags.Func("later", "Candidate artifact of the later cohort; repeat for every shard", func(value string) error {
		options.later = append(options.later, value)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return appearanceOptions{}, err
	}
	if flags.NArg() != 0 || options.kind == "" || len(options.earlier) == 0 || len(options.later) == 0 ||
		len(options.earlier)+len(options.later) > corpus.MaxShards {
		return appearanceOptions{}, fmt.Errorf("first-appearance requires --unit-kind, --earlier, and --later, at most %d artifacts",
			corpus.MaxShards)
	}
	return options, nil
}
