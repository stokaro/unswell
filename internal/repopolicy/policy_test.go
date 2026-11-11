package repopolicy_test

import (
	"testing"
	"testing/fstest"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/repopolicy"
)

func fixture() fstest.MapFS {
	return fstest.MapFS{
		".gomodules": {Data: []byte(". runtime\ntools tools\n")},
		".github/dependabot.yml": {Data: []byte("version: 2\nupdates:\n" +
			"- package-ecosystem: gomod\n  directories: [/, /tools]\n")},
		"go.mod":             {Data: []byte("module github.com/stokaro/unswell\n")},
		"tools/go.mod":       {Data: []byte("module example.com/tools\n")},
		"docs/public_api.md": {Data: []byte("`github.com/stokaro/unswell`\n")},
		"engine.go":          {Data: []byte("package unswell\nimport \"context\"\n")},
		".github/workflows/ci.yml": {Data: []byte("uses: actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09\n" +
			"os: [ubuntu-latest, macos-latest, windows-latest]\nGOTOOLCHAIN: local\ngo-version-file: go.mod\n" +
			"run: scripts/modules.sh test\nrun: make check\n")},
	}
}

func TestRepositoryPolicy(t *testing.T) {
	cases := []struct{ name, file, content, want string }{
		{"unlisted module", "forgotten/go.mod", "module example.com/forgotten", "unclassified Go module.*"},
		{"missing module", ".gomodules", ". runtime\nmissing consumer\n", ".*does not exist"},
		{"duplicate module", ".gomodules", ". runtime\n. consumer\n", "invalid or duplicate.*"},
		{"bad role", ".gomodules", ". skipped\n", "invalid module inventory.*"},
		{"root skipped", ".gomodules", ". tools\ntools tools\n", "root module must be runtime"},
		{"missing root", ".gomodules", "tools tools\n", "root module is missing"},
		{"unlisted API", "accidental/api.go", "package accidental", "public package absent.*"},
		{"filesystem", "engine.go", "package unswell; import \"os\"", "library boundary.*"},
		{"network", "engine.go", "package unswell; import \"net/http\"", "library boundary.*"},
		{"CLI dependency", "engine.go", "package unswell; import \"github.com/stokaro/unswell/internal/cli\"", "library boundary.*"},
		{"config loader", "engine.go", "package unswell; import \"github.com/stokaro/unswell/internal/appconfig\"", "library boundary.*"},
		{"unpinned CI", ".github/workflows/ci.yml", "- uses: actions/checkout@main", "unpinned action.*"},
		{"missing CI jobs", ".github/workflows/ci.yml", "", "missing CI coverage.*"},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			c.Assert(repopolicy.Check(tree), qt.IsNil)
			tree[row.file] = &fstest.MapFile{Data: []byte(row.content)}
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, row.want)
		})
	}
}

// A records module bounds the root module and holds no Go source of its own.
func TestRecordsModuleBoundsTheRootModule(t *testing.T) {
	recordsFixture := func() fstest.MapFS {
		tree := fixture()
		tree[".gomodules"] = &fstest.MapFile{Data: []byte(". runtime\nresearch records\nresearch/annotation consumer\ntools tools\n")}
		tree[".github/dependabot.yml"] = &fstest.MapFile{Data: []byte("version: 2\nupdates:\n" +
			"- package-ecosystem: gomod\n  directories: [/, /research, /research/annotation, /tools]\n")}
		tree["research/go.mod"] = &fstest.MapFile{Data: []byte("module github.com/stokaro/unswell/research\n")}
		tree["research/annotation/go.mod"] = &fstest.MapFile{Data: []byte("module github.com/stokaro/unswell/research/annotation\n")}
		tree["research/annotation/tool.go"] = &fstest.MapFile{Data: []byte("package annotation\n")}
		tree["research/reviews/record.json"] = &fstest.MapFile{Data: []byte("{}\n")}
		return tree
	}
	for _, row := range []struct{ name, file, content, want string }{
		{"source in records", "research/reviews/probe.go", "package reviews", "records module must not hold Go source.*"},
		{"records elsewhere", ".gomodules", ". runtime\nresearch records\nresearch/annotation records\ntools tools\n",
			"only the research directory can have the records role"},
		{"research not records", ".gomodules", ". runtime\nresearch consumer\nresearch/annotation consumer\ntools tools\n",
			"only the research directory can have the records role"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			tree := recordsFixture()
			c.Assert(repopolicy.Check(tree), qt.IsNil)
			tree[row.file] = &fstest.MapFile{Data: []byte(row.content)}
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, row.want)
		})
	}
}

func TestAdapterHasThePublicLibraryBoundary(t *testing.T) {
	for _, row := range []struct{ name, file, source, want string }{
		{"filesystem", "goanalysis/analyzer.go", "package goanalysis; import \"os\"", "library boundary.*"},
		{"internal", "goanalysis/analyzer.go", "package goanalysis; import \"github.com/stokaro/unswell/internal/mapping\"",
			"go/analysis adapter must use public.*"},
		{"driver internal", "goanalysis/cmd/driver/main.go", "package main; import \"github.com/stokaro/unswell/internal/cli\"",
			"go/analysis adapter must use public.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			tree[".gomodules"].Data = []byte(". runtime\ngoanalysis runtime\ntools tools\n")
			tree["goanalysis/go.mod"] = &fstest.MapFile{Data: []byte("module github.com/stokaro/unswell/goanalysis\n")}
			tree[".github/dependabot.yml"].Data = append(tree[".github/dependabot.yml"].Data,
				[]byte("- package-ecosystem: gomod\n  directory: /goanalysis\n")...)
			tree["docs/public_api.md"].Data = []byte("`github.com/stokaro/unswell`\n`github.com/stokaro/unswell/goanalysis`\n")
			tree["goanalysis/analyzer.go"] = &fstest.MapFile{Data: []byte("package goanalysis; import \"context\"")}
			c.Assert(repopolicy.Check(tree), qt.IsNil)
			tree[row.file] = &fstest.MapFile{Data: []byte(row.source)}
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, row.want)
		})
	}
}

// A tool directory can hold a whole second checkout of this repository. An
// agent worktree under .claude does. Walking into one reported every package
// of the copy as missing from the API ledger, so the check failed over a copy
// rather than over the repository's own source.
func TestToolDirectoriesAreNotWalked(t *testing.T) {
	for _, name := range []string{
		".claude/worktrees/agent-1/accidental/api.go",
		".idea/scratch/api.go",
		".cache/copy/api.go",
	} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			tree[name] = &fstest.MapFile{Data: []byte("package accidental")}
			c.Assert(repopolicy.Check(tree), qt.IsNil)
		})
	}
}

// The workflow pins live under .github, so that one dot directory is read.
func TestGithubDirectoryIsStillWalked(t *testing.T) {
	c := qt.New(t)
	tree := fixture()
	tree[".github/workflows/ci.yml"] = &fstest.MapFile{Data: []byte("- uses: actions/checkout@main")}
	c.Assert(repopolicy.Check(tree), qt.ErrorMatches, "unpinned action.*")
}
