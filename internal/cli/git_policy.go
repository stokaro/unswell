package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/stokaro/unswell/internal/appconfig"
)

func policyInputs(loaded appconfig.Loaded, rules map[string][]byte, baselinePaths []string, baseline []byte) map[string][]byte {
	for _, path := range loaded.Paths {
		name, _ := filepath.Rel(loaded.Root, path)
		rules[path] = loaded.Bundle.Files[filepath.ToSlash(name)]
	}
	for _, path := range baselinePaths {
		rules[path] = baseline
	}
	return rules
}

func (g *gitComparison) verifyPolicy(environment Environment, options checkOptions, resources map[string][]byte) error {
	for path, data := range resources {
		name, err := filepath.Rel(g.project, path)
		if err != nil || escapesRoot(name) {
			return fmt.Errorf("changed analysis requires policy inputs inside the project root: %s", path)
		}
		name = filepath.ToSlash(name)
		if err := g.unchangedPolicy(name); err != nil {
			return err
		}
		if err := g.verifyFile(name, data); err != nil {
			return err
		}
	}
	if options.config == "" && options.profile == "" {
		if err := g.verifyDiscovery(environment.Dir, resources); err != nil {
			return err
		}
	}
	return g.verifyRuleDirectories(environment, options.ruleSets, resources)
}

func (g *gitComparison) unchangedPolicy(name string) error {
	if g.before[g.prefix+name] != g.after[g.prefix+name] {
		return fmt.Errorf("policy input changed between merge base and HEAD: %s; use --policy-from-base or run a full check", name)
	}
	return nil
}

func (g *gitComparison) verifyDiscovery(dir string, resources map[string][]byte) error {
	for {
		name, err := filepath.Rel(g.project, filepath.Join(dir, ".unswell.yaml"))
		if err != nil || escapesRoot(name) {
			return fmt.Errorf("configuration discovery escapes project root")
		}
		name = filepath.ToSlash(name)
		if err := g.unchangedPolicy(name); err != nil {
			return err
		}
		if g.after[g.prefix+name].id != "" {
			return g.requirePolicyInput(name, resources)
		}
		if dir == g.project {
			return nil
		}
		dir = filepath.Dir(dir)
	}
}

func (g *gitComparison) verifyRuleDirectories(environment Environment, args []string, resources map[string][]byte) error {
	for _, arg := range args {
		path := absoluteArguments(environment.Dir, []string{arg})[0]
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			continue
		}
		name, err := filepath.Rel(g.project, path)
		if err != nil || escapesRoot(name) {
			return fmt.Errorf("ruleset directory escapes project root")
		}
		if err := g.unchangedRuleDirectory(filepath.ToSlash(name), resources); err != nil {
			return err
		}
	}
	return nil
}

func (g *gitComparison) unchangedRuleDirectory(dir string, resources map[string][]byte) error {
	for _, tree := range []map[string]gitEntry{g.before, g.after} {
		for path := range tree {
			name := strings.TrimPrefix(path, g.prefix)
			if filepath.ToSlash(filepath.Dir(name)) == dir && slices.Contains([]string{".yaml", ".yml"}, filepath.Ext(name)) {
				if err := g.unchangedPolicy(name); err != nil {
					return err
				}
				if err := g.requirePolicyInput(name, resources); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (g *gitComparison) requirePolicyInput(name string, resources map[string][]byte) error {
	if _, loaded := resources[filepath.Join(g.project, filepath.FromSlash(name))]; !loaded {
		return fmt.Errorf("missing committed policy input: %s", name)
	}
	return nil
}

func (g *gitComparison) verifySnapshot(ctx context.Context, sources committedSources, resources map[string][]byte, deleted []string) error {
	data, err := gitOutput(ctx, g.root, gitListLimit, "ls-files", "--stage", "--full-name", "-z")
	if err != nil {
		return err
	}
	g.index, err = gitEntries(data, true)
	if err != nil {
		return err
	}
	for _, source := range sources.after {
		if err := g.verifyBytes(ctx, source.Name, source.Bytes); err != nil {
			return err
		}
	}
	if err := g.verifyResources(ctx, resources); err != nil {
		return err
	}
	if err := g.verifyDeletions(deleted); err != nil {
		return err
	}
	for _, source := range sources.before {
		if g.after[g.prefix+source.Name].id == "" {
			if err := g.verifyDeleted(source.Name); err != nil {
				return err
			}
		}
	}
	return g.verifyHead(ctx)
}

func (g *gitComparison) verifyDeletions(names []string) error {
	for _, name := range names {
		if err := g.verifyDeleted(name); err != nil {
			return err
		}
	}
	return nil
}

func (g *gitComparison) verifyResources(ctx context.Context, resources map[string][]byte) error {
	for path, data := range resources {
		name, err := filepath.Rel(g.project, path)
		if err != nil {
			return err
		}
		if err := g.verifyBytes(ctx, filepath.ToSlash(name), data); err != nil {
			return err
		}
	}
	return nil
}

func (g *gitComparison) verifyBytes(ctx context.Context, name string, expected []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := regularWorktreeFile(g.project, name, len(expected))
	if err != nil {
		return err
	}
	if !bytes.Equal(data, expected) {
		return fmt.Errorf("selected input changed during analysis: %s", name)
	}
	return g.verifyFile(name, data)
}

func (g *gitComparison) verifyDeleted(name string) error {
	if g.index[g.prefix+name].id != "" {
		return fmt.Errorf("deleted file was restored in the index: %s", name)
	}
	_, err := os.Lstat(filepath.Join(g.project, filepath.FromSlash(name)))
	if !os.IsNotExist(err) {
		return fmt.Errorf("deleted file is present or unreadable in the worktree: %s", name)
	}
	return nil
}
