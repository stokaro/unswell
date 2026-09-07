package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/internal/cli"
)

func TestLanguageContextsThroughCLI(t *testing.T) {
	cases := []struct {
		name, source string
		code         int
	}{
		{
			"sample.cs",
			"// The client opens connections.\nclass Sample { string message = \"It is important to note that the client opens connections.\"; }",
			0,
		},
		{"sample.yaml", "# The client opens connections.\nmessage: It is important to note that the client opens connections.\n", 1},
		{"broken.cs", "class Sample { string value = \"unfinished", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			root := t.TempDir()
			policy := "version: 1\nextends: [builtin:strict-v1]\nextraction:\n  contexts: [comment]\n" +
				"  languages:\n    yaml:\n      contexts: [comment, string]\n"
			c.Assert(os.WriteFile(filepath.Join(root, ".unswell.yaml"), []byte(policy), 0o600), qt.IsNil)
			var output, stderr bytes.Buffer
			code := cli.Run(t.Context(), []string{"check", "--stdin", "--filename", tc.name, "--report", "json:-"},
				cli.Environment{Dir: root, In: strings.NewReader(tc.source), Out: &output, Err: &stderr})
			c.Assert(code, qt.Equals, tc.code, qt.Commentf("%s", stderr.String()))
			var result unswell.RunResult
			c.Assert(json.Unmarshal(output.Bytes(), &result), qt.IsNil)
			c.Assert(result.Gate.Passed, qt.Equals, tc.code == 0)
		})
	}
}
