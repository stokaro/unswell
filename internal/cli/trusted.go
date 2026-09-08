package cli

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/internal/appconfig"
	"github.com/stokaro/unswell/rule"
)

type trustedInputs struct {
	git      *gitComparison
	loaded   appconfig.Loaded
	rules    []rule.Rule
	baseline []byte
	observed map[string][]byte
	deleted  []string
	changes  []unswell.PolicyChange
}

func analyzeTrusted(
	ctx context.Context, environment Environment, options checkOptions, args []string,
) (unswell.RunResult, []string, error) {
	if err := trustedFlags(options); err != nil {
		return unswell.RunResult{}, nil, err
	}
	inputs, err := loadTrustedInputs(ctx, environment, options)
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	engine, err := unswell.New(unswell.Options{Features: options.features,
		ConfigBundle: &inputs.loaded.Bundle, Rules: append(builtin.Rules(), inputs.rules...), Baseline: inputs.baseline,
		Jobs: options.jobs, IncludeSource: options.includeSource, NoGate: options.noGate, AllowEmpty: options.allowEmpty,
	})
	if err != nil {
		return unswell.RunResult{}, nil, err
	}
	sources, err := inputs.git.sources(ctx, engine, options, absoluteArguments(environment.Dir, args))
	paths := committedInputPaths(sources.paths, inputs.observed, inputs.deleted, inputs.git.project)
	if err != nil {
		return unswell.RunResult{}, paths, err
	}
	result, err := engine.AnalyzeChangedWithOptions(ctx, sources.before, sources.after, unswell.ChangeOptions{
		TrustedPolicy: true, PolicyChanges: inputs.changes,
	})
	setGitManifest(&result, inputs.git, options.changedFrom, "git-committed-trusted-policy")
	if err != nil {
		return result, paths, err
	}
	err = inputs.git.verifySnapshot(ctx, sources, inputs.observed, inputs.deleted)
	return finishGitResult(result, paths, err)
}

func trustedFlags(options checkOptions) error {
	if options.changedFrom == "" || options.stdin || options.collectBaseline {
		return fmt.Errorf("--policy-from-base requires --changed-from and committed source paths without baseline capture")
	}
	if options.profile != "" || options.gateMode != "" || options.allowOutsideConfig {
		return fmt.Errorf("trusted policy does not accept profile, gate-mode, or outside-root overrides")
	}
	return nil
}

func loadTrustedInputs(ctx context.Context, environment Environment, options checkOptions) (*trustedInputs, error) {
	root, err := appconfig.ProjectRoot(appconfig.Options{Dir: environment.Dir, Root: options.projectRoot})
	if err != nil {
		return nil, err
	}
	comparison, err := openGitComparison(ctx, root, options.changedFrom)
	if err != nil {
		return nil, err
	}
	inputs := &trustedInputs{git: comparison, observed: make(map[string][]byte)}
	if err := inputs.configuration(ctx, environment, options); err != nil {
		return nil, err
	}
	if err := inputs.ruleFiles(ctx, environment, options.ruleSets); err != nil {
		return nil, err
	}
	if err := inputs.baselineFile(ctx, environment, options.baseline); err != nil {
		return nil, err
	}
	slices.Sort(inputs.deleted)
	return inputs, nil
}
