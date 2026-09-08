package appconfig_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/internal/appconfig"
)

func TestResourceGraphUsesOnlyTheSelectedSnapshot(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	writeConfig(c, root, ".unswell.yaml", "version: 1\nextends: [builtin:strict]\n")
	files := map[string][]byte{
		"docs/policy.yaml":  []byte("version: 1\nextends: [../policy/base.yaml]\n"),
		"policy/base.yaml":  []byte("version: 1\nvocabulary: {dictionaries: [terms.yaml]}\n"),
		"policy/terms.yaml": []byte("version: 1\nterms: [control plane]\n"),
	}
	loaded, err := appconfig.LoadResources(t.Context(), root, "docs/policy.yaml", resourceMap(files))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded.Paths, qt.HasLen, 3)
	c.Assert(loaded.Bundle.Root, qt.Equals, "docs/policy.yaml")
	plan, _, err := config.CompileBundle(loaded.Bundle, nil)
	c.Assert(err, qt.IsNil)
	policy, err := plan.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(policy.Profile, qt.Equals, "technical-v1")
	c.Assert(policy.Vocabulary.ResolvedTerms, qt.DeepEquals, []string{"control plane"})
	files["policy/terms.yaml"][0] = '!'
	c.Assert(string(loaded.Bundle.Files["policy/terms.yaml"]), qt.Equals, "version: 1\nterms: [control plane]\n")
	_, err = os.Stat(filepath.Join(root, "docs", "policy.yaml"))
	c.Assert(os.IsNotExist(err), qt.IsTrue)
}

func resourceMap(files map[string][]byte) func(string) ([]byte, error) {
	return func(name string) ([]byte, error) {
		data, ok := files[name]
		if !ok {
			return nil, fmt.Errorf("missing snapshot resource: %s", name)
		}
		return data, nil
	}
}

func TestResourceGraphRejectsMissingCyclicAndOutsideReferences(t *testing.T) {
	for _, source := range []string{
		"version: 1\nextends: [missing.yaml]\n",
		"version: 1\nextends: [policy.yaml]\n",
		"version: 1\nextends: [../outside.yaml]\n",
		"version: 1\nextends: ['https://example.com/policy.yaml']\n",
		strings.Repeat(" ", (1<<20)+1),
	} {
		c := qt.New(t)
		_, err := appconfig.LoadResources(t.Context(), t.TempDir(), "policy.yaml", resourceMap(map[string][]byte{
			"policy.yaml": []byte(source),
		}))
		c.Assert(err, qt.IsNotNil)
	}
}

func TestResourceGraphDefaultAndCancellation(t *testing.T) {
	c := qt.New(t)
	loaded, err := appconfig.LoadResources(t.Context(), t.TempDir(), "", nil)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded.Paths, qt.HasLen, 0)
	_, _, err = config.CompileBundle(loaded.Bundle, nil)
	c.Assert(err, qt.IsNil)
	_, err = appconfig.LoadResources(t.Context(), t.TempDir(), "policy.yaml", nil)
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	_, err = appconfig.LoadResources(ctx, t.TempDir(), "policy.yaml", func(_ string) ([]byte, error) {
		cancel()
		return []byte("version: 1\n"), nil
	})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
