package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

const workflowPath = ".github/workflows/check.yml"
const workflowProgram = `jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - name: Check the release.
        shell: bash
        run: |
          checker=$(cd tools && go tool -n govulncheck)
          while read -r directory role; do
            if [[ "$role" != tools ]]; then
              (cd "$directory" && "$checker" ./...)
            fi
          done <.gomodules
          # It is important to note that the client may retry.
          echo "It is important to note that the limit is 30 seconds."
`

func parseWorkflow(t *testing.T, text string, policy extract.Policy) document.Document {
	t.Helper()
	doc, err := extract.Parse(t.Context(), document.Source{Name: workflowPath, Format: document.YAML, Bytes: []byte(text)},
		extract.Options{Policy: policy, IncludeStructure: true})
	qt.New(t).Assert(err, qt.IsNil)
	return doc
}

func TestWorkflowShellRetainsProseAndOriginalMapping(t *testing.T) {
	c := qt.New(t)
	text := "\ufeff" + strings.ReplaceAll(workflowProgram, "\n", "\r\n")
	doc := parseWorkflow(t, text, extract.Policy{})
	c.Assert(string(doc.Source), qt.Equals, text)
	c.Assert(doc.Blocks, qt.HasLen, 5) // Runner, step name, shell name, comment, and echo message.
	c.Assert(strings.TrimSpace(doc.Blocks[3].Text), qt.Equals, "It is important to note that the client may retry.")
	c.Assert(doc.Blocks[4].Text, qt.Equals, "It is important to note that the limit is 30 seconds.")
	for _, block := range doc.Blocks[3:] {
		c.Assert(block.Kind, qt.Equals, "string")
		c.Assert(block.Context, qt.Contains, "block_mapping_pair:key:run")
		for i, origin := range block.Map {
			c.Assert(origin.Valid(len(text)), qt.IsTrue)
			if block.Text[i] != ' ' && block.Text[i] != '\r' {
				c.Assert(text[origin.Start:origin.End], qt.Equals, block.Text[i:i+1])
			}
		}
	}
	c.Assert(doc.Blocks[3].Context, qt.Contains, "embedded:bash:comment")
	c.Assert(doc.Blocks[4].Context, qt.Contains, "embedded:bash:string")
	c.Assert(strings.Join(excludedTexts(doc, text, "actions-shell-syntax"), ""),
		qt.Contains, "checker=$(cd tools && go tool -n govulncheck)")
}

func TestWorkflowShellDecodesYAMLBeforeShellStrings(t *testing.T) {
	for _, row := range []struct{ name, value, want string }{
		{"folded", ">-\n          echo \"It is important\n          to note that the limit is 30 seconds.\"\n",
			"It is important to note that the limit is 30 seconds."},
		{"double quoted", `"echo \"It is \\u0069mportant to note that.\""` + "\n", "It is \\u0069mportant to note that."},
		{"YAML unicode", `"echo \"It is \u0069mportant to note that.\""` + "\n", "It is important to note that."},
		{"shell unicode", `"echo $'Let\\u2019s dive into details.'"` + "\n", "Let’s dive into details."},
		{"single quoted", `'echo "It is important to note that."'` + "\n", "It is important to note that."},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			text := "jobs:\n  check:\n    runs-on: ubuntu-latest\n    steps:\n      - run: " + row.value
			doc := parseWorkflow(t, text, extract.Policy{})
			c.Assert(doc.Blocks, qt.HasLen, 2) // Runner label and one decoded message.
			c.Assert(doc.Blocks[1].Text, qt.Equals, row.want)
			for _, origin := range doc.Blocks[1].Map {
				c.Assert(origin.Valid(len(text)), qt.IsTrue)
			}
		})
	}
}

func TestWorkflowSelectionIsBoundedToRunSteps(t *testing.T) {
	c := qt.New(t)
	text := "jobs:\n  check:\n    steps:\n      - shell: bash\n        run: echo \"Keep the message.\"\n" +
		"        with:\n          run: echo \"This is an action input.\"\n"
	doc := parseWorkflow(t, text, extract.Policy{})
	c.Assert(doc.Blocks[1].Text, qt.Equals, "Keep the message.")
	c.Assert(doc.Blocks[2].Text, qt.Equals, `echo "This is an action input."`)
	for _, name := range []string{"settings.yaml", ".github/workflows/deep/check.yml", "docs/check.yml"} {
		doc, err := extract.Parse(t.Context(), document.Source{Name: name, Format: document.YAML, Bytes: []byte(text)}, extract.Options{})
		c.Assert(err, qt.IsNil)
		c.Assert(doc.Blocks[1].Text, qt.Equals, `echo "Keep the message."`)
	}
}

func TestWorkflowContextAndExceptionPrecedence(t *testing.T) {
	c := qt.New(t)
	doc := parseWorkflow(t, workflowProgram, extract.Policy{Languages: map[document.Format]extract.LanguagePolicy{
		document.YAML: {Contexts: []string{"string"}}, document.Bash: {Contexts: []string{"comment"}},
	}})
	c.Assert(doc.Blocks, qt.HasLen, 4)
	c.Assert(doc.Blocks[3].Context, qt.Contains, "embedded:bash:comment")
	doc = parseWorkflow(t, workflowProgram, extract.Policy{Languages: map[document.Format]extract.LanguagePolicy{
		document.YAML: {Contexts: []string{"comment"}},
	}})
	c.Assert(doc.Blocks, qt.HasLen, 0)
	policy := extract.Policy{Exceptions: []extract.Exception{{ID: "external-program", Paths: []string{workflowPath},
		Formats: []document.Format{document.YAML}, Kinds: []string{"string"}, Symbols: []string{"run"}, Reason: "External script contract."}}}
	doc = parseWorkflow(t, workflowProgram, policy)
	c.Assert(doc.Blocks, qt.HasLen, 3)
	c.Assert(excludedTexts(doc, workflowProgram, "config:external-program: External script contract."), qt.HasLen, 1)
	c.Assert(excludedTexts(doc, workflowProgram, "actions-shell-syntax"), qt.HasLen, 0)
	policy = extract.Policy{GitHubActions: "strings"}
	doc = parseWorkflow(t, workflowProgram, policy)
	c.Assert(doc.Blocks, qt.HasLen, 4)
	c.Assert(strings.HasPrefix(doc.Blocks[3].Text, "checker=$(cd tools"), qt.IsTrue)
}
