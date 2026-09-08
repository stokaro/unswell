package cli_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
)

func TestCheckAndExplanationResolveTheSameInheritedPolicy(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	for _, directory := range []string{"docs", "policies"} {
		c.Assert(os.Mkdir(filepath.Join(root, directory), 0o700), qt.IsNil)
	}
	files := map[string]string{
		".unswell.yaml": "version: 1\nextends: [policies/base.yaml]\noverrides:\n" +
			"  - files: [docs/reference.md]\n    rules: {policy.banned-phrases: {enabled: false}}\n",
		"policies/base.yaml": "version: 1\nextends: [builtin:custom]\n" +
			"rules: {policy.banned-phrases: {enabled: true, parameters: {phrases: [robust]}}}\n" +
			"vocabulary: {dictionaries: [terms.yaml], term_exemptions: [policy.banned-phrases]}\n",
		"policies/terms.yaml": "version: 1\nterms: [robust estimator]\n",
		"docs/guide.md":       "The robust estimator uses robust methods.\n",
		"docs/reference.md":   "The robust estimator uses robust methods.\n",
	}
	for name, data := range files {
		c.Assert(os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte(data), 0o600), qt.IsNil)
	}
	dir := filepath.Join(root, "docs")
	checked := runRulesCLI(t, dir, []string{"check", ".", "--project-root", root, "--report", "json:-"}, 1)
	var result unswell.RunResult
	c.Assert(json.Unmarshal([]byte(checked), &result), qt.IsNil)
	c.Assert(result.Documents, qt.HasLen, 2)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Path, qt.Equals, "docs/guide.md")
	for _, doc := range result.Documents {
		explained := runRulesCLI(t, dir, []string{"config", "explain", "--project-root", root, "--file", filepath.Base(doc.Name)}, 0)
		var explanation struct {
			Policy config.Policy `json:"policy"`
		}
		c.Assert(json.Unmarshal([]byte(explained), &explanation), qt.IsNil)
		c.Assert(doc.ConfigHash, qt.Equals, explanation.Policy.Hash)
	}
	before, err := fs.ReadFile(os.DirFS(root), "policies/terms.yaml")
	c.Assert(err, qt.IsNil)
	runRulesCLI(t, root, []string{"check", "docs", "--report", "json:policies/terms.yaml"}, 2)
	after, err := fs.ReadFile(os.DirFS(root), "policies/terms.yaml")
	c.Assert(err, qt.IsNil)
	c.Assert(after, qt.DeepEquals, before)
}

func TestDocumentedConfigurationExamples(t *testing.T) {
	c := qt.New(t)
	guide, err := fs.ReadFile(os.DirFS("../.."), "docs/configuration.md")
	c.Assert(err, qt.IsNil)
	examples := regexp.MustCompile("(?s)```yaml\\n(.*?)\\n```").FindAllSubmatch(guide, -1)
	c.Assert(examples, qt.HasLen, 3)
	root := t.TempDir()
	c.Assert(os.Mkdir(filepath.Join(root, "policies"), 0o700), qt.IsNil)
	for i, name := range []string{".unswell.yaml", "policies/team.yaml", "policies/terms.yaml"} {
		c.Assert(os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), examples[i][1], 0o600), qt.IsNil)
	}
	runRulesCLI(t, root, []string{"config", "validate"}, 0)
	for _, name := range []string{"docs/guide.md", "docs/reference/api.md"} {
		output := runRulesCLI(t, root, []string{"config", "explain", "--file", name}, 0)
		var explanation struct {
			Policy config.Policy `json:"policy"`
		}
		c.Assert(json.Unmarshal([]byte(output), &explanation), qt.IsNil)
		c.Assert(explanation.Policy.Sources, qt.HasLen, 3)
		c.Assert(explanation.Policy.Vocabulary.ResolvedTerms, qt.HasLen, 4)
		c.Assert(explanation.Policy.Rules["hype.modifier-cluster"].Enabled, qt.Equals, name == "docs/guide.md")
	}
}

func TestSurfaceConfigurationExample(t *testing.T) {
	c := qt.New(t)
	guide, err := fs.ReadFile(os.DirFS("../.."), "docs/surface-signals.md")
	c.Assert(err, qt.IsNil)
	examples := regexp.MustCompile("(?s)```yaml\\n(.*?)\\n```").FindAllSubmatch(guide, -1)
	c.Assert(examples, qt.HasLen, 1)
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, ".unswell.yaml"), examples[0][1], 0o600), qt.IsNil)
	runRulesCLI(t, root, []string{"config", "validate"}, 0)
	output := runRulesCLI(t, root, []string{"config", "explain", "--file", "guide.md"}, 0)
	var explanation struct {
		Policy config.Policy `json:"policy"`
	}
	c.Assert(json.Unmarshal([]byte(output), &explanation), qt.IsNil)
	for _, id := range []string{"syntax.nominalization-chain", "syntax.noun-stack", "readability.grade-metric", "format.list-fragmentation"} {
		c.Assert(explanation.Policy.Rules[id].Enabled, qt.IsTrue)
		c.Assert(explanation.Policy.Rules[id].Gate, qt.Equals, "none")
	}
	c.Assert(explanation.Policy.Rules["readability.grade-metric"].Score.Weight, qt.Equals, 0)
}
