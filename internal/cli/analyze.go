package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/internal/appconfig"
	"github.com/stokaro/unswell/rule"
)

func catalog() []rule.Descriptor {
	result := make([]rule.Descriptor, 0)
	for _, implementation := range builtin.Rules() {
		result = append(result, implementation.Descriptor())
	}
	return result
}

func configuration(ctx context.Context, environment Environment, options checkOptions) (appconfig.Loaded, error) {
	load := appconfig.Options{Dir: environment.Dir, Root: options.projectRoot, Path: options.config,
		Discover: true, AllowOutsideRoot: options.allowOutsideConfig}
	if options.profile != "" {
		if options.config != "" {
			return appconfig.Loaded{}, fmt.Errorf("--profile and --config cannot be combined")
		}
		load.Inline = []byte("version: 1\nextends: [builtin:" + options.profile + "]\n")
	}
	return appconfig.Load(ctx, load)
}

func readLimited(path string, limit int) ([]byte, error) {
	// #nosec G304 -- The CLI reads explicitly selected source, config, or saved-result paths under a byte limit.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if len(data) > limit {
		return nil, fmt.Errorf("%s exceeds byte limit %d", path, limit)
	}
	return data, nil
}

func analyze(ctx context.Context, environment Environment, options checkOptions, args []string) (unswell.RunResult, []string, error) {
	if err := validateCheckOptions(options); err != nil {
		return unswell.RunResult{}, nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()
	return analyzeInputs(ctx, environment, options, args)
}

func validateCheckOptions(options checkOptions) error {
	if !slices.Contains([]string{"note", "warning", "error"}, options.minSeverity) && options.minSeverity != "" {
		return fmt.Errorf("invalid minimum severity")
	}
	if options.maxFindings < 0 || options.timeout <= 0 {
		return fmt.Errorf("invalid display limit or timeout")
	}
	return nil
}

func analyzeInputs(ctx context.Context, environment Environment, options checkOptions, args []string) (unswell.RunResult, []string, error) {
	loaded, err := configuration(ctx, environment, options)
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	additional, rulePaths, ruleData, err := ruleInputs(environment, options.ruleSets)
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	registry := append(builtin.Rules(), additional...)
	baselineData, baselinePaths, err := baselineInput(environment, options)
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	engine, err := unswell.New(
		unswell.Options{
			Baseline: baselineData, CollectBaseline: options.collectBaseline, GateMode: options.gateMode,
			ConfigBundle:  &loaded.Bundle,
			Rules:         registry,
			Jobs:          options.jobs,
			IncludeSource: options.includeSource,
			NoGate:        options.noGate,
			AllowEmpty:    options.allowEmpty,
		},
	)
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	if options.changedFrom != "" {
		resources := policyInputs(loaded, ruleData, baselinePaths, baselineData)
		return analyzeCommitted(ctx, environment, options, engine, loaded.Root, args, resources)
	}
	sources, paths, mode, err := selectSources(ctx, environment, options, engine, loaded.Root, args)
	if err != nil {
		return unswell.RunResult{}, paths, err
	}
	result, err := engine.AnalyzeAll(ctx, sources)
	result.Manifest.SelectionMode = mode
	paths = append(paths, loaded.Paths...)
	paths = append(paths, baselinePaths...)
	return result, append(paths, rulePaths...), err
}

func baselineInput(environment Environment, options checkOptions) ([]byte, []string, error) {
	if options.baseline == "" || options.collectBaseline {
		return nil, nil, nil
	}
	path := absoluteArguments(environment.Dir, []string{options.baseline})[0]
	data, err := readLimited(path, baseline.MaxBytes)
	return data, []string{path}, err
}

func selectSources(
	ctx context.Context,
	environment Environment,
	options checkOptions,
	engine *unswell.Engine,
	root string,
	args []string,
) ([]document.Source, []string, string, error) {
	policy, err := engine.PolicyForFile("")
	if err != nil {
		return nil, nil, "", err
	}
	if options.stdin {
		if len(args) > 0 {
			return nil, nil, "", fmt.Errorf("stdin cannot be combined with source paths")
		}
		policy, err := engine.PolicyForFile(options.filename)
		if err != nil {
			return nil, nil, "", err
		}
		source, err := stdinSource(environment, options, policy.Analysis.MaxFileBytes)
		return []document.Source{source}, nil, "stdin", err
	}
	paths, mode, err := discover(ctx, root, absoluteArguments(environment.Dir, args), policy.Files)
	if err != nil {
		return nil, nil, "", err
	}
	sources, err := readSources(ctx, root, options, engine, paths)
	return sources, paths, mode, err
}

func absoluteArguments(dir string, args []string) []string {
	if len(args) == 0 {
		return []string{dir}
	}
	result := slices.Clone(args)
	for i, arg := range result {
		if !filepath.IsAbs(arg) {
			result[i] = filepath.Join(dir, arg)
		}
	}
	return result
}

func stdinSource(environment Environment, options checkOptions, limit int) (document.Source, error) {
	name := options.filename
	if name == "" && options.format == "" {
		return document.Source{}, fmt.Errorf("stdin requires --filename or --format")
	}
	if name == "" {
		name = "stdin"
	}
	data, err := io.ReadAll(io.LimitReader(environment.In, int64(limit)+1))
	if err != nil {
		return document.Source{}, err
	}
	format, err := inputFormat(name, options.format, data)
	return document.Source{Name: name, Format: format, Bytes: data}, err
}

func readSources(
	ctx context.Context,
	root string,
	options checkOptions,
	engine *unswell.Engine,
	paths []string,
) ([]document.Source, error) {
	sources := make([]document.Source, 0, len(paths))
	total := 0
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		policy, err := engine.PolicyForFile(filepath.ToSlash(name))
		if err != nil {
			return nil, err
		}
		data, err := readLimited(path, policy.Analysis.MaxFileBytes)
		if err != nil {
			return nil, err
		}
		format, err := inputFormat(path, options.format, data)
		if err != nil {
			return nil, err
		}
		total += len(data)
		if total > policy.Analysis.MaxTotalBytes {
			return nil, fmt.Errorf("selection exceeds max_total_bytes")
		}
		sources = append(sources, document.Source{Name: filepath.ToSlash(name), Format: format, Bytes: data})
	}
	return sources, nil
}

func inputFormat(name, explicit string, source []byte) (document.Format, error) {
	if explicit != "" {
		format := document.Format(explicit)
		if !slices.Contains(document.Formats(), format) {
			return "", fmt.Errorf("unsupported input format %q", explicit)
		}
		return format, nil
	}
	if format, ok := extract.Detect(name, source); ok {
		return format, nil
	}
	return "", fmt.Errorf("unsupported input extension for %q; specify a supported --format", name)
}
