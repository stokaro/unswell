package repopolicy_test

import (
	"strings"
	"testing"
	"testing/fstest"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/repopolicy"
)

// Each document that tells a reader which release to install has to name the
// version of the code. The fixture names alpha.21 everywhere.
func TestDocumentsNameTheVersionOfTheCode(t *testing.T) {
	for _, file := range []string{"mcp/server.json", "CHANGELOG.md", "README.md", "docs/installation.md", "docs/containers.md",
		"docs/site/src/content/docs/getting-started.mdx", "docs/site/src/content/docs/ci-action-mcp.mdx"} {
		t.Run(file, func(t *testing.T) {
			c := qt.New(t)
			tree := fixture()
			c.Assert(repopolicy.Check(tree), qt.IsNil)
			// The document still names the release before.
			stale := strings.ReplaceAll(string(tree[file].Data), "alpha.21", "alpha.20")
			tree[file] = &fstest.MapFile{Data: []byte(stale)}
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, file+` does not name release 0\.1\.0-alpha\.21: it lacks .*`)
			delete(tree, file)
			c.Assert(repopolicy.Check(tree), qt.ErrorMatches, ".*"+file+".*")
		})
	}
}

func TestTheVersionOfTheCodeMovesFirst(t *testing.T) {
	c := qt.New(t)
	// The code names the next release and no document does yet.
	tree := fixture()
	tree["result.go"] = &fstest.MapFile{Data: []byte("package unswell\n\nconst Version = \"0.1.0-alpha.22\"\n")}
	c.Assert(repopolicy.Check(tree), qt.ErrorMatches, `mcp/server\.json does not name release 0\.1\.0-alpha\.22: it lacks .*`)

	// A version that is the start of the one in the documents is another version.
	tree = fixture()
	tree["result.go"] = &fstest.MapFile{Data: []byte("package unswell\n\nconst Version = \"0.1.0-alpha.2\"\n")}
	c.Assert(repopolicy.Check(tree), qt.ErrorMatches, `mcp/server\.json does not name release 0\.1\.0-alpha\.2: it lacks .*`)

	// One place of several in a document is enough to fail.
	tree = fixture()
	page := "docs/site/src/content/docs/getting-started.mdx"
	tree[page] = &fstest.MapFile{Data: []byte(strings.Replace(string(tree[page].Data), "@v0.1.0-alpha.21", "@v0.1.0-alpha.20", 1))}
	c.Assert(repopolicy.Check(tree), qt.ErrorMatches,
		page+` does not name release 0\.1\.0-alpha\.21: it lacks "cmd/unswell@v0\.1\.0-alpha\.21"`)

	tree = fixture()
	tree["result.go"] = &fstest.MapFile{Data: []byte("package unswell\n")}
	c.Assert(repopolicy.Check(tree), qt.ErrorMatches, "result.go does not declare the release version")
	delete(tree, "result.go")
	c.Assert(repopolicy.Check(tree), qt.ErrorMatches, ".*result.go.*")
}
