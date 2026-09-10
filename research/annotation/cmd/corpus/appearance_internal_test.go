package main

// White-box tests: The first-appearance command and the analyze filter flags read candidate artifacts and
// filters from explicit local files; package main exposes no importable API for that wiring.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

// cohortCandidates plans and extracts the Ptah fixture under one cohort and
// writes the candidate artifact next to the returned path.
func cohortCandidates(t *testing.T, root, cohort, date string, edit func(*corpus.Manifest)) string {
	t.Helper()
	c := qt.New(t)
	var manifest corpus.Manifest
	c.Assert(json.Unmarshal(manifestBytes(c), &manifest), qt.IsNil)
	for i := range manifest.Sources {
		manifest.Sources[i].Snapshot = &corpus.Snapshot{Date: date, Confidence: "corroborated", Evidence: "Fixture", Cohort: cohort}
	}
	if edit != nil {
		edit(&manifest)
	}
	encoded, err := json.Marshal(manifest)
	c.Assert(err, qt.IsNil)
	var plan, artifact bytes.Buffer
	c.Assert(run(t.Context(), []string{"plan"}, bytes.NewReader(encoded), &plan), qt.IsNil)
	c.Assert(run(t.Context(), []string{"extract", "--root", root}, &plan, &artifact), qt.IsNil)
	path := filepath.Join(t.TempDir(), cohort+".json")
	c.Assert(os.WriteFile(path, artifact.Bytes(), 0o600), qt.IsNil)
	return path
}

// laterRoot copies the fixture and appends a paragraph to one document, so
// the later cohort repeats every other unit.
func laterRoot(t *testing.T) (string, func(*corpus.Manifest)) {
	t.Helper()
	c := qt.New(t)
	root := t.TempDir()
	c.Assert(os.CopyFS(root, os.DirFS("../../corpus/testdata/ptah")), qt.IsNil)
	target := filepath.Join(root, "docs", "sequences.md")
	// #nosec G304 -- The test reads the fixture copy it just made under its temporary directory.
	data, err := os.ReadFile(target)
	c.Assert(err, qt.IsNil)
	data = append(data, []byte("\nA later snapshot adds this paragraph about sequence ownership.\n")...)
	// #nosec G703 -- The test writes into the fixture copy it just made under its temporary directory.
	c.Assert(os.WriteFile(target, data, 0o600), qt.IsNil)
	return root, func(m *corpus.Manifest) {
		for i := range m.Sources {
			if m.Sources[i].Path == "docs/sequences.md" {
				m.Sources[i].SHA256, m.Sources[i].Bytes = sha(data), len(data)
			}
		}
	}
}

func measured(t *testing.T, root, candidates string) string {
	t.Helper()
	c := qt.New(t)
	dir := t.TempDir()
	policy := filepath.Join(dir, "policy.yaml")
	c.Assert(os.WriteFile(policy, []byte("version: 1\nextends: [builtin:technical-v1]\n"), 0o600), qt.IsNil)
	// #nosec G304 -- The test reads the artifact it wrote under its temporary directory.
	data, err := os.ReadFile(candidates)
	c.Assert(err, qt.IsNil)
	var findings bytes.Buffer
	c.Assert(run(t.Context(), []string{"measure", "--root", root, "--policy", policy}, bytes.NewReader(data), &findings), qt.IsNil)
	path := filepath.Join(dir, "findings.json")
	c.Assert(os.WriteFile(path, findings.Bytes(), 0o600), qt.IsNil)
	return path
}

func TestFirstAppearanceCommandFeedsTheUnitAnalysis(t *testing.T) {
	c := qt.New(t)
	earlierRoot := "../../corpus/testdata/ptah"
	earlier := cohortCandidates(t, earlierRoot, "historical", "2019-06-30", nil)
	root, edit := laterRoot(t)
	later := cohortCandidates(t, root, "contemporary", "2024-01-15", edit)
	var output bytes.Buffer
	c.Assert(run(t.Context(), []string{"first-appearance", "--unit-kind", "paragraph", "--earlier", earlier, "--later", later},
		nil, &output), qt.IsNil)
	filter, err := corpus.LoadAppearanceFilter(t.Context(), output.Bytes())
	c.Assert(err, qt.IsNil)
	c.Assert(filter.NewUnits, qt.Equals, 1)
	c.Assert(filter.RepeatedUnits, qt.Equals, filter.EarlierUnits)
	c.Assert(filter.LaterOnlyUnits, qt.Equals, 0)
	c.Assert(filter.Repositories, qt.HasLen, 1)
	filterPath := filepath.Join(t.TempDir(), "filter.json")
	c.Assert(os.WriteFile(filterPath, output.Bytes(), 0o600), qt.IsNil)
	classes := "../../../methods/rule-classes-v1.json"
	findings := []string{measured(t, earlierRoot, earlier), measured(t, root, later)}
	var tables bytes.Buffer
	c.Assert(run(t.Context(), []string{"analyze", "--classes", classes, "--findings", findings[0], "--findings", findings[1],
		"--unit-kind", "paragraph", "--first-appearance", filterPath}, nil, &tables), qt.IsNil)
	var decoded patterns.Tables
	c.Assert(json.Unmarshal(tables.Bytes(), &decoded), qt.IsNil)
	c.Assert(decoded.Unit, qt.Equals, "paragraph")
	c.Assert(decoded.FirstAppearance.Kept, qt.Equals, 1)
	c.Assert(decoded.FirstAppearance.Excluded, qt.Equals, filter.RepeatedUnits)
	c.Assert(decoded.Cohorts, qt.HasLen, 2)
	c.Assert(decoded.Cohorts[0].Cohort, qt.Equals, "contemporary")
	c.Assert(decoded.Cohorts[0].Units, qt.Equals, 1)
	c.Assert(decoded.Cohorts[0].Excluded, qt.Equals, filter.RepeatedUnits)
	c.Assert(decoded.Cohorts[1].Units, qt.Equals, filter.EarlierUnits)
	// The same cohort on both sides, a missing kind, and a filter without a
	// unit analysis are refused before any table is built.
	c.Assert(run(t.Context(), []string{"first-appearance", "--unit-kind", "paragraph", "--earlier", earlier, "--later", earlier},
		nil, &output), qt.IsNotNil)
	for _, args := range [][]string{{"first-appearance"}, {"first-appearance", "--unit-kind", "paragraph", "--earlier", earlier},
		{"first-appearance", "--earlier", earlier, "--later", later}, {"first-appearance", "--unit-kind", "paragraph",
			"--earlier", earlier, "--later", later, "extra"}} {
		c.Assert(run(t.Context(), args, nil, &output), qt.IsNotNil, qt.Commentf("%v", args))
	}
	c.Assert(run(t.Context(), []string{"analyze", "--classes", classes, "--findings", findings[0], "--first-appearance", filterPath},
		nil, &tables), qt.IsNotNil)
	c.Assert(run(t.Context(), []string{"analyze", "--classes", classes, "--findings", findings[0], "--unit-kind", "paragraph",
		"--first-appearance", filepath.Join(t.TempDir(), "missing.json")}, nil, &tables), qt.IsNotNil)
}
