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
	"github.com/stokaro/unswell/builtin"
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
	var path, profile, file string
	parent.PersistentFlags().StringVar(&path, "config", "", "Exact configuration path")
	initialize := &cobra.Command{Use: "init", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		return initializeConfig(environment, path, profile)
	}}
	initialize.Flags().StringVar(&profile, "profile", "technical", "Builtin profile")
	validate := &cobra.Command{Use: "validate", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		data, err := configuration(environment, checkOptions{config: path})
		if err != nil {
			return err
		}
		policy, err := config.Load(data, catalog())
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(environment.Out, "Configuration valid:", policy.Profile, policy.Hash)
		return err
	}}
	explain := &cobra.Command{Use: "explain", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		data, err := configuration(environment, checkOptions{config: path})
		if err != nil {
			return err
		}
		policy, err := config.Load(data, catalog())
		if err != nil {
			return err
		}
		return jsonOutput(environment, struct {
			File   string        `json:"file"`
			Policy config.Policy `json:"policy"`
		}{File: file, Policy: policy})
	}}
	explain.Flags().StringVar(&file, "file", "", "Logical filename to identify in the policy explanation")
	parent.AddCommand(initialize, validate, explain)
	return parent
}

func rulesCommand(environment Environment) *cobra.Command {
	parent := &cobra.Command{Use: "rules", Short: "Inspect and test the builtin rule catalog"}
	list := &cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		for _, descriptor := range catalog() {
			if _, err := fmt.Fprintf(environment.Out, "%s\t%s\t%s\n", descriptor.ID, descriptor.Status, descriptor.Summary); err != nil {
				return err
			}
		}
		return nil
	}}
	show := &cobra.Command{Use: "show <rule-id>", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		for _, descriptor := range catalog() {
			if descriptor.ID == args[0] {
				return jsonOutput(environment, descriptor)
			}
		}
		return fmt.Errorf("unknown rule ID %q", args[0])
	}}
	test := &cobra.Command{
		Use:   "test",
		Short: "Execute builtin catalog examples (custom DSL is planned for stage 2)",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runCatalog(command.Context(), environment)
		},
	}
	parent.AddCommand(list, show, test)
	return parent
}

func doctorCommand(environment Environment) *cobra.Command {
	return &cobra.Command{Use: "doctor", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		provider, err := english.New()
		if err != nil {
			return err
		}
		return jsonOutput(
			environment,
			map[string]any{"version": unswell.Version, "schema_version": unswell.SchemaVersion, "nlp": provider.Identity(),
				"rule_count": len(catalog()), "probability_status": "calibration_unavailable", "calibration_model": nil},
		)
	}}
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
	var at, configuration string
	command := &cobra.Command{
		Use:   "explain <path>",
		Short: "Explain the local score at a source line and column",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return explainAt(command.Context(), environment, args, at, configuration)
		},
	}
	command.Flags().StringVar(&at, "at", "1:1", "One-based line:Unicode-column")
	command.Flags().StringVar(&configuration, "config", "", "Exact configuration path")
	return command
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

func runCatalog(ctx context.Context, environment Environment) error {
	count := 0
	for _, implementation := range builtin.Rules() {
		for _, example := range implementation.Descriptor().Examples {
			engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{implementation}, Config: []byte(example.Config)})
			if err != nil {
				return err
			}
			result, err := engine.Analyze(
				ctx,
				document.Source{Name: "example.txt", Format: document.Plain, Bytes: []byte(example.Text)},
			)
			if err != nil {
				return err
			}
			if (len(result.Findings) > 0) != example.Match {
				return fmt.Errorf("catalog fixture failed for %s", implementation.Descriptor().ID)
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

func explainAt(ctx context.Context, environment Environment, args []string, at, configuration string) error {
	line, column, err := parsePosition(at)
	if err != nil {
		return err
	}
	result, _, err := analyze(
		ctx,
		environment,
		checkOptions{config: configuration, timeout: 30 * time.Second, includeSource: true, allowEmpty: true},
		args,
	)
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
