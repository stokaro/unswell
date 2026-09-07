package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/rule"
	"github.com/stokaro/unswell/ruleset"
)

func ruleFiles(environment Environment, paths []string) ([]rule.Rule, []string, error) {
	var implementations []rule.Rule
	var files []string
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(environment.Dir, path)
		}
		selected, err := selectRuleFiles(path)
		if err != nil {
			return nil, nil, err
		}
		files = append(files, selected...)
		if len(files) > 10 {
			return nil, nil, fmt.Errorf("at most 10 ruleset files may be loaded")
		}
	}
	slices.Sort(files)
	for _, path := range files {
		data, err := readLimited(path, 1<<20)
		if err != nil {
			return nil, nil, err
		}
		set, err := ruleset.Load(data)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", path, err)
		}
		implementations = append(implementations, set.Rules()...)
	}
	return implementations, files, nil
}

func selectRuleFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("ruleset must be a regular file: %s", path)
		}
		return []string{path}, nil
	}
	// #nosec G304 -- The user explicitly selects this local directory; enumeration is bounded and nonrecursive.
	handle, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	files, readErr := directoryRuleFiles(handle, path)
	return files, errors.Join(readErr, handle.Close())
}

func directoryRuleFiles(handle *os.File, path string) ([]string, error) {
	var files []string
	for scanned := 0; ; {
		entries, err := handle.ReadDir(128)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		scanned += len(entries)
		if scanned > 4096 {
			return nil, fmt.Errorf("ruleset directory exceeds 4096 entries: %s", path)
		}
		files = append(files, yamlEntries(path, entries)...)
		if len(files) > 10 {
			return nil, fmt.Errorf("at most 10 ruleset files may be loaded")
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("ruleset directory contains no YAML files: %s", path)
	}
	return files, nil
}

func yamlEntries(path string, entries []os.DirEntry) []string {
	var files []string
	for _, entry := range entries {
		if entry.Type().IsRegular() && slices.Contains([]string{".yaml", ".yml"}, filepath.Ext(entry.Name())) {
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}
	return files
}

func descriptors(registry []rule.Rule) []rule.Descriptor {
	result := make([]rule.Descriptor, 0, len(registry))
	for _, implementation := range registry {
		result = append(result, implementation.Descriptor())
	}
	slices.SortFunc(result, func(a, b rule.Descriptor) int { return strings.Compare(a.ID, b.ID) })
	return result
}

func configuredRules(environment Environment, configurationPath string, paths []string) ([]rule.Rule, error) {
	data, err := configuration(environment, checkOptions{config: configurationPath})
	if err != nil {
		return nil, err
	}
	additional, _, err := ruleFiles(environment, paths)
	if err != nil {
		return nil, err
	}
	registry := append(builtin.Rules(), additional...)
	_, inline, err := config.Compile(data, descriptors(registry))
	if err != nil {
		return nil, err
	}
	return append(registry, inline...), nil
}

func configuredPolicy(environment Environment, options checkOptions) (config.Policy, error) {
	data, err := configuration(environment, options)
	if err != nil {
		return config.Policy{}, err
	}
	additional, _, err := ruleFiles(environment, options.ruleSets)
	if err != nil {
		return config.Policy{}, err
	}
	registry := append(builtin.Rules(), additional...)
	if _, err := unswell.New(unswell.Options{Config: data, Rules: registry}); err != nil {
		return config.Policy{}, err
	}
	return config.Load(data, descriptors(registry))
}

func ruleListCommand(environment Environment, options *checkOptions) *cobra.Command {
	return &cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
		registry, err := configuredRules(environment, options.config, options.ruleSets)
		if err != nil {
			return err
		}
		for _, descriptor := range descriptors(registry) {
			if _, err := fmt.Fprintf(environment.Out, "%s\t%s\t%s\n", descriptor.ID, descriptor.Status, descriptor.Summary); err != nil {
				return err
			}
		}
		return nil
	}}
}

func ruleShowCommand(environment Environment, options *checkOptions) *cobra.Command {
	return &cobra.Command{Use: "show <rule-id>", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		registry, err := configuredRules(environment, options.config, options.ruleSets)
		if err != nil {
			return err
		}
		for _, descriptor := range descriptors(registry) {
			if descriptor.ID == args[0] {
				return jsonOutput(environment, descriptor)
			}
		}
		return fmt.Errorf("unknown rule ID %q", args[0])
	}}
}

func ruleTestCommand(environment Environment, options *checkOptions) *cobra.Command {
	return &cobra.Command{
		Use: "test [ruleset-path]", Short: "Execute builtin or declarative pass and fail examples", Args: cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			return testSelectedRules(command.Context(), environment, *options, args)
		},
	}
}

func testSelectedRules(ctx context.Context, environment Environment, options checkOptions, args []string) error {
	if len(args) == 0 {
		registry, err := configuredRules(environment, options.config, options.ruleSets)
		if err != nil {
			return err
		}
		return runRuleExamples(ctx, environment, registry)
	}
	if options.config != "" || len(options.ruleSets) > 0 {
		return fmt.Errorf("rules test path cannot be combined with --config or --ruleset")
	}
	registry, _, err := ruleFiles(environment, args)
	if err != nil {
		return err
	}
	if _, err := unswell.New(unswell.Options{Rules: registry}); err != nil {
		return err
	}
	return runRuleExamples(ctx, environment, registry)
}
