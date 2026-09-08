package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stokaro/unswell/baseline"
)

func baselineCommand(environment Environment, code *int) *cobra.Command {
	command := &cobra.Command{Use: "baseline", Short: "Create, update, or check explicit acceptance of existing debt"}
	check := checkCommand(environment, code)
	check.Use = "check [paths...]"
	check.Short = "Check selected sources against a baseline without changing it"
	check.PreRunE = func(command *cobra.Command, _ []string) error {
		path, err := command.Flags().GetString("baseline")
		if err != nil {
			return err
		}
		if path == "" {
			return fmt.Errorf("baseline check requires --baseline")
		}
		if !command.Flags().Changed("gate-mode") {
			return command.Flags().Set("gate-mode", "new")
		}
		return nil
	}
	command.AddCommand(captureCommand(environment, "create"), captureCommand(environment, "update"), check)
	return command
}

func captureCommand(environment Environment, operation string) *cobra.Command {
	var options checkOptions
	var output string
	command := &cobra.Command{Use: operation + " [paths...]", Short: "Explicitly accept debt from a complete scan",
		RunE: func(command *cobra.Command, args []string) error {
			return captureBaseline(command.Context(), environment, options, args, operation, output)
		}}
	checkFlags(command, &options)
	command.Flags().StringVar(&output, "output", "", "New baseline path (required for create)")
	return command
}

func captureBaseline(ctx context.Context, environment Environment, options checkOptions, args []string, operation, output string) error {
	path, err := captureDestination(environment, options, operation, output)
	if err != nil {
		return err
	}
	var previous baseline.File
	if operation == "update" {
		previous, err = readBaselineFile(ctx, path)
		if err != nil {
			return err
		}
	}
	options.collectBaseline, options.gateMode = true, "all"
	result, inputs, err := analyze(ctx, environment, options, args)
	if err != nil {
		return err
	}
	if err := protectInputs(path, inputs); err != nil {
		return err
	}
	file, err := capturedDebt(ctx, operation, previous, *result.BaselineSnapshot)
	if err != nil {
		return err
	}
	data, err := baseline.Encode(ctx, file)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := atomicWrite(path, data); err != nil {
		return err
	}
	_, err = fmt.Fprintf(environment.Out, "Baseline %s: %d accepted entries across %d documents.\n",
		operation, len(file.Entries), len(file.Documents))
	return err
}

func readBaselineFile(ctx context.Context, path string) (baseline.File, error) {
	data, err := readLimited(path, baseline.MaxBytes)
	if err != nil {
		return baseline.File{}, err
	}
	return baseline.Load(ctx, data)
}

func capturedDebt(ctx context.Context, operation string, previous baseline.File, snapshot baseline.Snapshot) (baseline.File, error) {
	if operation == "update" {
		return baseline.Update(ctx, previous, snapshot)
	}
	return baseline.Create(ctx, snapshot)
}

func captureDestination(environment Environment, options checkOptions, operation, output string) (string, error) {
	if options.stdin || len(options.reports) > 0 || options.noGate || options.gateMode != "" || options.changedFrom != "" {
		return "", fmt.Errorf("baseline capture accepts complete source paths; stdin, reports, and gate overrides are unsupported")
	}
	if operation == "update" {
		if options.baseline == "" || output != "" {
			return "", fmt.Errorf("baseline update requires --baseline and writes that file; --output is unsupported")
		}
		return absoluteArguments(environment.Dir, []string{options.baseline})[0], nil
	}
	return createDestination(environment, options, output)
}

func createDestination(environment Environment, options checkOptions, output string) (string, error) {
	if output == "" || output == "-" || options.baseline != "" {
		return "", fmt.Errorf("baseline create requires a new --output file and does not accept --baseline")
	}
	path := absoluteArguments(environment.Dir, []string{output})[0]
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		return "", fmt.Errorf("baseline destination already exists or cannot be inspected: %s", path)
	}
	return path, nil
}
