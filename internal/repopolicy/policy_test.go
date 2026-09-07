package repopolicy_test

import (
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
			"run: scripts/modules.sh test\nrun: make check\nrun: make race\nrun: make fuzz\n")},
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
