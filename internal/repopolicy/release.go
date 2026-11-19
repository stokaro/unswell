package repopolicy

import (
	"fmt"
	"io/fs"
	"regexp"
)

// versionConstant finds the version that the code reports.
var versionConstant = regexp.MustCompile(`(?m)^const Version = "([^"\s]+)"$`)

// releaseNames lists the files that tell a reader which release to install,
// each with the texts that hold the version. "%s" stands for the version
// without a leading "v".
var releaseNames = []struct {
	file  string
	texts []string
}{
	{"mcp/server.json", []string{`"version": "%s"`, `ghcr.io/stokaro/unswell-mcp:%s`}},
	{"CHANGELOG.md", []string{"## %s"}},
	{"README.md", []string{"The published release is `v%s`"}},
	{"docs/installation.md", []string{"The published release is `v%s`", "version: %s"}},
	{"docs/containers.md", []string{"ghcr.io/stokaro/unswell:%s", "ghcr.io/stokaro/unswell-mcp:%s"}},
	{"docs/site/src/content/docs/getting-started.mdx", []string{"The published release is `v%s`", "cmd/unswell@v%s", "version: %s"}},
	{"docs/site/src/content/docs/ci-action-mcp.mdx", []string{"version: %s"}},
}

// checkReleaseVersion requires the documents that tell a reader which release
// to install to name the version of the code. A release tag takes the tree as
// it is, and the site builds the pages of a release from its tag. A document
// that still names the release before it would tell the readers of a release
// to install the earlier one.
func checkReleaseVersion(tree fs.FS) error {
	source, err := fs.ReadFile(tree, "result.go")
	if err != nil {
		return err
	}
	match := versionConstant.FindSubmatch(source)
	if match == nil {
		return fmt.Errorf("result.go does not declare the release version")
	}
	version := string(match[1])
	for _, entry := range releaseNames {
		data, err := fs.ReadFile(tree, entry.file)
		if err != nil {
			return err
		}
		for _, text := range entry.texts {
			want := fmt.Sprintf(text, version)
			// The version ends where the text does: "alpha.2" is not named
			// by a document that says "alpha.21".
			named := regexp.MustCompile(regexp.QuoteMeta(want) + `(?:[^0-9A-Za-z.+-]|\.(?:[^0-9A-Za-z]|$)|$)`)
			if !named.Match(data) {
				return fmt.Errorf("%s does not name release %s: it lacks %q", entry.file, version, want)
			}
		}
	}
	return nil
}
