package main

// White-box tests: The analyze command reads finding artifacts and rule classes from explicit local files;
// package main exposes no importable API for that wiring or for its argument checks.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/patterns"
)

func TestAnalyzeCommandBuildsTablesFromMeasuredFindings(t *testing.T) {
	c := qt.New(t)
	var plan, artifact, measured bytes.Buffer
	c.Assert(run(t.Context(), []string{"plan"}, bytes.NewReader(manifestBytes(c)), &plan), qt.IsNil)
	root := "../../corpus/testdata/ptah"
	c.Assert(run(t.Context(), []string{"extract", "--root", root}, &plan, &artifact), qt.IsNil)
	dir := t.TempDir()
	policy := filepath.Join(dir, "policy.yaml")
	c.Assert(os.WriteFile(policy, []byte("version: 1\nextends: [builtin:technical-v1]\n"), 0o600), qt.IsNil)
	c.Assert(run(t.Context(), []string{"measure", "--root", root, "--policy", policy}, &artifact, &measured), qt.IsNil)
	findings := filepath.Join(dir, "findings.json")
	c.Assert(os.WriteFile(findings, measured.Bytes(), 0o600), qt.IsNil)
	classes := "../../../methods/rule-classes-v1.json"
	var tables bytes.Buffer
	c.Assert(run(t.Context(), []string{"analyze", "--classes", classes, "--findings", findings}, nil, &tables), qt.IsNil)
	var decoded patterns.Tables
	c.Assert(json.Unmarshal(tables.Bytes(), &decoded), qt.IsNil)
	c.Assert(decoded.Version, qt.Equals, patterns.Version)
	c.Assert(decoded.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(decoded.Rules, qt.HasLen, 40)
	// The Ptah fixture declares no snapshot, so every document stays unassigned.
	c.Assert(decoded.Unassigned, qt.Equals, 8)
	c.Assert(decoded.Cohorts, qt.HasLen, 0)
	c.Assert(decoded.Inputs, qt.HasLen, 1)
	c.Assert(decoded.Inputs[0].Documents, qt.Equals, 8)
	// An edited artifact fails its digest; bad arguments fail before any read.
	edited := bytes.Replace(measured.Bytes(), []byte(`"prose_words": `), []byte(`"prose_words": 1`), 1)
	c.Assert(os.WriteFile(findings, edited, 0o600), qt.IsNil)
	c.Assert(run(t.Context(), []string{"analyze", "--classes", classes, "--findings", findings}, nil, &tables), qt.IsNotNil)
	for _, args := range [][]string{{"analyze"}, {"analyze", "--classes", classes}, {"analyze", "--findings", findings},
		{"analyze", "--classes", classes, "--findings", findings, "extra"}} {
		c.Assert(run(t.Context(), args, nil, &tables), qt.IsNotNil, qt.Commentf("%v", args))
	}
}
