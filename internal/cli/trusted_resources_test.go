package cli_test

import (
	"io/fs"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestTrustedDictionaryExemptionsComeFromBase(t *testing.T) {
	c := qt.New(t)
	root, base := trustRepository(t, map[string]string{
		".unswell.yaml": "version: 1\nextends: [policies/base.yaml]\n",
		"policies/base.yaml": trustedPolicy +
			"vocabulary: {dictionaries: [terms.yaml], term_exemptions: [policy.banned-phrases]}\n",
		"policies/terms.yaml": "version: 1\nterms: [robust estimator]\n",
	})
	writeChanged(t, root, "policies/terms.yaml", "version: 1\nterms: [robust]\n")
	commitGit(t, root, "commit", "-am", "Expand candidate exemption")
	code, result, stderr := runTrusted(t, root, base, "guide.md")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	assertPolicyChange(t, result, "policies/terms.yaml", "dictionary")
	commitGit(t, root, "rm", "policies/base.yaml")
	commitGit(t, root, "commit", "-m", "Remove inherited candidate policy")
	code, result, stderr = runTrusted(t, root, base, "guide.md")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	assertPolicyChange(t, result, "policies/base.yaml", "config")
}

func TestTrustedRulesSurviveRemovalAndDirectoryMembershipChanges(t *testing.T) {
	c := qt.New(t)
	data, err := os.ReadFile("../../examples/rules/company.yaml")
	c.Assert(err, qt.IsNil)
	root, base := trustRepository(t, map[string]string{
		".unswell.yaml": "version: 1\nextends: [builtin:custom]\nrules:\n  company.no-dive-in: {enabled: true}\n",
		"guide.md":      "Let's dive into the configuration options.\n", "rules/company.yaml": string(data),
	})
	writeChanged(t, root, "rules/extra.yaml", "invalid candidate: [\n")
	commitGit(t, root, "add", "rules/extra.yaml")
	commitGit(t, root, "commit", "-m", "Add candidate ruleset")
	code, result, stderr := runTrusted(t, root, base, "guide.md", "--ruleset", "rules")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	c.Assert(result.Findings[0].RuleID, qt.Equals, "company.no-dive-in")
	assertPolicyChange(t, result, "rules", "ruleset-directory")
	commitGit(t, root, "rm", "rules/company.yaml")
	commitGit(t, root, "commit", "-m", "Remove candidate rule")
	code, result, stderr = runTrusted(t, root, base, "guide.md", "--ruleset", "rules")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	assertPolicyChange(t, result, "rules/company.yaml", "ruleset")
	writeChanged(t, root, "rules/company.yaml", string(data))
	code, _, stderr = runTrusted(t, root, base, "guide.md", "--ruleset", "rules", "--no-gate")
	c.Assert(code, qt.Equals, 2)
	c.Assert(stderr, qt.Contains, "deleted file is present")
}

func TestTrustedBaselineDoesNotAcceptExpandedCandidateDebt(t *testing.T) {
	c := qt.New(t)
	root, _ := trustRepository(t, map[string]string{".unswell.yaml": trustedPolicy + "gate: {mode: new}\n"})
	runRulesCLI(t, root, []string{"baseline", "create", "guide.md", "--output", "debt.json"}, 0)
	commitGit(t, root, "add", "debt.json")
	commitGit(t, root, "commit", "-m", "Accept initial debt")
	base := commitGit(t, root, "rev-parse", "HEAD")
	writeChanged(t, root, "guide.md", "The robust client retries.\n\nAnother robust implementation caches responses.\n")
	runRulesCLI(t, root, []string{"baseline", "update", "guide.md", "--baseline", "debt.json"}, 0)
	commitGit(t, root, "commit", "-am", "Expand candidate debt")
	code, result, stderr := runTrusted(t, root, base, "guide.md", "--baseline", "debt.json")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	c.Assert(len(result.Gate.Accepted) > 0, qt.IsTrue)
	c.Assert(len(result.Gate.Reasons) > 0, qt.IsTrue)
	assertPolicyChange(t, result, "debt.json", "baseline")
	runRulesCLI(t, root, []string{"baseline", "check", "guide.md", "--baseline", "debt.json"}, 0)
	runRulesCLI(t, root, []string{"baseline", "check", "guide.md", "--baseline", "debt.json",
		"--changed-from", base, "--policy-from-base"}, 1)
	before, err := fs.ReadFile(os.DirFS(root), "debt.json")
	c.Assert(err, qt.IsNil)
	code, _, _ = runTrusted(t, root, base, "guide.md", "--baseline", "debt.json", "--report", "json:debt.json")
	c.Assert(code, qt.Equals, 2)
	after, err := fs.ReadFile(os.DirFS(root), "debt.json")
	c.Assert(err, qt.IsNil)
	c.Assert(after, qt.DeepEquals, before)
}

func TestTrustedExplicitRulesetUsesBaseBytes(t *testing.T) {
	c := qt.New(t)
	data, err := os.ReadFile("../../examples/rules/company.yaml")
	c.Assert(err, qt.IsNil)
	root, base := trustRepository(t, map[string]string{"rules.yaml": string(data)})
	writeChanged(t, root, "rules.yaml", strings.ReplaceAll(string(data), "dive", "walk"))
	commitGit(t, root, "commit", "-am", "Change candidate matcher")
	code, result, stderr := runTrusted(t, root, base, "guide.md", "--ruleset", "rules.yaml")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	assertPolicyChange(t, result, "rules.yaml", "ruleset")
}
