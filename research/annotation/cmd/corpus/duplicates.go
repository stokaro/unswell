package main

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

type duplicateOptions struct {
	root      string
	threshold float64
	apply     bool
}

// runDuplicates reports near-duplicate sources of one manifest, or with
// --apply writes the manifest with the detected cluster keys added.
func runDuplicates(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	options, err := duplicateFlags(args[1:])
	if err != nil {
		return err
	}
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(input, int64(corpus.MaxManifestBytes)+1))
	})
	if err != nil {
		return err
	}
	manifest, err := corpus.LoadManifest(ctx, data)
	if err != nil {
		return err
	}
	files, err := loadShardFiles(ctx, options.root, sourceFiles(manifest))
	if err != nil {
		return err
	}
	report, err := corpus.DetectDuplicates(ctx, manifest, files, options.threshold)
	if err != nil {
		return err
	}
	if !options.apply {
		return writeResult(ctx, "duplicates", report, output)
	}
	applied, err := corpus.ApplyDuplicates(ctx, manifest, report)
	if err != nil {
		return err
	}
	return writeResult(ctx, "duplicates", applied, output)
}

func duplicateFlags(args []string) (duplicateOptions, error) {
	options := duplicateOptions{threshold: corpus.DefaultDuplicateThreshold}
	flags := flag.NewFlagSet("corpus duplicates", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.root, "root", "", "Local directory containing exactly pinned source files")
	flags.Float64Var(&options.threshold, "threshold", corpus.DefaultDuplicateThreshold, "Jaccard similarity that joins a cluster")
	flags.BoolVar(&options.apply, "apply", false, "Write the manifest with cluster keys instead of the report")
	if err := flags.Parse(args); err != nil {
		return duplicateOptions{}, err
	}
	if flags.NArg() != 0 || options.root == "" {
		return duplicateOptions{}, fmt.Errorf("duplicates requires --root")
	}
	return options, nil
}

func sourceFiles(manifest corpus.Manifest) []corpus.Notice {
	files := make([]corpus.Notice, 0, len(manifest.Sources))
	for _, source := range manifest.Sources {
		files = append(files, corpus.Notice{Path: source.Path, SHA256: source.SHA256, Bytes: source.Bytes})
	}
	return files
}
