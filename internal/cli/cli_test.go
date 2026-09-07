package cli_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/internal/cli"
)

func TestExitCodesAndCleanJSON(t *testing.T) {
	cases := []struct {
		name, text string
		args       []string
		code       int
	}{
		{"clean", "The client opens connections.", nil, 0},
		{"policy", "Certainly! The client opens connections.", nil, 1},
		{"empty", "", nil, 2},
		{"allow empty", "", []string{"--allow-empty"}, 0},
		{"no gate", "Certainly! The client opens connections.", []string{"--no-gate"}, 0},
		{"no gate error", "", []string{"--no-gate"}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			var output, stderr bytes.Buffer
			args := append([]string{"check", "--stdin", "--filename", "draft.md", "--report", "json:-"}, tc.args...)
			code := cli.Run(
				t.Context(),
				args,
				cli.Environment{Dir: t.TempDir(), In: strings.NewReader(tc.text), Out: &output, Err: &stderr},
			)
			c.Assert(code, qt.Equals, tc.code, qt.Commentf("stderr: %s", stderr.String()))
			var result unswell.RunResult
			c.Assert(json.Unmarshal(output.Bytes(), &result), qt.IsNil)
			c.Assert(result.SchemaVersion, qt.Equals, unswell.SchemaVersion)
		})
	}
}

func TestMultipleReportsAndSavedTransformation(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	path := filepath.Join(root, "draft.md")
	c.Assert(os.WriteFile(path, []byte("Certainly! It is **important** to note that the server may retry."), 0o600), qt.IsNil)
	var output, stderr bytes.Buffer
	args := []string{
		"check",
		"draft.md",
		"--include-source",
		"--report",
		"json:result.json",
		"--report",
		"sarif:result.sarif",
		"--report",
		"html:result.html",
		"--report",
		"markdown:result.md",
		"--report",
		"text:-",
	}
	environment := cli.Environment{Dir: root, In: strings.NewReader(""), Out: &output, Err: &stderr}
	c.Assert(cli.Run(t.Context(), args, environment), qt.Equals, 1, qt.Commentf("%s", stderr.String()))
	for _, name := range []string{"result.json", "result.sarif", "result.html", "result.md"} {
		info, err := os.Stat(filepath.Join(root, name))
		c.Assert(err, qt.IsNil)
		c.Assert(info.Size() > 0, qt.IsTrue)
	}
	c.Assert(os.Remove(path), qt.IsNil)
	c.Assert(
		cli.Run(t.Context(), []string{"report", "result.json", "--format", "html", "--output", "transformed.html"}, environment),
		qt.Equals,
		0,
	)
	first, err := fs.ReadFile(os.DirFS(root), "result.html")
	c.Assert(err, qt.IsNil)
	second, err := fs.ReadFile(os.DirFS(root), "transformed.html")
	c.Assert(err, qt.IsNil)
	c.Assert(string(second), qt.Equals, string(first))
}

func TestProtectInputsAndDuplicateOutputs(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	path := filepath.Join(root, "draft.md")
	source := "Certainly! The client opens connections."
	c.Assert(os.WriteFile(path, []byte(source), 0o600), qt.IsNil)
	cases := [][]string{
		{"check", "draft.md", "--report", "json:draft.md"},
		{"check", "draft.md", "--report", "json:-", "--report", "text:-"},
		{"check", "draft.md", "--report", "json:result.json", "--report", "text:./result.json"},
	}
	for _, args := range cases {
		var output, stderr bytes.Buffer
		c.Assert(
			cli.Run(t.Context(), args, cli.Environment{Dir: root, In: strings.NewReader(""), Out: &output, Err: &stderr}),
			qt.Equals,
			2,
		)
		actual, err := fs.ReadFile(os.DirFS(root), "draft.md")
		c.Assert(err, qt.IsNil)
		c.Assert(string(actual), qt.Equals, source)
	}
}

type failedWriter struct{}

func (failedWriter) Write(_ []byte) (int, error) { return 0, errors.New("injected writer failure") }

func TestWriterFailureAndUnknownFormat(t *testing.T) {
	c := qt.New(t)
	environment := cli.Environment{Dir: t.TempDir(), In: strings.NewReader("Clean prose."), Out: failedWriter{}, Err: io.Discard}
	c.Assert(cli.Run(t.Context(), []string{"check", "--stdin", "--filename", "a.md", "--report", "json:-"}, environment), qt.Equals, 2)
	environment.In = strings.NewReader("Clean prose.")
	c.Assert(cli.Run(t.Context(), []string{"check", "--stdin", "--format", "html"}, environment), qt.Equals, 2)
}
