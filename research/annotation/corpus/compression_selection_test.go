package corpus_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func TestCompressionSelectionRequiresExplicitSettings(t *testing.T) {
	c := qt.New(t)
	options := corpus.CompressionBankOptions{Version: corpus.CompressionBankVersion, Kind: "paragraph",
		Compression: feature.CompressionOptions{Level: 0, MaxInputBytes: 10000},
		Cohorts:     []corpus.CompressionCohort{{ID: "human", Origin: "human", UnitIDs: []string{"u1"}}}}
	encodedOptions := encoded(c, options)
	loaded, err := corpus.LoadCompressionBankOptions(t.Context(), encodedOptions)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, options)
	for _, row := range []struct{ name, old, replacement string }{
		{"missing level", `"level":0,`, ``},
		{"unknown setting", `"level":0`, `"level":0,"unknown":true`},
		{"duplicate key", `"level":0`, `"level":0,"level":0`},
		{"null simulation", `"allow_simulation":false`, `"allow_simulation":null`},
		{"unknown version", corpus.CompressionBankVersion, "unswell-compression-bank-v0"},
		{"duplicate target", `"unit_ids":["u1"]`, `"unit_ids":["u1","u1"]`},
		{"unknown origin", `"origin":"human"`, `"origin":"unknown"`},
		{"budget", `"max_input_bytes":10000`, `"max_input_bytes":1048577`},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			changed := strings.Replace(string(encodedOptions), row.old, row.replacement, 1)
			c.Assert(changed, qt.Not(qt.Equals), string(encodedOptions))
			result, err := corpus.LoadCompressionBankOptions(t.Context(), []byte(changed))
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, corpus.CompressionBankOptions{})
		})
	}
}

func TestCompressionBankRejectsMetadataExpansion(t *testing.T) {
	c := qt.New(t)
	manifest, files := compressionManifest()
	var paragraphs []string
	for i := range 40 {
		paragraphs = append(paragraphs, fmt.Sprintf("Cache block %d.", i))
	}
	files["readme.md"] = []byte(strings.Join(paragraphs, "\n\n"))
	manifest.Sources[0].SHA256, manifest.Sources[0].Bytes = hash(files["readme.md"]), len(files["readme.md"])
	artifact := compressionArtifact(t, manifest, files)
	var units []annotation.Unit
	var ids []string
	for _, target := range artifact.Units {
		if target.SourceID == "d0" && target.Unit.Kind == "paragraph" {
			unit := target.Unit
			unit.Origin = annotation.Origin{Label: "human", Scope: "unit", Evidence: strings.Repeat("simulation evidence ", 3000)}
			units = append(units, unit)
			ids = append(ids, unit.ID)
		}
	}
	c.Assert(units, qt.HasLen, 40)
	round := candidateRound(t, units)
	options := corpus.CompressionBankOptions{Version: corpus.CompressionBankVersion, Kind: "paragraph", AllowSimulation: true,
		Compression: feature.CompressionOptions{Level: 0, MaxInputBytes: 1 << 20}}
	for i := range 8 {
		options.Cohorts = append(options.Cohorts, corpus.CompressionCohort{ID: fmt.Sprint(i), Origin: "human", UnitIDs: ids})
	}
	bank, err := corpus.BuildCompressionBank(t.Context(), artifact, round, files, options)
	c.Assert(err, qt.ErrorMatches, ".*metadata budget.*")
	c.Assert(bank, qt.DeepEquals, corpus.CompressionBank{})
}

func TestCompressionSelectionBoundsPublicGoInputs(t *testing.T) {
	c := qt.New(t)
	manifest, files := compressionManifest()
	artifact := compressionArtifact(t, manifest, files)
	_, base := compressionRound(t, artifact)
	for _, edit := range []func(*corpus.CompressionBankOptions){
		func(o *corpus.CompressionBankOptions) { o.Cohorts = nil },
		func(o *corpus.CompressionBankOptions) { o.Cohorts = append(o.Cohorts, o.Cohorts[0]) },
		func(o *corpus.CompressionBankOptions) { o.Cohorts[0].UnitIDs = nil },
		func(o *corpus.CompressionBankOptions) { o.Cohorts[0].UnitIDs = []string{"same", "same"} },
		func(o *corpus.CompressionBankOptions) { o.Cohorts[0].ID = "\x00" },
		func(o *corpus.CompressionBankOptions) { o.Cohorts[0].UnitIDs = []string{strings.Repeat("x", 129)} },
	} {
		var changed corpus.CompressionBankOptions
		c.Assert(json.Unmarshal(encoded(c, base), &changed), qt.IsNil)
		edit(&changed)
		c.Assert(changed.Validate(), qt.IsNotNil)
	}
}
