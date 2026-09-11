package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

type compressionBankOptions struct{ root, round, labels, selection string }

func compressionBankFlags(args []string) (compressionBankOptions, error) {
	var options compressionBankOptions
	flags := flag.NewFlagSet("corpus reference-bank", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.root, "root", "", "Root containing declared source and notice files")
	flags.StringVar(&options.round, "round", "", "Validated round with independently curated unit origins")
	flags.StringVar(&options.labels, "labels", "",
		"Label source instead of a round: cohort seeds historical and contemporary cohorts from the corpus")
	flags.StringVar(&options.selection, "selection", "", "Explicit ordered cohorts and compressor settings")
	if err := flags.Parse(args[1:]); err != nil {
		return compressionBankOptions{}, err
	}
	if flags.NArg() != 0 || options.root == "" || (options.round == "") == (options.labels == "") || options.selection == "" {
		return compressionBankOptions{}, fmt.Errorf(
			"reference-bank requires --root, --selection, either --round or --labels, and no positional arguments")
	}
	return options, validateLabels(options.labels)
}

func runCompressionBank(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	options, err := compressionBankFlags(args)
	if err != nil {
		return err
	}
	data, err := readInput(ctx, "reference-bank", input)
	if err != nil {
		return err
	}
	result, err := compressionBankOperation(ctx, options, data)
	if err != nil {
		return err
	}
	return writeResult(ctx, "reference-bank", result, output)
}

func compressionBankOperation(ctx context.Context, options compressionBankOptions, data []byte) (corpus.CompressionBank, error) {
	artifact, err := corpus.LoadArtifact(ctx, data)
	if err != nil {
		return corpus.CompressionBank{}, err
	}
	selection, err := loadLocalArtifact(ctx, options.selection, corpus.MaxCompressionSelectionBytes, "compression selection")
	if err != nil {
		return corpus.CompressionBank{}, err
	}
	settings, err := corpus.LoadCompressionBankOptions(ctx, selection)
	if err != nil {
		return corpus.CompressionBank{}, err
	}
	files, err := loadFiles(ctx, options.root, artifact.Plan)
	if err != nil {
		return corpus.CompressionBank{}, err
	}
	if options.labels != "" {
		decisions, err := labelDecisions(ctx, options.labels, artifact)
		if err != nil {
			return corpus.CompressionBank{}, err
		}
		return corpus.BuildCompressionBankDecisions(ctx, artifact, decisions, files, settings)
	}
	round, err := loadRound(ctx, options.round)
	if err != nil {
		return corpus.CompressionBank{}, err
	}
	return corpus.BuildCompressionBank(ctx, artifact, round, files, settings)
}
