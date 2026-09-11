package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

type datasetOptions struct {
	root, output, pinned string
	union                corpus.UnionOptions
}

// runDataset plans, pins, verifies, and unites a sharded dataset. Shard
// manifests are read beneath --root by their declared paths; pinned copies are
// written beneath --output and never overwrite an existing file; a union is
// one manifest of selected shards whose paths name their checkouts.
func runDataset(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: corpus dataset {plan|pin|verify|union} --root DIR [--output DIR] [--pinned DIR] " +
			"[--id ID] [--cohort NAME]... [--repository NAME]... [--unit-kind KIND]... < input.json")
	}
	options, err := datasetFlags(args[1], args[2:])
	if err != nil {
		return err
	}
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(input, int64(corpus.MaxArtifactBytes)+1))
	})
	if err != nil {
		return err
	}
	result, err := datasetOperation(ctx, args[1], options, data)
	if err != nil {
		return err
	}
	return writeResult(ctx, "dataset", result, output)
}

func datasetFlags(name string, args []string) (datasetOptions, error) {
	var options datasetOptions
	flags := flag.NewFlagSet("corpus dataset", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.root, "root", "", "Local directory containing the shard manifests at their declared paths")
	switch name {
	case "pin":
		flags.StringVar(&options.output, "output", "", "Directory that receives the pinned shard manifests")
	case "verify":
		flags.StringVar(&options.pinned, "pinned", "", "Optional directory holding pinned shard copies to check")
	case "union":
		unionFlags(flags, &options.union)
	case "plan":
	default:
		return datasetOptions{}, fmt.Errorf("unknown dataset command %q", name)
	}
	if err := flags.Parse(args); err != nil {
		return datasetOptions{}, err
	}
	return options, validateDatasetFlags(name, flags, options)
}

func validateDatasetFlags(name string, flags *flag.FlagSet, options datasetOptions) error {
	if flags.NArg() != 0 || options.root == "" || (name == "pin") != (options.output != "") {
		return fmt.Errorf("dataset commands require --root; pin requires --output")
	}
	if name == "union" && options.union.ID == "" {
		return fmt.Errorf("dataset union requires --id")
	}
	return nil
}

func unionFlags(flags *flag.FlagSet, options *corpus.UnionOptions) {
	flags.StringVar(&options.ID, "id", "", "Identifier of the union manifest")
	appendTo := func(target *[]string) func(string) error {
		return func(value string) error {
			*target = append(*target, value)
			return nil
		}
	}
	flags.Func("cohort", "Cohort to include; repeat for a set, omit for all", appendTo(&options.Cohorts))
	flags.Func("repository", "Repository to include; repeat for a set, omit for all", appendTo(&options.Repositories))
	flags.Func("role", "Source role to include; repeat for a set, omit for all", appendTo(&options.Roles))
	flags.Func("unit-kind", "Unit kind to keep; repeat for a set, omit for the dataset's kinds", appendTo(&options.UnitKinds))
	flags.IntVar(&options.MaxPerCheckout, "max-per-checkout", 0, "Keep the first N sources of each checkout in ID order; 0 keeps all")
	flags.Func("uncapped-cohort", "Cohort whose checkouts keep every source under the limit; repeat for a set",
		appendTo(&options.UncappedCohorts))
	flags.IntVar(&options.MaxSourceBytes, "max-source-bytes", 0, "Drop sources above this many bytes; 0 keeps all")
}

func datasetOperation(ctx context.Context, name string, options datasetOptions, data []byte) (any, error) {
	if name == "plan" {
		dataset, err := corpus.LoadDataset(ctx, data)
		if err != nil {
			return nil, err
		}
		shards, err := loadShardFiles(ctx, options.root, dataset.Shards)
		if err != nil {
			return nil, err
		}
		return corpus.MakeDatasetPlan(ctx, dataset, shards)
	}
	plan, err := corpus.LoadDatasetPlan(ctx, data)
	if err != nil {
		return nil, err
	}
	shards, err := loadShardFiles(ctx, options.root, plan.Dataset.Shards)
	if err != nil {
		return nil, err
	}
	if name == "pin" {
		return pinDataset(ctx, plan, shards, options.output)
	}
	if name == "union" {
		pinned, err := corpus.PinShards(ctx, plan, shards)
		if err != nil {
			return nil, err
		}
		return corpus.UnionManifest(ctx, plan, pinned, options.union)
	}
	return verifyDataset(ctx, plan, shards, options.pinned)
}

func loadShardFiles(ctx context.Context, directory string, shards []corpus.Notice) (map[string][]byte, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	files, readErr := readFiles(ctx, root, shards)
	closeErr := root.Close()
	if readErr != nil {
		return nil, readErr
	}
	return files, closeErr
}

// DatasetPinResult lists the pinned shard files a pin command wrote.
type DatasetPinResult struct {
	Status string          `json:"status"`
	Shards []corpus.Notice `json:"shards"`
}

func pinDataset(ctx context.Context, plan corpus.DatasetPlan, shards map[string][]byte, output string) (any, error) {
	pinned, err := corpus.PinShards(ctx, plan, shards)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(output)
	if err != nil {
		return nil, err
	}
	result := DatasetPinResult{Status: "pinned_shards_written"}
	writeErr := func() error {
		for _, shard := range plan.Shards {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := writePinnedShard(root, shard.Path, pinned[shard.Path]); err != nil {
				return err
			}
			result.Shards = append(result.Shards, corpus.Notice{Path: shard.Path, SHA256: shard.PinnedSHA256, Bytes: shard.PinnedBytes})
		}
		return nil
	}()
	closeErr := root.Close()
	if writeErr != nil {
		return nil, writeErr
	}
	return result, closeErr
}

func writePinnedShard(root *os.Root, name string, data []byte) error {
	if directory := path.Dir(name); directory != "." {
		if err := root.MkdirAll(directory, 0o750); err != nil {
			return err
		}
	}
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

// DatasetVerification reports what a verify command established.
type DatasetVerification struct {
	Status  string `json:"status"`
	Shards  int    `json:"shards"`
	Groups  int    `json:"groups"`
	Sources int    `json:"sources"`
	Pinned  bool   `json:"pinned_copies_verified"`
}

func verifyDataset(ctx context.Context, plan corpus.DatasetPlan, shards map[string][]byte, pinnedDir string) (any, error) {
	if err := corpus.VerifyDatasetPlan(ctx, plan, shards); err != nil {
		return nil, err
	}
	result := DatasetVerification{Status: "dataset_plan_reproduced", Shards: len(plan.Shards), Groups: len(plan.Groups),
		Sources: len(plan.Sources)}
	if pinnedDir == "" {
		return result, nil
	}
	declared := make([]corpus.Notice, 0, len(plan.Shards))
	for _, shard := range plan.Shards {
		declared = append(declared, corpus.Notice{Path: shard.Path, SHA256: shard.PinnedSHA256, Bytes: shard.PinnedBytes})
	}
	pinned, err := loadShardFiles(ctx, pinnedDir, declared)
	if err != nil {
		return nil, err
	}
	if err := corpus.VerifyPinnedShards(ctx, plan, pinned); err != nil {
		return nil, err
	}
	result.Pinned = true
	return result, nil
}
