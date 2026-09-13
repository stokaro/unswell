package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestWorkflowShellPrecedenceAndDefaults(t *testing.T) {
	for _, row := range []struct{ name, workflow, job, step, format string }{
		{"step", "defaults:\n  run:\n    shell: fish\n", "    defaults:\n      run:\n        shell: pwsh\n", "        shell: bash\n", "bash"},
		{"job", "defaults:\n  run:\n    shell: fish\n", "    defaults:\n      run:\n        shell: pwsh\n", "", "powershell"},
		{"workflow", "defaults:\n  run:\n    shell: fish\n", "", "", "fish"},
		{"Linux", "", "    runs-on: ubuntu-latest\n", "", "bash"},
		{"macOS", "", "    runs-on: macos-15-intel\n", "", "bash"},
		{"Windows", "", "    runs-on: windows-latest\n", "", "powershell"},
		{"container", "", "    runs-on: ubuntu-latest\n    container: alpine\n", "", "sh"},
		{"container override", "defaults:\n  run:\n    shell: bash\n", "    container: alpine\n", "", "bash"},
		{"template", "", "", "        shell: bash --noprofile --norc -e -o pipefail {0}\n", "bash"},
		{"zsh", "", "", "        shell: zsh {0}\n", "zsh"},
		{"PowerShell Desktop", "", "", "        shell: powershell\n", "powershell"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			text := row.workflow + "jobs:\n  check:\n" + row.job + "    steps:\n      - run: |\n          # Keep the comment.\n" + row.step
			doc := parseWorkflow(t, text, extract.Policy{})
			found := false
			for _, block := range doc.Blocks {
				if strings.TrimSpace(block.Text) == "Keep the comment." {
					found = true
					c.Assert(block.Context, qt.Contains, "embedded:"+row.format+":comment")
				}
			}
			c.Assert(found, qt.IsTrue)
		})
	}
}

func TestWorkflowUnknownShellsAndExpressionsHaveReasons(t *testing.T) {
	for _, row := range []struct{ name, job, shell, program, reason string }{
		{"unknown", "", "xonsh", "echo 'Keep the message.'", "actions-shell-unknown"},
		{"cmd", "", "cmd", "echo Keep the message.", "actions-shell-unknown"},
		{"python", "", "python", "print('Keep the message.')", "actions-shell-unknown"},
		{"wrapper", "", "env bash {0}", "echo 'Keep the message.'", "actions-shell-unknown"},
		{"bad template", "", "bash -c", "echo 'Keep the message.'", "actions-shell-unknown"},
		{"shell expression", "", "${{ matrix.shell }}", "echo 'Keep the message.'", "actions-shell-unknown"},
		{"unknown runner", "    runs-on: custom-linux\n", "", "echo 'Keep the message.'", "actions-shell-unknown"},
		{"runner expression", "    runs-on: ${{ matrix.os }}\n", "", "echo 'Keep the message.'", "actions-shell-unknown"},
		{"body expression", "", "bash", "echo '${{ inputs.message }}'", "actions-expression"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			text := "jobs:\n  check:\n" + row.job + "    steps:\n      - name: Keep the step name.\n        run: |\n          " + row.program + "\n"
			if row.shell != "" {
				text += "        shell: " + row.shell + "\n"
			}
			doc := parseWorkflow(t, text, extract.Policy{})
			c.Assert(excludedTexts(doc, text, row.reason), qt.HasLen, 1)
			for _, block := range doc.Blocks {
				c.Assert(block.Text, qt.Not(qt.Contains), "echo")
			}
		})
	}
}

func TestWorkflowInvalidProgramsAndLimitsAreErrors(t *testing.T) {
	for _, row := range []struct {
		name, text, want string
		max              int
	}{
		{"shell syntax", "jobs:\n  check:\n    steps:\n      - shell: bash\n        run: echo \"unfinished\n", ".*GitHub Actions bash run.*", 0},
		{"duplicate shell", "jobs:\n  check:\n    steps:\n      - shell: bash\n        shell: fish\n        run: echo hello\n",
			".*duplicate GitHub Actions field.*", 0},
		{"block budget", workflowProgram, ".*source exceeds.*prose.*", 3},
	} {
		t.Run(row.name, func(t *testing.T) {
			_, err := extract.Parse(t.Context(), document.Source{Name: workflowPath, Format: document.YAML, Bytes: []byte(row.text)},
				extract.Options{MaxBlocks: row.max})
			qt.New(t).Assert(err, qt.ErrorMatches, row.want)
		})
	}
}

func TestWorkflowSyntaxDoesNotConsumeTheProseBlockBudget(t *testing.T) {
	for _, run := range []string{"exit 1", "''", "|\n"} {
		t.Run(run, func(t *testing.T) {
			c := qt.New(t)
			source := "jobs:\n  check:\n    steps:\n      - shell: bash\n        run: " + run + "\n"
			doc, err := extract.Parse(t.Context(), document.Source{Name: workflowPath, Format: document.YAML, Bytes: []byte(source)},
				extract.Options{MaxBlocks: 1})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Equals, "bash")
		})
	}
}
