package appconfig_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/internal/appconfig"
)

func writeConfig(c *qt.C, root, name, data string) {
	c.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	c.Assert(os.MkdirAll(filepath.Dir(path), 0o700), qt.IsNil)
	c.Assert(os.WriteFile(path, []byte(data), 0o600), qt.IsNil)
}

func TestNearestConfigAndContainingFileReferences(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	c.Assert(os.Mkdir(filepath.Join(root, ".git"), 0o700), qt.IsNil)
	writeConfig(c, root, ".unswell.yaml", "version: 1\nextends: [builtin:strict]\n")
	writeConfig(c, root, "docs/.unswell.yaml", "version: 1\nextends: [../policies/base.yaml]\n")
	writeConfig(c, root, "policies/base.yaml", "version: 1\nvocabulary: {dictionaries: [terms.yaml]}\n")
	writeConfig(c, root, "policies/terms.yaml", "version: 1\nterms: [control plane]\n")
	dir := filepath.Join(root, "docs", "reference")
	c.Assert(os.MkdirAll(dir, 0o700), qt.IsNil)
	loaded, err := appconfig.Load(t.Context(), appconfig.Options{Dir: dir, Discover: true})
	c.Assert(err, qt.IsNil)
	c.Assert(loaded.Root, qt.Equals, root)
	c.Assert(loaded.Bundle.Root, qt.Equals, "docs/.unswell.yaml")
	c.Assert(loaded.Paths, qt.HasLen, 3)
	plan, _, err := config.CompileBundle(loaded.Bundle, nil)
	c.Assert(err, qt.IsNil)
	policy, err := plan.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(policy.Profile, qt.Equals, "technical-v1")
	c.Assert(policy.Vocabulary.ResolvedTerms, qt.DeepEquals, []string{"control plane"})
	c.Assert(policy.Sources[1].Path, qt.Equals, "policies/base.yaml")
	startup, err := appconfig.Load(t.Context(), appconfig.Options{Dir: dir, Path: filepath.Join(root, "docs", ".unswell.yaml")})
	c.Assert(err, qt.IsNil)
	c.Assert(startup.Bundle, qt.DeepEquals, loaded.Bundle)
}

func TestRootBoundaryAndExplicitOutsidePermission(t *testing.T) {
	c := qt.New(t)
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	c.Assert(os.Mkdir(root, 0o700), qt.IsNil)
	writeConfig(c, parent, ".unswell.yaml", "version: 1\nextends: [builtin:strict]\n")
	loaded, err := appconfig.Load(t.Context(), appconfig.Options{Dir: root, Discover: true})
	c.Assert(err, qt.IsNil)
	c.Assert(loaded.Paths, qt.HasLen, 0)
	writeConfig(c, root, ".unswell.yaml", "version: 1\nextends: [../.unswell.yaml]\n")
	_, err = appconfig.Load(t.Context(), appconfig.Options{Dir: root, Discover: true})
	c.Assert(err, qt.IsNotNil)
	loaded, err = appconfig.Load(t.Context(), appconfig.Options{Dir: root, Discover: true, AllowOutsideRoot: true})
	c.Assert(err, qt.IsNil)
	c.Assert(loaded.Paths, qt.HasLen, 2)
	_, _, err = config.CompileBundle(loaded.Bundle, nil)
	c.Assert(err, qt.IsNil)
	_, err = appconfig.Load(t.Context(), appconfig.Options{Dir: root, Root: root, Path: filepath.Join(parent, ".unswell.yaml")})
	c.Assert(err, qt.IsNotNil)
}

func TestConfigSymlinkBoundariesAndAliases(t *testing.T) {
	c := qt.New(t)
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	c.Assert(os.Mkdir(root, 0o700), qt.IsNil)
	writeConfig(c, parent, "external.yaml", "version: 1\n")
	if err := os.Symlink(filepath.Join(parent, "external.yaml"), filepath.Join(root, "link.yaml")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	writeConfig(c, root, ".unswell.yaml", "version: 1\nextends: [link.yaml]\n")
	_, err := appconfig.Load(t.Context(), appconfig.Options{Dir: root, Discover: true})
	c.Assert(err, qt.IsNotNil)
	writeConfig(c, root, "base.yaml", "version: 1\n")
	c.Assert(os.Remove(filepath.Join(root, "link.yaml")), qt.IsNil)
	c.Assert(os.Symlink("base.yaml", filepath.Join(root, "link.yaml")), qt.IsNil)
	_, err = appconfig.Load(t.Context(), appconfig.Options{Dir: root, Discover: true})
	c.Assert(err, qt.IsNil)
	writeConfig(c, root, ".unswell.yaml", "version: 1\nextends: [base.yaml, link.yaml]\n")
	_, err = appconfig.Load(t.Context(), appconfig.Options{Dir: root, Discover: true})
	c.Assert(err, qt.ErrorMatches, ".*duplicate configuration file.*")
}

func TestLoaderFailuresAndCancellation(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	for _, text := range []string{
		"version: 1\nextends: [.unswell.yaml]\n",
		"version: 1\nextends: [missing.yaml]\n",
		"version: 1\nextends: ['https://example.com/config.yaml']\n",
		"version: 1\nvocabulary: {dictionaries: [directory]}\n",
	} {
		writeConfig(c, root, ".unswell.yaml", text)
		c.Assert(os.MkdirAll(filepath.Join(root, "directory"), 0o700), qt.IsNil)
		_, err := appconfig.Load(t.Context(), appconfig.Options{Dir: root, Discover: true})
		c.Assert(err, qt.IsNotNil)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := appconfig.Load(ctx, appconfig.Options{Dir: root})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
