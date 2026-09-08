package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/stokaro/unswell"
)

func analyzeCommitted(
	ctx context.Context, environment Environment, options checkOptions, engine *unswell.Engine,
	root string, args []string, resources map[string][]byte,
) (unswell.RunResult, []string, error) {
	if options.stdin || options.collectBaseline {
		return unswell.RunResult{}, nil, fmt.Errorf("--changed-from requires committed source paths and cannot capture a baseline")
	}
	comparison, err := openGitComparison(ctx, root, options.changedFrom)
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	if err := comparison.verifyPolicy(environment, options, resources); err != nil {
		return unswell.RunResult{}, nil, err
	}
	sources, err := comparison.sources(ctx, engine, options, absoluteArguments(environment.Dir, args))
	inputs := slices.Clone(sources.paths)
	for path := range resources {
		inputs = append(inputs, path)
	}
	slices.Sort(inputs)
	if err != nil {
		return unswell.RunResult{}, inputs, err
	}
	result, err := engine.AnalyzeChanged(ctx, sources.before, sources.after)
	setGitManifest(&result, comparison, options.changedFrom, "git-committed-changed-units")
	if err != nil {
		return result, inputs, err
	}
	err = comparison.verifySnapshot(ctx, sources, resources, nil)
	return finishGitResult(result, inputs, err)
}

func setGitManifest(result *unswell.RunResult, comparison *gitComparison, ref, mode string) {
	result.Manifest.SelectionMode = mode
	result.Manifest.Git = &unswell.GitSelection{RequestedRef: ref, BaseCommit: comparison.base, HeadCommit: comparison.head}
}

func finishGitResult(result unswell.RunResult, inputs []string, err error) (unswell.RunResult, []string, error) {
	if err != nil {
		result.Status = "incomplete"
		result.Manifest.Complete, result.Changes.Complete, result.Gate.Passed = false, false, false
		if result.PolicyComparison != nil {
			result.PolicyComparison.Complete = false
		}
		result.Errors = append(result.Errors, unswell.RunError{Message: err.Error()})
		return result, inputs, err
	}
	result.Manifest.Git.Clean = true
	return result, inputs, nil
}

func committedInputPaths(sources []string, resources map[string][]byte, deleted []string, root string) []string {
	paths := slices.Clone(sources)
	for path := range resources {
		paths = append(paths, path)
	}
	for _, name := range deleted {
		paths = append(paths, filepath.Join(root, filepath.FromSlash(name)))
	}
	slices.Sort(paths)
	return slices.Compact(paths)
}
