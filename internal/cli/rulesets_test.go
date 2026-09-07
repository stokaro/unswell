package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/internal/cli"
	"github.com/stokaro/unswell/rule"
)

const customPolicy = "version: 1\nextends: [builtin:custom-v1]\nrules:\n  company.no-dive-in:\n    enabled: true\n"

func rulesWorkspace(t *testing.T) (string, []byte) {
	t.Helper()
	c := qt.New(t)
	root := t.TempDir()
	data, err := os.ReadFile("../../examples/rules/company.yaml")
	c.Assert(err, qt.IsNil)
	for name, content := range map[string][]byte{
		"company.yaml": data, "policy.yaml": []byte(customPolicy),
		"draft.md": []byte("Let's dive into the configuration options."),
	} {
		c.Assert(os.WriteFile(filepath.Join(root, name), content, 0o600), qt.IsNil)
	}
	return root, data
}

func runRulesCLI(t *testing.T, root string, args []string, code int) string {
	t.Helper()
	c := qt.New(t)
	var output, stderr bytes.Buffer
	actual := cli.Run(t.Context(), args, cli.Environment{Dir: root, In: strings.NewReader(""), Out: &output, Err: &stderr})
	c.Assert(actual, qt.Equals, code, qt.Commentf("stdout: %s\nstderr: %s", output.String(), stderr.String()))
	return output.String() + stderr.String()
}

func TestRulePackCommands(t *testing.T) {
	t.Parallel()
	root, _ := rulesWorkspace(t)
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"examples", []string{"rules", "test", "company.yaml"}, "Passed 2 catalog examples."},
		{"list", []string{"rules", "list", "--ruleset", "company.yaml"}, "company.no-dive-in\tcustom\t"},
		{"show", []string{"rules", "show", "company.no-dive-in", "--ruleset", "company.yaml"}, `"namespace": "company"`},
		{"validate", []string{"config", "validate", "--config", "policy.yaml", "--ruleset", "company.yaml"}, "Configuration valid:"},
		{"policy", []string{"config", "explain", "--config", "policy.yaml", "--ruleset", "company.yaml"}, `"rule_sets": [`},
		{"doctor", []string{"doctor", "--config", "policy.yaml", "--ruleset", "company.yaml"}, `"namespace": "company"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			output := runRulesCLI(t, root, tc.args, 0)
			c.Assert(output, qt.Contains, tc.want)
		})
	}
}

func TestExternalAndInlinePacksPreserveFindingsAndIdentifySources(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	root, data := rulesWorkspace(t)
	inline := customPolicy + "rule_sets:\n  - " + strings.ReplaceAll(strings.TrimSpace(string(data)), "\n", "\n    ") + "\n"
	c.Assert(os.WriteFile(filepath.Join(root, "inline.yaml"), []byte(inline), 0o600), qt.IsNil)
	externalArgs := []string{"check", "draft.md", "--config", "policy.yaml", "--ruleset", "company.yaml", "--report", "json:-"}
	inlineArgs := []string{"check", "draft.md", "--config", "inline.yaml", "--report", "json:-"}
	var external, embedded unswell.RunResult
	c.Assert(json.Unmarshal([]byte(runRulesCLI(t, root, externalArgs, 1)), &external), qt.IsNil)
	c.Assert(json.Unmarshal([]byte(runRulesCLI(t, root, inlineArgs, 1)), &embedded), qt.IsNil)
	c.Assert(external.Findings, qt.HasLen, 1)
	c.Assert(external.Findings[0].RuleID, qt.Equals, "company.no-dive-in")
	c.Assert(external.Manifest.ConfigHash, qt.Not(qt.Equals), embedded.Manifest.ConfigHash)
	c.Assert(external.Manifest.ConfigSources[0].Path, qt.Equals, "policy.yaml")
	c.Assert(embedded.Manifest.ConfigSources[0].Path, qt.Equals, "inline.yaml")
	c.Assert(external.Manifest.RulesetHash, qt.Equals, embedded.Manifest.RulesetHash)
	c.Assert(external.Findings, qt.DeepEquals, embedded.Findings)
}

func TestRulePackFailuresCannotPass(t *testing.T) {
	t.Parallel()
	root, data := rulesWorkspace(t)
	c := qt.New(t)
	wrongExample := strings.Replace(string(data), "- \"Let's dive into the configuration options.\"",
		"- \"The client opens a connection.\"", 1)
	c.Assert(os.WriteFile(filepath.Join(root, "wrong-example.yaml"), []byte(wrongExample), 0o600), qt.IsNil)
	invalid := strings.Replace(string(data), "type: phrase", "type: shell", 1)
	c.Assert(os.WriteFile(filepath.Join(root, "invalid.yaml"), []byte(invalid), 0o600), qt.IsNil)
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"bad example", []string{"rules", "test", "wrong-example.yaml"}, "expected match=true, got 0 findings"},
		{"before scan", []string{"check", "missing.md", "--ruleset", "invalid.yaml"}, `unknown matcher type "shell"`},
		{"doctor", []string{"doctor", "--ruleset", "invalid.yaml"}, `unknown matcher type "shell"`},
		{"unknown rule", []string{"rules", "show", "company.missing", "--ruleset", "company.yaml"}, `unknown rule ID`},
		{"duplicate", []string{"rules", "list", "--ruleset", "company.yaml", "--ruleset", "company.yaml"}, `duplicate rule ID`},
		{"conflicting flags", []string{"rules", "test", "company.yaml", "--config", "policy.yaml"}, "cannot be combined"},
		{"unused flag", []string{"config", "init", "--ruleset", "company.yaml"}, "does not accept --ruleset"},
		{"overwrite", []string{"check", "draft.md", "--ruleset", "company.yaml", "--report", "json:company.yaml"}, "cannot overwrite source"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			output := runRulesCLI(t, root, tc.args, 2)
			c.Assert(output, qt.Contains, tc.want, qt.Commentf("%s", output))
		})
	}
	after, err := fs.ReadFile(os.DirFS(root), "company.yaml")
	c.Assert(err, qt.IsNil)
	c.Assert(after, qt.DeepEquals, data)
}

func TestRuleDirectoryBoundsAndOrder(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	root, data := rulesWorkspace(t)
	directory := filepath.Join(root, "rules")
	c.Assert(os.Mkdir(directory, 0o700), qt.IsNil)
	output := runRulesCLI(t, root, []string{"rules", "test", "rules"}, 2)
	c.Assert(output, qt.Contains, "contains no YAML files")
	for _, name := range []string{"zebra", "alpha"} {
		pack := strings.ReplaceAll(string(data), "company", name)
		c.Assert(os.WriteFile(filepath.Join(directory, name+".yaml"), []byte(pack), 0o600), qt.IsNil)
	}
	output = runRulesCLI(t, root, []string{"rules", "list", "--ruleset", "rules"}, 0)
	c.Assert(strings.Index(output, "alpha.no-dive-in") < strings.Index(output, "zebra.no-dive-in"), qt.IsTrue)
	for i := range 9 {
		c.Assert(os.WriteFile(filepath.Join(directory, fmt.Sprintf("pack%d.yaml", i)), data, 0o600), qt.IsNil)
	}
	output = runRulesCLI(t, root, []string{"rules", "test", "rules"}, 2)
	c.Assert(output, qt.Contains, "at most 10 ruleset files")
}

func TestRuleCommandWriterFailure(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	root, _ := rulesWorkspace(t)
	environment := cli.Environment{Dir: root, In: strings.NewReader(""), Out: failedWriter{}, Err: io.Discard}
	c.Assert(cli.Run(t.Context(), []string{"rules", "test", "company.yaml"}, environment), qt.Equals, 2)
	c.Assert(cli.Run(t.Context(), []string{"rules", "show", "company.no-dive-in", "--ruleset", "company.yaml"}, environment), qt.Equals, 2)
}

func TestFormattedRuleExamples(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	root, data := rulesWorkspace(t)
	pack := strings.Replace(string(data), "- \"Let's dive into the configuration options.\"",
		`- {format: go, text: "package sample\n// Let's dive into the configuration options.\n"}`, 1)
	c.Assert(os.WriteFile(filepath.Join(root, "company.yaml"), []byte(pack), 0o600), qt.IsNil)
	runRulesCLI(t, root, []string{"rules", "test", "company.yaml"}, 0)
	var descriptor rule.Descriptor
	output := runRulesCLI(t, root, []string{"rules", "show", "company.no-dive-in", "--ruleset", "company.yaml"}, 0)
	c.Assert(json.Unmarshal([]byte(output), &descriptor), qt.IsNil)
	c.Assert(string(descriptor.Examples[0].Format), qt.Equals, "go")
}

func TestExplainUsesCustomRuleWeights(t *testing.T) {
	t.Parallel()
	c := qt.New(t)
	root, _ := rulesWorkspace(t)
	policy := customPolicy + "    score: {weight: 12, cap: 12}\n"
	c.Assert(os.WriteFile(filepath.Join(root, "policy.yaml"), []byte(policy), 0o600), qt.IsNil)
	output := runRulesCLI(t, root, []string{"explain", "draft.md", "--at", "1:1",
		"--config", "policy.yaml", "--ruleset", "company.yaml"}, 0)
	var assessments []unswell.Assessment
	c.Assert(json.Unmarshal([]byte(output), &assessments), qt.IsNil)
	c.Assert(len(assessments) > 0, qt.IsTrue)
	for _, assessment := range assessments {
		c.Assert(assessment.SlopScore, qt.Equals, float64(12))
		c.Assert(assessment.Contributions, qt.HasLen, 1)
		c.Assert(assessment.Contributions[0].RuleID, qt.Equals, "company.no-dive-in")
		c.Assert(assessment.SlopProbability, qt.IsNil)
	}
}
