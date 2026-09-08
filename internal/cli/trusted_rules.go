package cli

import (
	"context"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/stokaro/unswell/ruleset"
)

func (s *trustedInputs) ruleFiles(ctx context.Context, environment Environment, args []string) error {
	var names []string
	for _, arg := range args {
		name, err := s.git.resourceName(environment.Dir, arg)
		if err != nil {
			return err
		}
		selected, err := s.ruleSelection(ctx, name)
		if err != nil {
			return err
		}
		names = append(names, selected...)
	}
	if len(names) > 10 {
		return fmt.Errorf("at most 10 trusted ruleset files may be loaded")
	}
	slices.Sort(names)
	for _, name := range names {
		data, err := s.readBase(ctx, name, 1<<20)
		if err != nil {
			return err
		}
		set, err := ruleset.Load(data)
		if err != nil {
			return fmt.Errorf("trusted ruleset %s: %w", name, err)
		}
		s.rules = append(s.rules, set.Rules()...)
		if err := s.auditResource(ctx, name, "ruleset", data, 1<<20); err != nil {
			return err
		}
	}
	return nil
}

func (s *trustedInputs) ruleSelection(ctx context.Context, name string) ([]string, error) {
	if s.git.before[s.git.prefix+name].id != "" {
		return []string{name}, nil
	}
	before, err := s.git.ruleDirectory(ctx, name, s.git.before)
	if err != nil {
		return nil, err
	}
	if len(before) == 0 {
		return nil, fmt.Errorf("trusted ruleset directory has no YAML files: %s", name)
	}
	after, err := s.git.ruleDirectory(ctx, name, s.git.after)
	if err != nil {
		return nil, err
	}
	s.addChange(name, "ruleset-directory", resourceHash([]byte(strings.Join(before, "\x00"))),
		resourceHash([]byte(strings.Join(after, "\x00"))))
	return before, nil
}

func (g *gitComparison) ruleDirectory(ctx context.Context, dir string, tree map[string]gitEntry) ([]string, error) {
	var names []string
	for entry := range tree {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name, ok := strings.CutPrefix(entry, g.prefix)
		if ok && path.Dir(name) == dir && slices.Contains([]string{".yaml", ".yml"}, path.Ext(name)) {
			names = append(names, name)
			if len(names) > 4096 {
				return nil, fmt.Errorf("ruleset directory exceeds 4096 YAML entries")
			}
		}
	}
	slices.Sort(names)
	return names, nil
}
