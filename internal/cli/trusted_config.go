package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/internal/appconfig"
)

func (s *trustedInputs) configuration(ctx context.Context, environment Environment, options checkOptions) error {
	selected, err := s.git.configName(environment, options.config, s.git.before)
	if err != nil {
		return err
	}
	loaded, err := appconfig.LoadResources(ctx, s.git.project, selected, func(name string) ([]byte, error) {
		return s.readBase(ctx, name, 1<<20)
	})
	if err != nil {
		return err
	}
	s.loaded = loaded
	if selected != "" {
		if err := s.auditConfiguration(ctx); err != nil {
			return err
		}
	}
	if options.config != "" {
		return nil
	}
	head, err := s.git.configName(environment, "", s.git.after)
	if err != nil {
		return err
	}
	return s.auditDiscovery(ctx, selected, head)
}

func (g *gitComparison) configName(environment Environment, explicit string, tree map[string]gitEntry) (string, error) {
	if explicit != "" {
		return g.resourceName(environment.Dir, explicit)
	}
	for dir := environment.Dir; ; dir = filepath.Dir(dir) {
		name, err := filepath.Rel(g.project, filepath.Join(dir, ".unswell.yaml"))
		if err != nil || escapesRoot(name) {
			return "", fmt.Errorf("trusted configuration discovery escapes project root")
		}
		name = filepath.ToSlash(name)
		if tree[g.prefix+name].id != "" {
			return name, nil
		}
		if dir == g.project {
			return "", nil
		}
	}
}

func (s *trustedInputs) auditConfiguration(ctx context.Context) error {
	kinds, err := configurationKinds(s.loaded.Bundle)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(kinds))
	for name := range kinds {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if err := s.auditResource(ctx, name, kinds[name], s.loaded.Bundle.Files[name], 1<<20); err != nil {
			return err
		}
	}
	return nil
}

func configurationKinds(bundle config.Bundle) (map[string]string, error) {
	kinds := make(map[string]string, len(bundle.Files))
	queue := []config.Reference{{Path: bundle.Root, Kind: "config"}}
	for len(queue) > 0 {
		ref := queue[0]
		queue = queue[1:]
		if kinds[ref.Path] != "" {
			continue
		}
		kinds[ref.Path] = ref.Kind
		if ref.Kind == "dictionary" {
			continue
		}
		dependencies, err := config.References(bundle.Files[ref.Path])
		if err != nil {
			return nil, err
		}
		for _, dependency := range dependencies {
			name, err := config.ResolveReference(ref.Path, dependency.Path, false)
			if err != nil {
				return nil, err
			}
			queue = append(queue, config.Reference{Path: name, Kind: dependency.Kind})
		}
	}
	return kinds, nil
}

func (s *trustedInputs) auditDiscovery(ctx context.Context, base, head string) error {
	if base == head {
		return nil
	}
	if base != "" {
		s.addChange(base, "config-discovery", resourceHash(s.loaded.Bundle.Files[base]), "")
	}
	if head != "" {
		data, err := s.observe(ctx, head, 1<<20)
		if err != nil {
			return err
		}
		s.addChange(head, "config-discovery", "", resourceHash(data))
	}
	return nil
}
