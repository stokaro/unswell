package cli

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
)

func (s *trustedInputs) readBase(ctx context.Context, name string, limit int) ([]byte, error) {
	entry := s.git.before[s.git.prefix+name]
	if entry.id == "" {
		return nil, fmt.Errorf("trusted resource is absent from merge base: %s", name)
	}
	return s.git.blob(ctx, entry, limit)
}

func (s *trustedInputs) observe(ctx context.Context, name string, limit int) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path := filepath.Join(s.git.project, filepath.FromSlash(name))
	if data, exists := s.observed[path]; exists {
		return data, nil
	}
	if s.git.after[s.git.prefix+name].id == "" {
		if err := s.git.verifyDeleted(name); err != nil {
			return nil, err
		}
		if !slices.Contains(s.deleted, name) {
			s.deleted = append(s.deleted, name)
		}
		return nil, nil
	}
	data, err := regularWorktreeFile(s.git.project, name, limit)
	if err != nil {
		return nil, err
	}
	if err := s.git.verifyFile(name, data); err != nil {
		return nil, err
	}
	s.observed[path] = data
	return data, nil
}

func (s *trustedInputs) auditResource(ctx context.Context, name, kind string, before []byte, limit int) error {
	after, err := s.observe(ctx, name, limit)
	if err != nil {
		return err
	}
	afterHash := ""
	if s.git.after[s.git.prefix+name].id != "" {
		afterHash = resourceHash(after)
	}
	s.addChange(name, kind, resourceHash(before), afterHash)
	return nil
}

func (s *trustedInputs) addChange(name, kind, before, after string) {
	if before == after {
		return
	}
	if slices.ContainsFunc(s.changes, func(change unswell.PolicyChange) bool { return change.Path == name && change.Kind == kind }) {
		return
	}
	s.changes = append(s.changes, unswell.PolicyChange{Path: name, Kind: kind, BeforeHash: before, AfterHash: after})
}

func resourceHash(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (g *gitComparison) resourceName(dir, arg string) (string, error) {
	path := absoluteArguments(dir, []string{arg})[0]
	name, err := filepath.Rel(g.project, path)
	if err != nil || escapesRoot(name) {
		return "", fmt.Errorf("trusted resource must be inside the project root: %s", arg)
	}
	return filepath.ToSlash(name), nil
}

func (s *trustedInputs) baselineFile(ctx context.Context, environment Environment, arg string) error {
	if arg == "" {
		return nil
	}
	name, err := s.git.resourceName(environment.Dir, arg)
	if err != nil {
		return err
	}
	data, err := s.readBase(ctx, name, baseline.MaxBytes)
	if err != nil {
		return err
	}
	s.baseline = data
	return s.auditResource(ctx, name, "baseline", data, baseline.MaxBytes)
}
