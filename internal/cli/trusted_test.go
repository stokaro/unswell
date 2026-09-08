package cli_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

const trustedPolicy = "version: 1\nextends: [builtin:custom]\nrules:\n" +
	"  policy.banned-phrases: {enabled: true, gate: forbid, parameters: {phrases: [robust]}}\n"

func trustRepository(t *testing.T, files map[string]string) (string, string) {
	t.Helper()
	root, _ := changedRepository(t)
	writeChanged(t, root, ".unswell.yaml", trustedPolicy)
	writeChanged(t, root, "guide.md", "The robust client retries.\n")
	for name, data := range files {
		qt.New(t).Assert(os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0o700), qt.IsNil)
		writeChanged(t, root, name, data)
	}
	commitGit(t, root, "add", ".")
	commitGit(t, root, "commit", "-m", "Establish trusted policy")
	return root, commitGit(t, root, "rev-parse", "HEAD")
}

func runTrusted(t *testing.T, root, base string, extra ...string) (int, unswell.RunResult, string) {
	t.Helper()
	return runChanged(t, root, base, append([]string{"--policy-from-base"}, extra...)...)
}

func assertPolicyChange(t *testing.T, result unswell.RunResult, path, kind string) {
	t.Helper()
	c := qt.New(t)
	c.Assert(result.PolicyComparison, qt.IsNotNil)
	c.Assert(result.PolicyComparison.Complete, qt.IsTrue)
	c.Assert(slices.ContainsFunc(result.PolicyComparison.Changes, func(change unswell.PolicyChange) bool {
		return change.Path == path && change.Kind == kind && change.BeforeHash != change.AfterHash
	}), qt.IsTrue, qt.Commentf("%+v", result.PolicyComparison.Changes))
}

func TestTrustedConfigurationCannotRelaxRulesOrSourceSelection(t *testing.T) {
	for _, candidate := range []string{
		"version: 1\nextends: [builtin:custom]\n",
		trustedPolicy + "files:\n  exclude: ['**/*.md']\n",
		"invalid: [\n",
		"",
	} {
		t.Run(candidate, func(t *testing.T) {
			c := qt.New(t)
			root, base := trustRepository(t, map[string]string{"docs/guide.md": "The robust client retries.\n"})
			if candidate == "" {
				commitGit(t, root, "rm", ".unswell.yaml")
			} else {
				writeChanged(t, root, ".unswell.yaml", candidate)
			}
			commitGit(t, root, "commit", "-am", "Change candidate policy")
			selected := "guide.md"
			if strings.Contains(candidate, "files:") {
				selected = "docs"
				runRulesCLI(t, root, []string{"config", "validate"}, 0)
				runRulesCLI(t, root, []string{"check", selected, "--allow-empty"}, 0)
			}
			code, result, stderr := runTrusted(t, root, base, selected)
			c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
			c.Assert(result.Findings[0].RuleID, qt.Equals, "policy.banned-phrases")
			c.Assert(result.Manifest.Git.BaseCommit, qt.Equals, base)
			c.Assert(result.Manifest.Git.Clean, qt.IsTrue)
			c.Assert(result.Changes.SelectedUnits, qt.Equals, result.Changes.ComparedUnits)
			c.Assert(result.Changes.Documents[0].FullReason, qt.Equals, "policy_changed")
			assertPolicyChange(t, result, ".unswell.yaml", "config")
		})
	}
}

func TestTrustedDiscoveryUsesBaseAndKeepsCleanProsePassing(t *testing.T) {
	c := qt.New(t)
	root, base := trustRepository(t, map[string]string{"docs/guide.md": "The robust client retries.\n"})
	writeChanged(t, root, "docs/.unswell.yaml", "invalid candidate: [\n")
	commitGit(t, root, "add", "docs/.unswell.yaml")
	commitGit(t, root, "commit", "-m", "Add closer candidate policy")
	code, result, stderr := runTrusted(t, filepath.Join(root, "docs"), base, "--project-root", root, "guide.md")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	assertPolicyChange(t, result, "docs/.unswell.yaml", "config-discovery")
	writeChanged(t, root, "docs/guide.md", "The client retries.\n")
	commitGit(t, root, "commit", "-am", "Edit the prose")
	code, result, stderr = runTrusted(t, filepath.Join(root, "docs"), base, "--project-root", root, "guide.md")
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s", stderr))
	c.Assert(result.PolicyComparison.FullScan, qt.IsTrue)
}

func TestTrustedSourcePermissionsDoNotAcceptANewException(t *testing.T) {
	c := qt.New(t)
	root, base := trustRepository(t, nil)
	const prose = "<!-- unswell-disable-next-block policy.banned-phrases -- External wording. -->\n\nThe robust client retries.\n"
	writeChanged(t, root, "guide.md", prose)
	commitGit(t, root, "commit", "-am", "Add a candidate exception")
	code, result, stderr := runTrusted(t, root, base, "guide.md")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	c.Assert(result.Suppressions[0].TrustState, qt.Equals, "untrusted")
	c.Assert(result.Findings[0].Suppressed, qt.IsFalse)
	assertPolicyChange(t, result, "guide.md", "source-permissions")
	base = commitGit(t, root, "rev-parse", "HEAD")
	writeChanged(t, root, "guide.md", "A separate introduction.\n\n"+prose)
	commitGit(t, root, "commit", "-am", "Move previously accepted prose")
	code, result, stderr = runTrusted(t, root, base, "guide.md")
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s", stderr))
	c.Assert(result.Suppressions[0].TrustState, qt.Equals, "trusted")
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
}

func TestTrustedModeRejectsOverridesAndDirtyPolicyEvenWithoutGate(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*testing.T, string)
		extra []string
	}{
		{"dirty config", func(t *testing.T, root string) { writeChanged(t, root, ".unswell.yaml", "invalid: [\n") }, nil},
		{"missing config", func(t *testing.T, root string) {
			qt.New(t).Assert(os.Remove(filepath.Join(root, ".unswell.yaml")), qt.IsNil)
		}, nil},
		{"index", dirtyChangedIndex, nil},
		{"profile", nil, []string{"--profile", "technical"}},
		{"gate", nil, []string{"--gate-mode", "all"}},
		{"outside", nil, []string{"--allow-config-outside-root"}},
		{"missing base resource", nil, []string{"--config", "missing.yaml"}},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			root, base := trustRepository(t, nil)
			if row.setup != nil {
				row.setup(t, root)
			}
			code, _, stderr := runTrusted(t, root, base, append(row.extra, "guide.md", "--no-gate")...)
			qt.New(t).Assert(code, qt.Equals, 2, qt.Commentf("%s", stderr))
		})
	}
}
