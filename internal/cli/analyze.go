package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func catalog() []rule.Descriptor {
	result := make([]rule.Descriptor, 0)
	for _, implementation := range builtin.Rules() {
		result = append(result, implementation.Descriptor())
	}
	return result
}

func configuration(environment Environment, options checkOptions) ([]byte, error) {
	if options.profile != "" {
		if options.config != "" {
			return nil, fmt.Errorf("--profile and --config cannot be combined")
		}
		return []byte("version: 1\nextends: [builtin:" + options.profile + "]\n"), nil
	}
	path := options.config
	if path == "" {
		path = filepath.Join(environment.Dir, ".unswell.yaml")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, nil
		} else if err != nil {
			return nil, err
		}
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(environment.Dir, path)
	}
	return readLimited(path, 1<<20)
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
	if !slices.Contains([]string{"note", "warning", "error"}, options.minSeverity) && options.minSeverity != "" {
		return unswell.RunResult{}, nil, fmt.Errorf("invalid minimum severity")
	}
	if options.maxFindings < 0 || options.timeout <= 0 {
		return unswell.RunResult{}, nil, fmt.Errorf("invalid display limit or timeout")
	}
	data, err := configuration(environment, options)
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	policy, err := config.Load(data, catalog())
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	engine, err := unswell.New(
		unswell.Options{
			Config:        data,
			Jobs:          options.jobs,
			IncludeSource: options.includeSource,
			NoGate:        options.noGate,
			AllowEmpty:    options.allowEmpty,
		},
	)
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()
	sources, paths, mode, err := selectSources(ctx, environment, options, policy, args)
	if err != nil {
		return unswell.RunResult{}, paths, err
	}
	result, err := engine.AnalyzeAll(ctx, sources)
	result.Manifest.SelectionMode = mode
	return result, paths, err
}

func selectSources(
	ctx context.Context,
	environment Environment,
	options checkOptions,
	policy config.Policy,
	args []string,
) ([]document.Source, []string, string, error) {
	if options.stdin {
		if len(args) > 0 {
			return nil, nil, "", fmt.Errorf("stdin cannot be combined with source paths")
		}
		source, err := stdinSource(environment, options, policy.Analysis.MaxFileBytes)
		return []document.Source{source}, nil, "stdin", err
	}
	paths, mode, err := discover(ctx, environment.Dir, args, policy.Files)
	if err != nil {
		return nil, nil, "", err
	}
	sources, err := readSources(ctx, environment, options, policy, paths)
	return sources, paths, mode, err
}

func stdinSource(environment Environment, options checkOptions, limit int) (document.Source, error) {
	name := options.filename
	if name == "" && options.format == "" {
		return document.Source{}, fmt.Errorf("stdin requires --filename or --format")
	}
	if name == "" {
		name = "stdin"
	}
	format, err := inputFormat(name, options.format)
	if err != nil {
		return document.Source{}, err
	}
	data, err := io.ReadAll(io.LimitReader(environment.In, int64(limit)+1))
	return document.Source{Name: name, Format: format, Bytes: data}, err
}

func readSources(
	ctx context.Context,
	environment Environment,
	options checkOptions,
	policy config.Policy,
	paths []string,
) ([]document.Source, error) {
	sources := make([]document.Source, 0, len(paths))
	total := 0
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		format, err := inputFormat(path, options.format)
		if err != nil {
			return nil, err
		}
		data, err := readLimited(path, policy.Analysis.MaxFileBytes)
		if err != nil {
			return nil, err
		}
		total += len(data)
		if total > policy.Analysis.MaxTotalBytes {
			return nil, fmt.Errorf("selection exceeds max_total_bytes")
		}
		name, err := filepath.Rel(environment.Dir, path)
		if err != nil {
			return nil, err
		}
		sources = append(sources, document.Source{Name: filepath.ToSlash(name), Format: format, Bytes: data})
	}
	return sources, nil
}

func inputFormat(name, explicit string) (document.Format, error) {
	if explicit != "" {
		format := document.Format(explicit)
		if format != document.Plain && format != document.Markdown && format != document.Go {
			return "", fmt.Errorf("unsupported input format %q", explicit)
		}
		return format, nil
	}
	switch filepath.Ext(name) {
	case ".md":
		return document.Markdown, nil
	case ".txt":
		return document.Plain, nil
	case ".go":
		return document.Go, nil
	}
	return "", fmt.Errorf("unsupported input extension for %q; specify a supported --format", name)
}
