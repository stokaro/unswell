package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

type dedupeOptions struct {
	kind       string
	order      []string
	candidates []string
}

// runDedupe builds a unit selection that keeps the first occurrence of each
// unit text across the cohorts in the stated order. Artifacts are read one at
// a time and dropped after indexing.
func runDedupe(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := dedupeFlags(args[1:])
	if err != nil {
		return err
	}
	builder, err := corpus.NewDedupe(options.kind, options.order)
	if err != nil {
		return err
	}
	for _, path := range options.candidates {
		if err := addAppearance(ctx, builder.Add, path); err != nil {
			return err
		}
	}
	selection, err := builder.Selection()
	if err != nil {
		return err
	}
	return writeResult(ctx, "dedupe", selection, output)
}

func dedupeFlags(args []string) (dedupeOptions, error) {
	var options dedupeOptions
	var order string
	flags := flag.NewFlagSet("corpus dedupe", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.kind, "unit-kind", "", "Unit kind the selection compares")
	flags.StringVar(&order, "order", "", "Comma-separated cohorts, earliest first")
	flags.Func("candidates", "Candidate artifact; repeat for every shard", func(value string) error {
		options.candidates = append(options.candidates, value)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		return dedupeOptions{}, err
	}
	for cohort := range strings.SplitSeq(order, ",") {
		if cohort = strings.TrimSpace(cohort); cohort != "" {
			options.order = append(options.order, cohort)
		}
	}
	if flags.NArg() != 0 || options.kind == "" || len(options.order) == 0 || len(options.candidates) == 0 ||
		len(options.candidates) > corpus.MaxShards {
		return dedupeOptions{}, fmt.Errorf("dedupe requires --unit-kind, --order, and 1 through %d --candidates", corpus.MaxShards)
	}
	return options, nil
}
