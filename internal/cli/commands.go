package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/report"
	"github.com/stokaro/unswell/rule"
)

func jsonOutput(environment Environment, value any) error {
	encoder := json.NewEncoder(environment.Out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func configCommand(environment Environment) *cobra.Command {
	parent := &cobra.Command{Use: "config", Short: "Validate and inspect the effective editorial policy"}
	var options checkOptions
	var profile, file string
	parent.PersistentFlags().StringVar(&options.config, "config", "", "Exact configuration path")
	parent.PersistentFlags().StringArrayVar(&options.ruleSets, "ruleset", nil, "Local ruleset file or directory")
	configurationFlags(parent, &options, true)
	initialize := &cobra.Command{Use: "init", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		if len(options.ruleSets) > 0 {
			return fmt.Errorf("config init does not accept --ruleset; select rule packs when validating or scanning")
		}
		return initializeConfig(environment, options.config, profile)
	}}
	initialize.Flags().StringVar(&profile, "profile", "technical", "Builtin profile")
	validate := &cobra.Command{Use: "validate", Args: cobra.NoArgs, RunE: func(command *cobra.Command, _ []string) error {
		policy, err := configuredPolicy(command.Context(), environment, options)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(environment.Out, "Configuration valid:", policy.Profile, policy.Hash)
		return err
	}}
	explain := &cobra.Command{Use: "explain", Args: cobra.NoArgs, RunE: func(command *cobra.Command, _ []string) error {
		options.filename = file
		policy, err := configuredPolicy(command.Context(), environment, options)
		if err != nil {
			return err
		}
		return jsonOutput(environment, struct {
			File   string        `json:"file"`
			Policy config.Policy `json:"policy"`
		}{File: file, Policy: policy})
	}}
	explain.Flags().StringVar(&file, "file", "", "Source path for resolving the effective file policy")
	parent.AddCommand(initialize, validate, explain)
	return parent
}

func rulesCommand(environment Environment) *cobra.Command {
	var options checkOptions
	parent := &cobra.Command{Use: "rules", Short: "Inspect and test builtin and declarative rules"}
	parent.PersistentFlags().StringVar(&options.config, "config", "", "Exact configuration path")
	parent.PersistentFlags().StringArrayVar(&options.ruleSets, "ruleset", nil, "Local ruleset file or directory")
	configurationFlags(parent, &options, true)
	parent.AddCommand(ruleListCommand(environment, &options), ruleShowCommand(environment, &options), ruleTestCommand(environment, &options))
	return parent
}

func doctorCommand(environment Environment) *cobra.Command {
	var options checkOptions
	command := &cobra.Command{Use: "doctor", Args: cobra.NoArgs, RunE: func(command *cobra.Command, _ []string) error {
		policy, err := configuredPolicy(command.Context(), environment, options)
		if err != nil {
			return err
		}
		provider, err := english.New()
		if err != nil {
			return err
		}
		return jsonOutput(
			environment,
			map[string]any{"version": unswell.Version, "schema_version": unswell.SchemaVersion, "nlp": provider.Identity(),
				"rule_count": len(policy.Rules), "rule_sets": policy.RuleSets, "config_hash": policy.Hash,
				"config_sources": policy.Sources, "config_overrides": policy.Overrides,
				"probability_status": "calibration_unavailable", "calibration_model": nil},
		)
	}}
	command.Flags().StringVar(&options.config, "config", "", "Exact configuration path")
	command.Flags().StringArrayVar(&options.ruleSets, "ruleset", nil, "Local ruleset file or directory")
	configurationFlags(command, &options, false)
	return command
}

func reportCommand(environment Environment) *cobra.Command {
	var format, output string
	command := &cobra.Command{
		Use:   "report <result.json>",
		Short: "Transform a saved result without rereading its sources",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			path := args[0]
			if !filepath.IsAbs(path) {
				path = filepath.Join(environment.Dir, path)
			}
			data, err := readLimited(path, 128<<20)
			if err != nil {
				return err
			}
			result, err := report.Read(strings.NewReader(string(data)))
			if err != nil {
				return err
			}
			return writeReports(environment, checkOptions{reports: []string{format + ":" + output}}, result, []string{path})
		},
	}
	command.Flags().StringVar(&format, "format", "html", "Output format")
	command.Flags().StringVar(&output, "output", "-", "Output path or - for stdout")
	return command
}

func explainCommand(environment Environment) *cobra.Command {
	var at string
	var options checkOptions
	command := &cobra.Command{
		Use:   "explain <path>",
		Short: "Explain the local score at a source line and column",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return explainAt(command.Context(), environment, args, at, options)
		},
	}
	command.Flags().StringVar(&at, "at", "1:1", "One-based line:Unicode-column")
	command.Flags().StringVar(&options.config, "config", "", "Exact configuration path")
	command.Flags().StringArrayVar(&options.ruleSets, "ruleset", nil, "Local ruleset file or directory")
	configurationFlags(command, &options, false)
	return command
}

func configurationFlags(command *cobra.Command, options *checkOptions, persistent bool) {
	flags := command.Flags()
	if persistent {
		flags = command.PersistentFlags()
	}
	flags.StringVar(&options.projectRoot, "project-root", "", "Configuration and source root (default: nearest Git root or working directory)")
	flags.BoolVar(&options.allowOutsideConfig, "allow-config-outside-root", false,
		"Explicitly permit local config resources outside the project root")
}

func parsePosition(text string) (int, int, error) {
	left, right, ok := strings.Cut(text, ":")
	if !ok {
		return 0, 0, fmt.Errorf("position must use line:column")
	}
	line, err := strconv.Atoi(left)
	if err != nil {
		return 0, 0, err
	}
	column, err := strconv.Atoi(right)
	if err != nil {
		return 0, 0, err
	}
	if line < 1 || column < 1 {
		return 0, 0, fmt.Errorf("position must be one-based")
	}
	return line, column, nil
}

func initializeConfig(environment Environment, path, profile string) error {
	data := []byte("version: 1\nextends: [builtin:" + profile + "-v1]\nlanguage: en\ncalibration:\n  model: none\n")
	if _, err := config.Load(data, catalog()); err != nil {
		return err
	}
	target := path
	if target == "" {
		target = ".unswell.yaml"
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(environment.Dir, target)
	}
	// #nosec G304 -- The destination is explicitly selected; O_EXCL prevents overwriting an existing configuration.
	handle, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := handle.Write(data)
	closeErr := handle.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func runRuleExamples(ctx context.Context, environment Environment, registry []rule.Rule) error {
	count := 0
	for _, implementation := range registry {
		for index, example := range implementation.Descriptor().Examples {
			if err := runRuleExample(ctx, implementation, example); err != nil {
				return fmt.Errorf("%s example %d: %w", implementation.Descriptor().ID, index+1, err)
			}
			count++
		}
	}
	if count == 0 {
		return fmt.Errorf("catalog has no executable examples")
	}
	_, err := fmt.Fprintf(environment.Out, "Passed %d catalog examples.\n", count)
	return err
}

func runRuleExample(ctx context.Context, implementation rule.Rule, example rule.Example) error {
	engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{implementation}, Config: []byte(example.Config)})
	if err != nil {
		return err
	}
	format := example.Format
	if format == "" {
		format = document.Plain
	}
	result, err := engine.Analyze(ctx, document.Source{Name: "example.txt", Format: format, Bytes: []byte(example.Text)})
	if err != nil {
		return err
	}
	if (len(result.Findings) > 0) != example.Match {
		return fmt.Errorf("expected match=%t, got %d findings", example.Match, len(result.Findings))
	}
	return nil
}

func explainAt(ctx context.Context, environment Environment, args []string, at string, options checkOptions) error {
	line, column, err := parsePosition(at)
	if err != nil {
		return err
	}
	options.timeout, options.includeSource, options.allowEmpty = 30*time.Second, true, true
	result, _, err := analyze(ctx, environment, options, args)
	if err != nil {
		return err
	}
	if len(result.Documents) != 1 {
		return fmt.Errorf("explain requires exactly one source document")
	}
	assessments := make([]unswell.Assessment, 0)
	source := []byte(result.Documents[0].Source)
	for _, assessment := range result.Assessments {
		start, err := document.Locate(source, assessment.Span.Start)
		if err != nil {
			return err
		}
		end, err := document.Locate(source, assessment.Span.End)
		if err != nil {
			return err
		}
		if containsPosition(start, end, line, column) {
			assessments = append(assessments, assessment)
		}
	}
	if len(assessments) == 0 {
		return fmt.Errorf("position has no applicable prose assessment")
	}
	return jsonOutput(environment, assessments)
}

func containsPosition(start, end document.Position, line, column int) bool {
	after := line > start.Line || line == start.Line && column >= start.Column
	before := line < end.Line || line == end.Line && column < end.Column
	return after && before
}
