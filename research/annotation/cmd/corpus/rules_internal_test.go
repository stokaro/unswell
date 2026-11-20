package main

// White-box tests: Probe the command private rule-config byte limit and regular-file reader;
// the command exposes neither the limit constant nor a callable reader API.

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestRuleConfigReadsOnlyBoundedRegularFiles(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	path := filepath.Join(root, "rules.yaml")
	data := []byte("version: 1\nextends: [builtin:custom]\n")
	c.Assert(os.WriteFile(path, data, 0o600), qt.IsNil)
	read, err := readRuleConfig(path)
	c.Assert(err, qt.IsNil)
	c.Assert(read, qt.DeepEquals, data)
	for _, invalid := range []string{root, filepath.Join(root, "missing")} {
		_, err := readRuleConfig(invalid)
		c.Assert(err, qt.IsNotNil)
	}
	c.Assert(os.WriteFile(path, make([]byte, maxRuleConfigBytes+1), 0o600), qt.IsNil)
	_, err = readRuleConfig(path)
	c.Assert(err, qt.ErrorMatches, ".*byte limit.*")
}

// A repository may publish its LICENSE as a link into its own tree, as
// FasterXML/jackson-databind did at the revision that the historical run
// acquired. The root confines where such a link may point, so the reader
// follows one inside the root and refuses one that leaves it.
func TestRuleConfigFollowsLinksInsideItsRoot(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	target := filepath.Join(root, "actual.yaml")
	data := []byte("version: 1\nextends: [builtin:custom]\n")
	c.Assert(os.WriteFile(target, data, 0o600), qt.IsNil)

	inside := filepath.Join(root, "inside.yaml")
	c.Assert(os.Symlink("actual.yaml", inside), qt.IsNil)
	read, err := readRuleConfig(inside)
	c.Assert(err, qt.IsNil)
	c.Assert(read, qt.DeepEquals, data)

	outside := filepath.Join(root, "outside.yaml")
	escape := filepath.Join(t.TempDir(), "escape.yaml")
	c.Assert(os.WriteFile(escape, data, 0o600), qt.IsNil)
	c.Assert(os.Symlink(escape, outside), qt.IsNil)
	_, err = readRuleConfig(outside)
	c.Assert(err, qt.IsNotNil)

	dangling := filepath.Join(root, "dangling.yaml")
	c.Assert(os.Symlink("missing.yaml", dangling), qt.IsNil)
	_, err = readRuleConfig(dangling)
	c.Assert(err, qt.IsNotNil)
}
