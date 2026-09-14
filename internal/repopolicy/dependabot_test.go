package repopolicy_test

import (
	"testing"
	"testing/fstest"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/repopolicy"
)

func TestDependabotRequiresEnabledCoverageOfEveryModule(t *testing.T) {
	for _, module := range []string{"goanalysis", "research/annotation", "research/dependencies", ".private"} {
		t.Run(module, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			tree[".gomodules"].Data = append(tree[".gomodules"].Data, []byte(module+" consumer\n")...)
			tree[module+"/go.mod"] = &fstest.MapFile{Data: []byte("module example.com/extra\n")}
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, "missing default-branch Dependabot version updates for Go module: .*")
			tree[".github/dependabot.yml"].Data = append(tree[".github/dependabot.yml"].Data,
				[]byte("- package-ecosystem: gomod\n  directory: /"+module+"\n")...)
			c.Assert(repopolicy.Check(tree), qt.IsNil)
		})
	}
}

func TestDependabotDirectoryPolicy(t *testing.T) {
	for _, row := range []struct{ name, updates, want string }{
		{"directory list", "- package-ecosystem: gomod\n  directories: [/, /tools]\n", ""},
		{"separate entries", "- package-ecosystem: gomod\n  directory: /\n- package-ecosystem: gomod\n  directory: /tools\n", ""},
		{"missing root", "- package-ecosystem: gomod\n  directory: /tools\n", "missing default-branch Dependabot.*"},
		{"missing tools", "- package-ecosystem: gomod\n  directory: /\n", "missing default-branch Dependabot.*tools"},
		{"other ecosystem", "- package-ecosystem: npm\n  directories: [/, /tools]\n", "missing default-branch Dependabot.*"},
		{"other branch", "- package-ecosystem: gomod\n  directories: [/, /tools]\n  target-branch: release\n",
			"missing default-branch Dependabot.*"},
		{"disabled updates", "- package-ecosystem: gomod\n  directories: [/, /tools]\n  open-pull-requests-limit: 0\n",
			"missing default-branch Dependabot.*"},
		{"unknown module", "- package-ecosystem: gomod\n  directories: [/, /tools, /forgotten]\n", "dependabot gomod directory must name.*"},
		{"glob", "- package-ecosystem: gomod\n  directories: ['/**']\n", "dependabot gomod directory must name.*"},
		{"duplicate directory", "- package-ecosystem: gomod\n  directories: [/, /tools, /tools]\n", "duplicate Dependabot gomod directory.*"},
		{"duplicate entry", "- package-ecosystem: gomod\n  directories: [/, /tools]\n- package-ecosystem: gomod\n  directory: /\n",
			"duplicate Dependabot gomod directory.*"},
		{"ambiguous location", "- package-ecosystem: gomod\n  directory: /\n  directories: [/tools]\n",
			"dependabot gomod update must set exactly one.*"},
		{"no location", "- package-ecosystem: gomod\n", "dependabot gomod update must set exactly one.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			tree[".github/dependabot.yml"].Data = []byte("version: 2\nupdates:\n" + row.updates)
			err := repopolicy.Check(tree)
			if row.want == "" {
				c.Assert(err, qt.IsNil)
			} else {
				c.Assert(err, qt.ErrorMatches, row.want)
			}
		})
	}
}

func TestDependabotRejectsMissingOrInvalidConfiguration(t *testing.T) {
	for _, row := range []struct{ name, content, want string }{
		{"missing", "", ".*does not exist"},
		{"malformed", "version: [", "(?s)invalid Dependabot configuration: .*"},
		{"unsupported version", "version: 1\n", "dependabot configuration must use version 2"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			if row.content == "" {
				delete(tree, ".github/dependabot.yml")
			} else {
				tree[".github/dependabot.yml"].Data = []byte(row.content)
			}
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, row.want)
		})
	}
}
