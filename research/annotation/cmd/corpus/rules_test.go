package main

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
