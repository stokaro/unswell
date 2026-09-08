package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
)

type committedSources struct {
	before, after []document.Source
	paths         []string
}

func (g *gitComparison) selectNames(args []string, files config.Files) ([]string, error) {
	include, err := compileGlobs(files.Include)
	if err != nil {
		return nil, err
	}
	exclude, err := compileGlobs(files.Exclude)
	if err != nil {
		return nil, err
	}
	candidates := g.candidateNames()
	selected := make(map[string]bool)
	for _, arg := range args {
		name, err := filepath.Rel(g.project, arg)
		if err != nil || escapesRoot(name) {
			return nil, fmt.Errorf("changed source selection escapes project root: %s", arg)
		}
		if err := selectCommittedName(filepath.ToSlash(name), candidates, selected, include, exclude); err != nil {
			return nil, err
		}
	}
	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	slices.Sort(names)
	return names, nil
}

func (g *gitComparison) candidateNames() map[string]bool {
	candidates := make(map[string]bool, len(g.before)+len(g.after))
	for _, tree := range []map[string]gitEntry{g.before, g.after, g.index} {
		for path := range tree {
			if name, ok := strings.CutPrefix(path, g.prefix); ok {
				candidates[name] = true
			}
		}
	}
	return candidates
}

func selectCommittedName(name string, candidates, selected map[string]bool, include, exclude []*regexp.Regexp) error {
	if candidates[name] {
		selected[name] = true
		return nil
	}
	found := false
	for candidate := range candidates {
		if name != "." && !strings.HasPrefix(candidate, name+"/") {
			continue
		}
		found = true
		if eligiblePath(candidate, include, exclude) {
			selected[candidate] = true
		}
	}
	if !found && name != "." {
		return fmt.Errorf("source path is absent from both committed trees: %s", name)
	}
	return nil
}

func (g *gitComparison) sources(
	ctx context.Context, engine *unswell.Engine, options checkOptions, args []string,
) (committedSources, error) {
	policy, err := engine.PolicyForFile("")
	if err != nil {
		return committedSources{}, err
	}
	names, err := g.selectNames(args, policy.Files)
	if err != nil {
		return committedSources{}, err
	}
	var result committedSources
	beforeBytes, afterBytes := 0, 0
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		policy, err := engine.PolicyForFile(name)
		if err != nil {
			return result, err
		}
		before, after, err := g.sourcePair(ctx, name, options.format, policy.Analysis.MaxFileBytes)
		if err != nil {
			return result, fmt.Errorf("%s: %w", name, err)
		}
		beforeBytes += len(before.Bytes)
		afterBytes += len(after.Bytes)
		if beforeBytes > policy.Analysis.MaxTotalBytes || afterBytes > policy.Analysis.MaxTotalBytes {
			return result, fmt.Errorf("committed selection exceeds max_total_bytes")
		}
		result.appendPair(before, after)
		result.paths = append(result.paths, filepath.Join(g.project, filepath.FromSlash(name)))
	}
	return result, nil
}

func (s *committedSources) appendPair(before, after document.Source) {
	if before.Name != "" {
		s.before = append(s.before, before)
	}
	if after.Name != "" {
		s.after = append(s.after, after)
	}
}

func (g *gitComparison) sourcePair(ctx context.Context, name, explicit string, limit int) (document.Source, document.Source, error) {
	old, current := g.before[g.prefix+name], g.after[g.prefix+name]
	var before, after document.Source
	if old.id == "" && current.id == "" {
		return before, after, fmt.Errorf("selected index path is absent from both committed trees")
	}
	var err error
	after, err = g.currentSource(name, explicit, limit)
	if err != nil {
		return before, after, err
	}
	if old.id == "" {
		return before, after, nil
	}
	if old == current {
		return after, after, nil
	}
	data, err := g.blob(ctx, old, limit)
	if err != nil {
		return before, after, err
	}
	before, err = committedSource(name, explicit, data)
	return before, after, err
}

func (g *gitComparison) currentSource(name, explicit string, limit int) (document.Source, error) {
	if g.after[g.prefix+name].id == "" {
		return document.Source{}, g.verifyDeleted(name)
	}
	data, err := regularWorktreeFile(g.project, name, limit)
	if err != nil {
		return document.Source{}, err
	}
	if err := g.verifyFile(name, data); err != nil {
		return document.Source{}, err
	}
	return committedSource(name, explicit, data)
}

func committedSource(name, explicit string, data []byte) (document.Source, error) {
	format, err := inputFormat(name, explicit, data)
	return document.Source{Name: name, Format: format, Bytes: data}, err
}
