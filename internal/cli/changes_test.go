package cli_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/internal/cli"
)

func commitGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	argv := append([]string{"-C", root, "-c", "commit.gpgsign=false", "-c", "core.autocrlf=false",
		"-c", "user.name=Unswell Tests", "-c", "user.email=tests@example.invalid"}, args...)
	// #nosec G204 -- Tests use fixed Git arguments in a test-owned repository.
	command := exec.CommandContext(t.Context(), "git", argv...)
	output, err := command.CombinedOutput()
	qt.New(t).Assert(err, qt.IsNil, qt.Commentf("git %v: %s", args, output))
	return strings.TrimSpace(string(output))
}

func writeChanged(t *testing.T, root, name, value string) {
	t.Helper()
	c := qt.New(t)
	handle, err := os.OpenRoot(root)
	c.Assert(err, qt.IsNil)
	c.Assert(handle.WriteFile(name, []byte(value), 0o600), qt.IsNil)
	c.Assert(handle.Close(), qt.IsNil)
}

func changedRepository(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	commitGit(t, root, "init", "-b", "main")
	writeChanged(t, root, "guide.md", "# Client\n\nCertainly! The client retries.\n")
	writeChanged(t, root, ".unswell.yaml", "version: 1\nrules:\n  scaffold.chat-preamble:\n    parameters: {positions: [sentence-start]}\n")
	commitGit(t, root, "add", ".")
	commitGit(t, root, "commit", "-m", "Initial prose")
	return root, commitGit(t, root, "rev-parse", "HEAD")
}

func runChanged(t *testing.T, root, base string, extra ...string) (int, unswell.RunResult, string) {
	t.Helper()
	var out, stderr bytes.Buffer
	args := append([]string{"check", "--changed-from", base, "--report", "json:-"}, extra...)
	code := cli.Run(t.Context(), args, cli.Environment{Dir: root, In: bytes.NewReader(nil), Out: &out, Err: &stderr})
	var result unswell.RunResult
	if out.Len() > 0 {
		qt.New(t).Assert(json.Unmarshal(out.Bytes(), &result), qt.IsNil, qt.Commentf("%s", out.String()))
	}
	return code, result, stderr.String()
}

func TestCommittedChangesUseMergeBaseAndPreserveAllFindings(t *testing.T) {
	c := qt.New(t)
	root, base := changedRepository(t)
	commitGit(t, root, "checkout", "-b", "topic")
	writeChanged(t, root, "guide.md", "A separate introduction.\n\n# Client\n\nCertainly! The client retries.\n")
	commitGit(t, root, "commit", "-am", "Move existing prose")
	head := commitGit(t, root, "rev-parse", "HEAD")
	commitGit(t, root, "checkout", "main")
	writeChanged(t, root, "guide.md", "The upstream branch replaces this document.\n")
	commitGit(t, root, "commit", "-am", "Advance upstream")
	commitGit(t, root, "checkout", "topic")
	code, result, stderr := runChanged(t, root, "main")
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s", stderr))
	c.Assert(result.Manifest.Git.BaseCommit, qt.Equals, base)
	c.Assert(result.Manifest.Git.HeadCommit, qt.Equals, head)
	c.Assert(result.Manifest.Git.Clean, qt.IsTrue)
	c.Assert(result.Changes.Complete, qt.IsTrue)
	c.Assert(len(result.Findings) > 0, qt.IsTrue)
	c.Assert(len(result.Gate.Unchanged) > 0, qt.IsTrue)
	writeChanged(t, root, "guide.md", "# Client\n\nCertainly! The client retries after a timeout.\n")
	commitGit(t, root, "commit", "-am", "Change the full paragraph")
	code, result, stderr = runChanged(t, root, "main")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	c.Assert(len(result.Gate.Reasons) > 0, qt.IsTrue)
}

func TestCommittedChangesRejectDirtyAndUnavailableInputs(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, string)
		extra  []string
		want   string
	}{
		{"worktree", func(t *testing.T, root string) { writeChanged(t, root, "guide.md", "Clean local draft.\n") },
			nil, "bytes do not match HEAD"},
		{"index", dirtyChangedIndex, nil, "index does not match HEAD"},
		{"assume unchanged", func(t *testing.T, root string) {
			commitGit(t, root, "update-index", "--assume-unchanged", "guide.md")
			writeChanged(t, root, "guide.md", "Clean local draft.\n")
		}, nil, "bytes do not match HEAD"},
		{"staged new file", func(t *testing.T, root string) {
			writeChanged(t, root, "new.md", "Clean local draft.\n")
			commitGit(t, root, "add", "new.md")
		}, nil, "index path is absent"},
		{"untracked explicit", func(t *testing.T, root string) { writeChanged(t, root, "new.md", "Local draft.\n") },
			[]string{"new.md"}, "absent from both committed trees"},
		{"dirty policy", func(t *testing.T, root string) {
			writeChanged(t, root, ".unswell.yaml", "version: 1\n# Local policy.\n")
		},
			nil, "bytes do not match HEAD"},
		{"committed policy", func(t *testing.T, root string) {
			writeChanged(t, root, ".unswell.yaml", "version: 1\n# Changed policy.\n")
			commitGit(t, root, "commit", "-am", "Change policy")
		}, nil, "policy input changed"},
		{"removed policy", func(t *testing.T, root string) {
			commitGit(t, root, "rm", ".unswell.yaml")
			commitGit(t, root, "commit", "-m", "Remove policy")
		}, nil, "policy input changed"},
		{"locally removed discovered policy", func(t *testing.T, root string) {
			qt.New(t).Assert(os.Remove(filepath.Join(root, ".unswell.yaml")), qt.IsNil)
		}, []string{"guide.md"}, "missing committed policy input"},
		{"stdin", func(_ *testing.T, _ string) {}, []string{"--stdin", "--filename", "guide.md"}, "requires committed source paths"},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			root, base := changedRepository(t)
			row.change(t, root)
			code, _, stderr := runChanged(t, root, base, append(row.extra, "--no-gate")...)
			c.Assert(code, qt.Equals, 2, qt.Commentf("%s", stderr))
			c.Assert(stderr, qt.Contains, row.want)
		})
	}
}

func dirtyChangedIndex(t *testing.T, root string) {
	t.Helper()
	writeChanged(t, root, "guide.md", "Clean staged draft.\n")
	commitGit(t, root, "add", "guide.md")
	writeChanged(t, root, "guide.md", "# Client\n\nCertainly! The client retries.\n")
}

func TestCommittedChangesDeletionAndExplicitEmptyPermission(t *testing.T) {
	c := qt.New(t)
	root, base := changedRepository(t)
	commitGit(t, root, "rm", "guide.md")
	commitGit(t, root, "commit", "-m", "Delete selected document")
	code, result, _ := runChanged(t, root, base, "guide.md")
	c.Assert(code, qt.Equals, 2)
	c.Assert(result.Manifest.Complete, qt.IsFalse)
	code, result, stderr := runChanged(t, root, base, "guide.md", "--allow-empty")
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s", stderr))
	c.Assert(result.Changes.Documents, qt.HasLen, 1)
	c.Assert(result.Changes.Documents[0].Status, qt.Equals, "deleted")
	c.Assert(result.Manifest.Git.Clean, qt.IsTrue)
	writeChanged(t, root, "guide.md", "Restored untracked draft.\n")
	code, _, stderr = runChanged(t, root, base, "guide.md", "--allow-empty", "--no-gate")
	c.Assert(code, qt.Equals, 2)
	c.Assert(stderr, qt.Contains, "deleted file is present")
}

func TestCommittedChangesDoNotEvaluateReferenceArgumentsAsOptions(t *testing.T) {
	c := qt.New(t)
	root, _ := changedRepository(t)
	for _, ref := range []string{"missing-reference", "--help", "; touch injected", "HEAD:guide.md"} {
		code, _, stderr := runChanged(t, root, ref, "--no-gate")
		c.Assert(code, qt.Equals, 2, qt.Commentf("%q: %s", ref, stderr))
	}
	_, err := os.Stat(filepath.Join(root, "injected"))
	c.Assert(os.IsNotExist(err), qt.IsTrue)
}

func TestCommittedChangesSupportProjectSubdirectories(t *testing.T) {
	c := qt.New(t)
	root, _ := changedRepository(t)
	c.Assert(os.Mkdir(filepath.Join(root, "docs"), 0o700), qt.IsNil)
	commitGit(t, root, "mv", "guide.md", "docs/guide.md")
	commitGit(t, root, "commit", "-m", "Move documentation")
	base := commitGit(t, root, "rev-parse", "HEAD")
	code, result, stderr := runChanged(t, filepath.Join(root, "docs"), base, "--project-root", ".")
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s", stderr))
	c.Assert(result.Documents[0].Name, qt.Equals, "guide.md")
}

func TestCommittedRenameSelectsTheNewLogicalPath(t *testing.T) {
	c := qt.New(t)
	root, base := changedRepository(t)
	commitGit(t, root, "mv", "guide.md", "renamed.md")
	commitGit(t, root, "commit", "-m", "Rename document")
	code, result, stderr := runChanged(t, root, base, "guide.md", "renamed.md")
	c.Assert(code, qt.Equals, 1, qt.Commentf("%s", stderr))
	c.Assert(result.Changes.Documents, qt.HasLen, 2)
	c.Assert(result.Changes.Documents[0].Status, qt.Equals, "deleted")
	c.Assert(result.Changes.Documents[1].Status, qt.Equals, "added")
	c.Assert(result.Changes.Documents[1].FullReason, qt.Equals, "added")
}

func TestCommittedSourcesRejectSymlinksAndCheckoutFilters(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Creating symlinks requires a Windows host privilege; byte verification is covered on every platform.")
	}
	c := qt.New(t)
	root, base := changedRepository(t)
	data, err := fs.ReadFile(os.DirFS(root), "guide.md")
	c.Assert(err, qt.IsNil)
	c.Assert(os.Rename(filepath.Join(root, "guide.md"), filepath.Join(root, "copy.md")), qt.IsNil)
	c.Assert(os.Symlink("copy.md", filepath.Join(root, "guide.md")), qt.IsNil)
	code, _, stderr := runChanged(t, root, base, "guide.md")
	c.Assert(code, qt.Equals, 2)
	c.Assert(stderr, qt.Contains, "does not follow symlinks")
	c.Assert(os.Remove(filepath.Join(root, "guide.md")), qt.IsNil)
	writeChanged(t, root, "guide.md", strings.ReplaceAll(string(data), "\n", "\r\n"))
	code, _, stderr = runChanged(t, root, base, "guide.md")
	c.Assert(code, qt.Equals, 2)
	c.Assert(stderr, qt.Contains, "bytes do not match HEAD")
}

func TestCommittedPolicyDirectoryCannotLoseALocalRuleFile(t *testing.T) {
	c := qt.New(t)
	root, _ := changedRepository(t)
	data, err := os.ReadFile("../../examples/rules/company.yaml")
	c.Assert(err, qt.IsNil)
	c.Assert(os.Mkdir(filepath.Join(root, "rules"), 0o700), qt.IsNil)
	writeChanged(t, root, "rules/first.yaml", string(data))
	second := strings.ReplaceAll(string(data), "company", "second")
	writeChanged(t, root, "rules/second.yaml", second)
	commitGit(t, root, "add", "rules")
	commitGit(t, root, "commit", "-m", "Add ruleset directory")
	base := commitGit(t, root, "rev-parse", "HEAD")
	code, _, stderr := runChanged(t, root, base, "guide.md", "--ruleset", "rules")
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s", stderr))
	c.Assert(os.Remove(filepath.Join(root, "rules", "first.yaml")), qt.IsNil)
	code, _, stderr = runChanged(t, root, base, "guide.md", "--ruleset", "rules", "--no-gate")
	c.Assert(code, qt.Equals, 2)
	c.Assert(stderr, qt.Contains, "missing committed policy input")
}
