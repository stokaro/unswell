package cli

import (
	"context"
	"fmt"
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
	result.Manifest.SelectionMode = "git-committed-changed-units"
	result.Manifest.Git = &unswell.GitSelection{
		RequestedRef: options.changedFrom, BaseCommit: comparison.base, HeadCommit: comparison.head,
	}
	if err != nil {
		return result, inputs, err
	}
	if err := comparison.verifySnapshot(ctx, sources, resources); err != nil {
		result.Status = "incomplete"
		result.Manifest.Complete, result.Changes.Complete, result.Gate.Passed = false, false, false
		result.Errors = append(result.Errors, unswell.RunError{Message: err.Error()})
		return result, inputs, err
	}
	result.Manifest.Git.Clean = true
	return result, inputs, nil
}
