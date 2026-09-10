package main

// White-box tests: The duplicates command reads pinned sources through the process root and writes
// either a report or an applied manifest; package main exposes no importable API for that wiring.

import (
	"bytes"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

func TestDuplicatesCommandReportsAndApplies(t *testing.T) {
	c := qt.New(t)
	root := "../../corpus/testdata/ptah"
	var report bytes.Buffer
	c.Assert(run(t.Context(), []string{"duplicates", "--root", root}, bytes.NewReader(manifestBytes(c)), &report), qt.IsNil)
	var decoded corpus.DuplicateReport
	c.Assert(json.Unmarshal(report.Bytes(), &decoded), qt.IsNil)
	c.Assert(decoded.Version, qt.Equals, corpus.DuplicatesVersion)
	c.Assert(decoded.Sources, qt.Equals, 8)
	c.Assert(decoded.Threshold, qt.Equals, corpus.DefaultDuplicateThreshold)
	var applied bytes.Buffer
	c.Assert(run(t.Context(), []string{"duplicates", "--root", root, "--apply", "--threshold", "0.9"},
		bytes.NewReader(manifestBytes(c)), &applied), qt.IsNil)
	manifest, err := corpus.LoadManifest(t.Context(), applied.Bytes())
	c.Assert(err, qt.IsNil)
	c.Assert(manifest.Sources, qt.HasLen, 8)
	for _, args := range [][]string{{"duplicates"}, {"duplicates", "--root", root, "--threshold", "0"},
		{"duplicates", "--root", root, "extra"}} {
		c.Assert(run(t.Context(), args, bytes.NewReader(manifestBytes(c)), &applied), qt.IsNotNil, qt.Commentf("%v", args))
	}
}
