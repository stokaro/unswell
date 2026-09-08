package e2e_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

type expectation struct {
	path       string
	line       int
	rule       string
	suppressed bool
}

func prepareSources(t *testing.T, fixture, workspace string, spec scenario) (map[string][]byte, []expectation) {
	t.Helper()
	c := qt.New(t)
	sources := make(map[string][]byte)
	var wants []expectation
	for _, name := range spec.Files {
		c.Assert(fs.ValidPath(name), qt.IsTrue)
		data, err := fs.ReadFile(os.DirFS(fixture), name+".txt")
		c.Assert(err, qt.IsNil)
		clean, annotations, err := annotatedSource(name, string(data))
		c.Assert(err, qt.IsNil)
		if spec.CRLF {
			clean = strings.ReplaceAll(clean, "\n", "\r\n")
		}
		if spec.BOM {
			clean = "\ufeff" + clean
		}
		sources[name] = []byte(clean)
		wants = append(wants, annotations...)
		writeFixture(c, workspace, name, sources[name])
	}
	return sources, wants
}

func prepareResources(t *testing.T, fixture, workspace string, names []string, sources map[string][]byte) {
	t.Helper()
	c := qt.New(t)
	for _, name := range names {
		c.Assert(fs.ValidPath(name), qt.IsTrue)
		_, exists := sources[name]
		c.Assert(exists || name == "policy.yaml", qt.IsFalse)
		data, err := fs.ReadFile(os.DirFS(fixture), name)
		c.Assert(err, qt.IsNil)
		sources[name] = data
		writeFixture(c, workspace, name, data)
	}
}

func writeFixture(c *qt.C, workspace, name string, data []byte) {
	c.Helper()
	c.Assert(fs.ValidPath(name), qt.IsTrue)
	c.Assert(strings.ContainsAny(name, "\\:"), qt.IsFalse)
	target := filepath.Join(workspace, filepath.FromSlash(name))
	c.Assert(os.MkdirAll(filepath.Dir(target), 0o700), qt.IsNil)
	root, err := os.OpenRoot(workspace)
	c.Assert(err, qt.IsNil)
	writeErr := root.WriteFile(filepath.FromSlash(name), data, 0o600)
	closeErr := root.Close()
	c.Assert(writeErr, qt.IsNil)
	c.Assert(closeErr, qt.IsNil)
}

func annotatedSource(name, source string) (string, []expectation, error) {
	marker := regexp.MustCompile(`(?:\/\/|#|<!--) want(-suppressed)? (.*?)(?: -->)?$`)
	quoted := regexp.MustCompile(`"[a-z][a-z0-9.-]+"`)
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	var wants []expectation
	for i, line := range lines {
		match := marker.FindStringSubmatchIndex(line)
		if match == nil {
			continue
		}
		value := line[match[4]:match[5]]
		rules := quoted.FindAllString(value, -1)
		if len(rules) == 0 || strings.TrimSpace(quoted.ReplaceAllString(value, "")) != "" {
			return "", nil, fmt.Errorf("%s:%d: invalid want annotation", name, i+1)
		}
		for _, rule := range rules {
			id, err := strconv.Unquote(rule)
			if err != nil {
				return "", nil, err
			}
			wants = append(wants, expectation{name, i + 1, id, match[2] >= 0})
		}
		// Keep source offsets stable while excluding test annotations from prose analysis.
		lines[i] = line[:match[0]] + strings.Repeat(" ", len(line)-match[0])
	}
	return strings.Join(lines, "\n"), wants, nil
}

func matchWants(findings []unswell.Finding, expected []expectation) error {
	remaining := slices.Clone(expected)
	for _, finding := range findings {
		actual := expectation{finding.Primary.Path, finding.Primary.Start.Line, finding.RuleID, finding.Suppressed}
		index := slices.Index(remaining, actual)
		if index < 0 {
			return fmt.Errorf("unexpected finding %s:%d: %s (suppressed=%t)", actual.path, actual.line, actual.rule, actual.suppressed)
		}
		remaining = slices.Delete(remaining, index, index+1)
	}
	if len(remaining) > 0 {
		return fmt.Errorf("missing findings: %+v", remaining)
	}
	return nil
}

func assertInputsUnchanged(t *testing.T, workspace string, sources map[string][]byte) {
	t.Helper()
	c := qt.New(t)
	for name, expected := range sources {
		actual, err := fs.ReadFile(os.DirFS(workspace), name)
		c.Assert(err, qt.IsNil)
		c.Assert(actual, qt.DeepEquals, expected)
	}
}

func verifyLocations(t *testing.T, result unswell.RunResult, sources map[string][]byte) {
	t.Helper()
	c := qt.New(t)
	for _, doc := range result.Documents {
		c.Assert(doc.Source, qt.Equals, "")
	}
	for _, finding := range result.Findings {
		locations := append([]unswell.Location{finding.Primary}, finding.Related...)
		for _, location := range locations {
			source, ok := sources[location.Path]
			c.Assert(ok, qt.IsTrue)
			c.Assert(location.Snippet, qt.Equals, "")
			c.Assert(location.Span.Start >= 0 && location.Span.End > location.Span.Start, qt.IsTrue)
			c.Assert(location.Span.End <= len(source), qt.IsTrue)
			c.Assert(location.Start, qt.Equals, position(source, location.Span.Start))
			c.Assert(location.End, qt.Equals, position(source, location.Span.End))
		}
	}
}

func position(source []byte, offset int) document.Position {
	before := string(source[:offset])
	return document.Position{Line: strings.Count(before, "\n") + 1,
		Column: utf8.RuneCountInString(before[strings.LastIndexByte(before, '\n')+1:]) + 1}
}

func TestExpectationsRejectIncorrectDetections(t *testing.T) {
	want := []expectation{{"sample.go", 2, "filler.announced-importance", false}}
	finding := unswell.Finding{RuleID: want[0].rule,
		Primary: unswell.Location{Path: "sample.go", Start: document.Position{Line: 2, Column: 4}}}
	c := qt.New(t)
	c.Assert(matchWants([]unswell.Finding{finding}, want), qt.IsNil)
	c.Assert(matchWants(nil, want), qt.ErrorMatches, "missing findings:.*")
	c.Assert(matchWants([]unswell.Finding{finding}, nil), qt.ErrorMatches, "unexpected finding.*")
	c.Assert(matchWants([]unswell.Finding{finding, finding}, want), qt.ErrorMatches, "unexpected finding.*")
	finding.Suppressed = true
	c.Assert(matchWants([]unswell.Finding{finding}, want), qt.ErrorMatches, "unexpected finding.*")
	want[0].suppressed = true
	c.Assert(matchWants([]unswell.Finding{finding}, want), qt.IsNil)
	finding.Primary.Start.Line = 3
	c.Assert(matchWants([]unswell.Finding{finding}, want), qt.ErrorMatches, "unexpected finding.*")
	_, _, err := annotatedSource("sample.go", "// want broken")
	c.Assert(err, qt.ErrorMatches, ".*invalid want annotation")
}
