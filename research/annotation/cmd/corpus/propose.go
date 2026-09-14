package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

type proposeOptions struct {
	root     string
	baseline string
	measure  string
	role     string
	minCount int
	minLift  float64
	top      int
	config   bool
}

func proposeFlags(args []string) (proposeOptions, error) {
	var options proposeOptions
	flags := flag.NewFlagSet("corpus propose", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.root, "root", "", "Directory of the tree to measure")
	flags.StringVar(&options.baseline, "baseline", "", "Frequency tables holding the baseline key tables")
	flags.StringVar(&options.measure, "measure", patterns.MeasureClosed3, "Measure to compare")
	flags.StringVar(&options.role, "role", "comment", "Baseline role to compare against")
	flags.IntVar(&options.minCount, "min-count", 0, "Occurrences a construction needs in the tree")
	flags.Float64Var(&options.minLift, "min-lift", 0, "Factor over the baseline rate a construction needs")
	flags.IntVar(&options.top, "top", 0, "Constructions kept")
	flags.BoolVar(&options.config, "config", false,
		"Write a policy.banned-phrases fragment instead of the proposal record")
	if err := flags.Parse(args); err != nil {
		return proposeOptions{}, err
	}
	if flags.NArg() != 0 || options.root == "" || options.baseline == "" {
		return proposeOptions{}, fmt.Errorf("propose requires --root and --baseline and no positional arguments")
	}
	return options, nil
}

// runPropose measures the prose of one tree and names the constructions it uses
// more than a baseline corpus does. It writes a proposal for a person to read.
// It establishes nothing about how the tree was written: a construction can
// stand out because a project has a house style, because one author wrote most
// of it, or because the subject demands it.
func runPropose(ctx context.Context, args []string, _ io.Reader, output io.Writer) error {
	options, err := proposeFlags(args[1:])
	if err != nil {
		return err
	}
	table, err := loadBaselineTable(ctx, options)
	if err != nil {
		return err
	}
	counter, err := patterns.NewCounter(options.measure)
	if err != nil {
		return err
	}
	if err := countTree(ctx, counter, options.root); err != nil {
		return err
	}
	deviation, err := counter.Compare(table, patterns.DeviationOptions{Measure: options.measure,
		MinCount: options.minCount, MinLift: options.minLift, Top: options.top})
	if err != nil {
		return err
	}
	if options.config {
		return writeProposalConfig(ctx, deviation, output)
	}
	return writeResult(ctx, "propose", deviation, output)
}

// writeProposalConfig prints the fragment a reader pastes after judging the
// list. Each line carries its two rates and no cause. A construction stands out
// here; why it does is not something a frequency decides. The rule this feeds
// refuses text outright, so a person cuts the list down before it applies.
func writeProposalConfig(ctx context.Context, deviation patterns.Deviation, output io.Writer) error {
	var builder strings.Builder
	builder.WriteString("# Proposed from " + fmt.Sprint(deviation.TreeWords) + " prose words against a baseline of " +
		fmt.Sprint(deviation.BaselineWords) + " words, measure " + deviation.Measure + ".\n")
	builder.WriteString("# Each line stands out against that baseline. Standing out is not a defect and not\n")
	builder.WriteString("# a sign of how the text was written; read the list and keep what you agree with.\n")
	builder.WriteString("version: 1\nrules:\n  policy.banned-phrases:\n    parameters:\n      phrases:\n")
	for _, term := range deviation.Terms {
		note := fmt.Sprintf("%.3f here against %.4f in the baseline", term.TreeRate, term.BaselineRate)
		if term.Status == patterns.DeviationAbsent {
			note = fmt.Sprintf("%.3f here, below the baseline's own minimum", term.TreeRate)
		}
		builder.WriteString("        - \"" + term.Key + "\"  # " + note + "\n")
	}
	_, err := commandio.Await(ctx, func() (int, error) { return output.Write([]byte(builder.String())) })
	return err
}

func loadBaselineTable(ctx context.Context, options proposeOptions) (patterns.BaselineTable, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return readLocalArtifact(options.baseline, corpus.MaxArtifactBytes, "frequency tables")
	})
	if err != nil {
		return patterns.BaselineTable{}, err
	}
	var tables patterns.FrequencyTables
	if err := json.Unmarshal(data, &tables); err != nil {
		return patterns.BaselineTable{}, fmt.Errorf("%s: %w", options.baseline, err)
	}
	for _, table := range tables.Baselines {
		if table.Measure == options.measure && table.Role == options.role {
			return table, nil
		}
	}
	return patterns.BaselineTable{}, fmt.Errorf("%s holds no %s baseline for role %q; "+
		"produce one with corpus frequencies --baselines", options.baseline, options.measure, options.role)
}

// countTree walks the tree and counts the prose of every file whose extension
// the engine recognizes. A file it cannot parse stops the run rather than
// silently leaving a hole in the denominator.
func countTree(ctx context.Context, counter *patterns.Counter, root string) error {
	// A rooted filesystem, so a symbolic link inside the tree cannot make the
	// walk read a file outside it.
	opened, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	walkErr := fs.WalkDir(opened.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case entry.IsDir():
			return skipHidden(entry, path)
		default:
			return countFile(ctx, counter, opened.FS(), path)
		}
	})
	if closeErr := opened.Close(); walkErr == nil {
		return closeErr
	}
	return walkErr
}

// skipHidden leaves a dotted directory out of the walk. A tree's history and
// its editor state are not its prose.
func skipHidden(entry fs.DirEntry, path string) error {
	if strings.HasPrefix(entry.Name(), ".") && path != "." {
		return fs.SkipDir
	}
	return nil
}

// countFile counts one file, or none when the engine does not recognize it.
func countFile(ctx context.Context, counter *patterns.Counter, tree fs.FS, path string) error {
	data, err := fs.ReadFile(tree, path)
	if err != nil {
		return err
	}
	if len(data) > corpus.MaxSourceBytes {
		return fmt.Errorf("%s exceeds the source byte limit", path)
	}
	format, known := extract.Detect(path, data)
	if !known {
		return nil
	}
	return countDocument(ctx, counter, path, format, data)
}

func countDocument(ctx context.Context, counter *patterns.Counter, path string,
	format document.Format, data []byte,
) error {
	doc, err := extract.Parse(ctx, document.Source{Name: path, Format: format, Bytes: data},
		extract.Options{MaxBytes: corpus.MaxSourceBytes, MaxBlocks: corpus.MaxUnits})
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	for _, block := range doc.Blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := counter.Add(ctx, block.Text); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}
	return nil
}
