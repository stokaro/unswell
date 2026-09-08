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
	"time"
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
	flags.StringVar(&opts.model, "model", "", "Local en_core_web_sm 3.8.0 directory")
	flags.StringVar(&opts.input, "input", "", "JSON array of source cases")
	flags.IntVar(&opts.repeat, "repeat", 1, "Measured repetitions per source (1 through 10000)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if opts.model == "" || opts.input == "" || flags.NArg() != 0 || opts.repeat < 1 || opts.repeat > 10000 {
		return fmt.Errorf("require --model, --input, and repeat in [1,10000], without positional arguments")
	}
	result, err := evaluate(ctx, opts)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

type observation struct {
	UnswellCommit string            `json:"unswell_commit"`
	Schema        string            `json:"schema"`
	Go            string            `json:"go"`
	Platform      string            `json:"platform"`
	Provider      string            `json:"provider"`
	Revision      string            `json:"revision"`
	ModelFiles    map[string]string `json:"model_files_sha256"`
	ModelBytes    int64             `json:"model_bytes"`
	InputHash     string            `json:"input_sha256"`
	LoadNS        int64             `json:"load_ns"`
	AnalysisNS    int64             `json:"analysis_ns"`
	Repeat        int               `json:"repeat"`
	Sources       []parsedSource    `json:"sources"`
	AllocBytes    uint64            `json:"analysis_allocated_bytes"`
	HeapRetained  uint64            `json:"heap_retained_bytes"`
}

func newObservation(repeat int) observation {
	return observation{
		UnswellCommit: buildCommit, Schema: "unswell-dependency-probe-v1", Go: runtime.Version(), Platform: runtime.GOOS + "/" + runtime.GOARCH,
		Provider: "gospacy/v3@v3.8.14-port.2", Revision: "e2766da9ab71ffc55a5967a76a474046a96402bc", Repeat: repeat,
	}
}

func elapsedNS(start time.Time) int64 { return time.Since(start).Nanoseconds() }
