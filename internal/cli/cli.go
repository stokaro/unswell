// Package cli owns command parsing, filesystem discovery, streams, and exit codes.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/stokaro/unswell"
)

// Environment makes the working directory and streams explicit in CLI tests.
type Environment struct {
	Dir string
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

type trackedWriter struct {
	writer io.Writer
	err    error
}

func (w *trackedWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if w.err == nil {
		w.err = err
	}
	return n, err
}

// Run returns 0 for a complete pass, 1 for a complete policy failure, and 2 for
// operational errors. Cancellation returns 130. It never exits the process.
func Run(ctx context.Context, args []string, environment Environment) int {
	out := &trackedWriter{writer: environment.Out}
	errOut := &trackedWriter{writer: environment.Err}
	environment.Out = out
	environment.Err = errOut
	root := &cobra.Command{
		Use:           "unswell",
		Short:         "Reduce AI-style wording in code and documentation",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       unswell.Version,
	}
	root.SetArgs(args)
	root.SetIn(environment.In)
	root.SetOut(out)
	root.SetErr(errOut)
	root.CompletionOptions.DisableDefaultCmd = true
	code := 0
	root.AddCommand(
		checkCommand(environment, &code),
		reportCommand(environment),
		configCommand(environment),
		rulesCommand(environment),
		doctorCommand(environment),
		explainCommand(environment),
		baselineCommand(environment, &code),
	)
	root.AddCommand(
		&cobra.Command{
			Use:  "version",
			Args: cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error { _, err := fmt.Fprintln(out, unswell.Version); return err },
		},
	)
	if err := root.ExecuteContext(ctx); err != nil {
		_, writeErr := fmt.Fprintln(errOut, "unswell:", err)
		if writeErr != nil {
			return 2
		}
		if errors.Is(err, context.Canceled) {
			return 130
		}
		return 2
	}
	if out.err != nil || errOut.err != nil {
		return 2
	}
	return code
}

type checkOptions struct {
	changedFrom        string
	baseline           string
	gateMode           string
	collectBaseline    bool
	config             string
	projectRoot        string
	allowOutsideConfig bool
	ruleSets           []string
	profile            string
	stdin              bool
	filename           string
	format             string
	reports            []string
	jobs               int
	timeout            time.Duration
	includeSource      bool
	noGate             bool
	allowEmpty         bool
	minSeverity        string
	maxFindings        int
}

func checkCommand(environment Environment, code *int) *cobra.Command {
	var options checkOptions
	command := &cobra.Command{
		Use:   "check [paths...]",
		Short: "Analyze prose and evaluate the configured editorial gate",
		RunE: func(command *cobra.Command, args []string) error {
			result, inputs, err := analyze(command.Context(), environment, options, args)
			if result.SchemaVersion != "" {
				if reportErr := writeReports(environment, options, result, inputs); reportErr != nil {
					return reportErr
				}
			}
			if err != nil {
				return err
			}
			if !result.Gate.Passed && !options.noGate {
				*code = 1
			}
			return nil
		},
	}
	checkFlags(command, &options)
	return command
}

func checkFlags(command *cobra.Command, options *checkOptions) {
	flags := command.Flags()
	flags.StringVar(&options.changedFrom, "changed-from", "",
		"Check committed changes from merge-base(REF, HEAD) with clean-source verification")
	flags.StringVar(&options.baseline, "baseline", "", "Read accepted debt from this local baseline; checking never writes it")
	flags.StringVar(&options.gateMode, "gate-mode", "", "Override gate mode: all or new (new requires --baseline)")
	flags.StringVar(&options.config, "config", "", "Use this exact YAML configuration file")
	configurationFlags(command, options, false)
	flags.StringArrayVar(&options.ruleSets, "ruleset", nil, "Load a local declarative ruleset file or directory; repeat to combine")
	flags.StringVar(&options.profile, "profile", "", "Select a builtin profile (cannot be combined with a config file)")
	flags.BoolVar(&options.stdin, "stdin", false, "Read source bytes from stdin")
	flags.StringVar(&options.filename, "filename", "", "Logical filename for stdin")
	flags.StringVar(&options.format, "format", "", "Input syntax: text, markdown, go, javascript, typescript, tsx, "+
		"python, rust, java, c, cpp, csharp, yaml, bash, sh, zsh, fish, powershell")
	flags.StringArrayVar(&options.reports, "report", nil, "Report format:path; repeat for one analysis and multiple reports")
	flags.IntVar(&options.jobs, "jobs", 1, "Concurrent source workers (1–64)")
	flags.DurationVar(&options.timeout, "timeout", 30*time.Second, "Analysis timeout")
	flags.BoolVar(&options.includeSource, "include-source", false, "Include source text in saved reports and HTML highlights")
	flags.BoolVar(&options.noGate, "no-gate", false, "Report findings without a style failure exit code")
	flags.BoolVar(&options.allowEmpty, "allow-empty", false, "Explicitly allow a scan without applicable prose")
	flags.StringVar(&options.minSeverity, "min-severity", "note", "Display severity floor; never changes the gate")
	flags.IntVar(&options.maxFindings, "max-findings", 0, "Presentation limit; zero displays all findings")
}
