// Command dependencyprobe measures a local GoSpacy model without downloading it.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/stokaro/unswell/research/dependencies/internal/provenance"
)

// buildCommit is set by the reproducible build command in the README.
var buildCommit = "unknown"

type options struct {
	model  string
	input  string
	repeat int
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(ctx context.Context, args []string, out, stderr io.Writer) error {
	flags := flag.NewFlagSet("dependencyprobe", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var opts options
	buildInfo := flags.Bool("build-info", false, "Print build and dependency identities without loading a model")
	flags.StringVar(&opts.model, "model", "", "Local en_core_web_sm 3.8.0 directory")
	flags.StringVar(&opts.input, "input", "", "JSON array of source cases")
	flags.IntVar(&opts.repeat, "repeat", 1, "Measured repetitions per source (1 through 10000)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *buildInfo {
		if flags.NArg() != 0 || flags.NFlag() != 1 {
			return fmt.Errorf("--build-info cannot be combined with analysis options or positional arguments")
		}
		return writeBuildInfo(out)
	}
	if err := validateAnalysis(opts, flags.Args()); err != nil {
		return err
	}
	result, err := evaluate(ctx, opts)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func validateAnalysis(opts options, positional []string) error {
	if opts.model == "" || opts.input == "" || len(positional) != 0 || opts.repeat < 1 || opts.repeat > 10000 {
		return fmt.Errorf("require --model, --input, and repeat in [1,10000], without positional arguments")
	}
	return nil
}

type metadata struct {
	UnswellCommit string              `json:"unswell_commit"`
	Go            string              `json:"go"`
	Platform      string              `json:"platform"`
	Provider      provenance.Provider `json:"provider"`
}

type observation struct {
	metadata
	Schema       string            `json:"schema"`
	ModelFiles   map[string]string `json:"model_files_sha256"`
	ModelBytes   int64             `json:"model_bytes"`
	InputHash    string            `json:"input_sha256"`
	LoadNS       int64             `json:"load_ns"`
	AnalysisNS   int64             `json:"analysis_ns"`
	Repeat       int               `json:"repeat"`
	Sources      []parsedSource    `json:"sources"`
	AllocBytes   uint64            `json:"analysis_allocated_bytes"`
	HeapRetained uint64            `json:"heap_retained_bytes"`
}

func newObservation(repeat int) (observation, error) {
	meta, err := buildMetadata()
	return observation{metadata: meta, Schema: "unswell-dependency-probe-v2", Repeat: repeat}, err
}

func buildMetadata() (metadata, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return metadata{}, fmt.Errorf("dependency provenance: build information unavailable")
	}
	provider, err := provenance.FromBuildInfo(info)
	if err != nil {
		return metadata{}, err
	}
	return metadata{
		UnswellCommit: buildCommit, Go: runtime.Version(), Platform: runtime.GOOS + "/" + runtime.GOARCH, Provider: provider,
	}, nil
}

func writeBuildInfo(out io.Writer) error {
	meta, err := buildMetadata()
	if err != nil {
		return err
	}
	result := struct {
		metadata
		Schema string `json:"schema"`
	}{metadata: meta, Schema: "unswell-dependency-probe-build-v1"}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func elapsedNS(start time.Time) int64 { return time.Since(start).Nanoseconds() }
