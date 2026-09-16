package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

type rhetoricCase struct {
	SourcePath   string             `json:"source_path"`
	SourceSHA256 string             `json:"source_sha256"`
	Start        int                `json:"start"`
	End          int                `json:"end"`
	ID           string             `json:"id"`
	Text         string             `json:"text"`
	SHA256       string             `json:"sha256"`
	Revision     string             `json:"revision"`
	Rationale    string             `json:"rationale"`
	Expected     []rhetoricExpected `json:"expected"`
}

type rhetoricExpected struct {
	Rule  string `json:"rule"`
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

var rhetoricIDs = []string{
	"filler.document-justification", "filler.evaluative-closure", "repetition.definition-echo", "syntax.slogan-contrast",
}

// These are development judgments tied to real source, not independent labels
// or an estimate of precision. Both profiles must expose the same new warnings.
func TestPtahRhetoric(t *testing.T) {
	c := qt.New(t)
	var book struct {
		Version    int            `json:"version"`
		Repository string         `json:"repository"`
		Commit     string         `json:"commit"`
		Annotation string         `json:"annotation"`
		Cases      []rhetoricCase `json:"cases"`
	}
	decodeFile(t, "rhetoricdata/ptah.json", &book)
	c.Assert(book.Cases, qt.HasLen, 21)
	c.Assert(book.Version, qt.Equals, 1)
	c.Assert(book.Commit, qt.Equals, "654eae5591392278e6c8bce8e54737f780766f19")
	binary := buildCLI(t)
	for _, profile := range []string{"technical", "strict"} {
		t.Run(profile, func(t *testing.T) {
			runRhetoricCases(t, binary, profile, book.Cases)
		})
	}
}

func runRhetoricCases(t *testing.T, binary, profile string, cases []rhetoricCase) {
	t.Helper()
	c := qt.New(t)
	workspace := t.TempDir()
	args := []string{"check", "--profile", profile, "--include-source", "--report", "json:result.json"}
	wants := make(map[string][]rhetoricExpected)
	for _, row := range cases {
		digest := sha256.Sum256([]byte(row.Text))
		c.Assert(hex.EncodeToString(digest[:]), qt.Equals, row.SHA256, qt.Commentf("%s", row.ID))
		c.Assert(row.Rationale, qt.Not(qt.Equals), "")
		c.Assert(row.End-row.Start, qt.Equals, len(row.Text))
		for _, expected := range row.Expected {
			c.Assert(expected.Start >= 0 && expected.End <= len(row.Text) && expected.Start < expected.End, qt.IsTrue)
			c.Assert(row.Text[expected.Start:expected.End], qt.Equals, expected.Text)
		}
		for name, text := range map[string]string{row.ID: row.Text, row.ID + "-revised": row.Revision} {
			if text == "" {
				continue
			}
			name += ".md"
			// #nosec G703 -- Names come from the checked-in casebook, within t.TempDir.
			c.Assert(os.WriteFile(filepath.Join(workspace, name), []byte(text), 0o600), qt.IsNil)
			args = append(args, name)
			wants[name] = nil
		}
		wants[row.ID+".md"] = append([]rhetoricExpected(nil), row.Expected...)
	}
	stdout, stderr, code := invoke(t, binary, workspace, args, nil)
	c.Assert(code == 0 || code == 1, qt.IsTrue, qt.Commentf("%s\n%s", stdout, stderr))
	var result unswell.RunResult
	decodeFile(t, filepath.Join(workspace, "result.json"), &result)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	assertRhetoricFindings(t, result, wants)
}

func assertRhetoricFindings(t *testing.T, result unswell.RunResult, wants map[string][]rhetoricExpected) {
	t.Helper()
	c := qt.New(t)
	got := make(map[string][]rhetoricExpected)
	for name := range wants {
		got[name] = nil
	}
	for _, finding := range result.Findings {
		if !slices.Contains(rhetoricIDs, finding.RuleID) {
			continue
		}
		c.Assert(finding.Severity, qt.Equals, "warning")
		c.Assert(finding.Evidence.Suggestion, qt.Not(qt.Equals), "")
		c.Assert(finding.Related, qt.HasLen, 0)
		name := finding.Primary.Path
		got[name] = append(got[name], rhetoricExpected{finding.RuleID, finding.Primary.Snippet,
			finding.Primary.Span.Start, finding.Primary.Span.End})
	}
	c.Assert(got, qt.DeepEquals, wants)
}
