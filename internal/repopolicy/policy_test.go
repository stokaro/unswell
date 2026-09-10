package repopolicy_test

import (
	"strings"
	"testing"
	"testing/fstest"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/repopolicy"
)

func fixture() fstest.MapFS {
	return fstest.MapFS{
		".gomodules":         {Data: []byte(". runtime\ntools tools\n")},
		"go.mod":             {Data: []byte("module github.com/stokaro/unswell\n")},
		"tools/go.mod":       {Data: []byte("module example.com/tools\n")},
		"docs/public_api.md": {Data: []byte("`github.com/stokaro/unswell`\n")},
		"engine.go":          {Data: []byte("package unswell\nimport \"context\"\n")},
		".github/workflows/ci.yml": {Data: []byte("uses: actions/checkout@fbc6f3992d24b796d5a048ff273f7fcc4a7b6c09\n" +
			"os: [ubuntu-latest, macos-latest, windows-latest]\nGOTOOLCHAIN: local\ngo-version-file: go.mod\n" +
			"run: scripts/modules.sh test\nrun: make check\nrun: make race\nrun: make fuzz\n" +
			"artifacts/coverage/\n")},
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

// An empty workflow proves the check runs; each required step also needs its own
// negative case so a single deleted line cannot pass unnoticed.
func TestCIPolicyRejectsRemovedRequiredSteps(t *testing.T) {
	for _, row := range []struct{ name, line, want string }{
		{"tests", "run: scripts/modules.sh test", "missing CI coverage: scripts/modules.sh test"},
		{"check", "run: make check", "missing CI coverage: make check"},
		{"race", "run: make race", "missing CI coverage: make race"},
		{"fuzz", "run: make fuzz", "missing CI coverage: make fuzz"},
		{"coverage evidence", "artifacts/coverage/", "missing CI coverage: artifacts/coverage/"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			c.Assert(repopolicy.Check(tree), qt.IsNil)
			workflow := string(tree[".github/workflows/ci.yml"].Data)
			reduced := strings.ReplaceAll(workflow, row.line+"\n", "")
			c.Assert(reduced, qt.Not(qt.Equals), workflow)
			tree[".github/workflows/ci.yml"] = &fstest.MapFile{Data: []byte(reduced)}
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
			tree["docs/public_api.md"].Data = []byte("`github.com/stokaro/unswell`\n`github.com/stokaro/unswell/goanalysis`\n")
			tree["goanalysis/analyzer.go"] = &fstest.MapFile{Data: []byte("package goanalysis; import \"context\"")}
			c.Assert(repopolicy.Check(tree), qt.IsNil)
			tree[row.file] = &fstest.MapFile{Data: []byte(row.source)}
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, row.want)
		})
	}
}
