package corpus_test

import (
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// A bank under cohort labels seeds historical and contemporary cohorts from
// resolved snapshot labels, refuses a unit of the other class, and restores
// from its own bytes with the same digest.
func TestCompressionBankFromCohortDecisions(t *testing.T) {
	c := qt.New(t)
	manifest, files := provenanceSample()
	for i := range manifest.Sources {
		manifest.Sources[i].Partition = "training"
	}
	artifact := compressionArtifact(t, manifest, files)
	decisions, err := corpus.CohortDecisions(t.Context(), artifact)
	c.Assert(err, qt.IsNil)
	seeds := map[string]string{}
	for _, target := range artifact.Units {
		if _, taken := seeds[target.SourceID]; !taken {
			seeds[target.SourceID] = target.Unit.ID
		}
	}
	options := corpus.CompressionBankOptions{Version: corpus.CompressionBankVersion, Kind: "paragraph",
		Compression: feature.CompressionOptions{Level: 0, MaxInputBytes: 1 << 20}, Cohorts: []corpus.CompressionCohort{
			{ID: "historical", Origin: "historical", UnitIDs: []string{seeds["s2"]}},
			{ID: "contemporary", Origin: "contemporary", UnitIDs: []string{seeds["s3"]}},
			{ID: "mixed", Origin: "mixed", UnitIDs: []string{seeds["s3"], seeds["s2"]}},
		}}
	bank, err := corpus.BuildCompressionBankDecisions(t.Context(), artifact, decisions, files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(bank.Basis, qt.Equals, "declared_provenance")
	c.Assert(bank.RoundSHA256, qt.Equals, artifact.SHA256)
	c.Assert(bank.OriginsSHA256, qt.Equals, decisions.SHA256)
	c.Assert(bank.Cohorts, qt.HasLen, 3)
	c.Assert(bank.Cohorts[0].Units[0].Label, qt.Equals, "historical")
	c.Assert(bank.Cohorts[0].Units[0].Origin.Label, qt.Equals, "unknown")
	c.Assert(bank.Cohorts[1].Units[0].Label, qt.Equals, "contemporary")
	c.Assert(bank.Cohorts[2].Reference, qt.Equals, bank.Cohorts[1].Reference+"\n"+bank.Cohorts[0].Reference)
	c.Assert(bank.ReservedGroups, qt.HasLen, 1)

	encoded, err := json.Marshal(bank)
	c.Assert(err, qt.IsNil)
	restored, err := corpus.LoadCompressionBank(t.Context(), encoded)
	c.Assert(err, qt.IsNil)
	c.Assert(restored, qt.DeepEquals, bank)
	restored.Cohorts[0].Reference += "x"
	encoded, err = json.Marshal(restored)
	c.Assert(err, qt.IsNil)
	_, err = corpus.LoadCompressionBank(t.Context(), encoded)
	c.Assert(err, qt.ErrorMatches, ".*digest, version, or status mismatch")

	for _, row := range []struct{ name, origin, unit, message string }{
		{"wrong-class", "historical", seeds["s3"], ".*does not match cohort origin historical"},
		{"controlled", "contemporary", seeds["s0"], ".*requires a resolved historical or contemporary label"},
		{"round-policy", "human", seeds["s2"], ".*does not match cohort origin human"},
		{"one-sided-mixed", "mixed", seeds["s2"], ".*requires both historical and contemporary targets"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			options.Cohorts = []corpus.CompressionCohort{{ID: "x", Origin: row.origin, UnitIDs: []string{row.unit}}}
			_, err := corpus.BuildCompressionBankDecisions(t.Context(), artifact, decisions, files, options)
			c.Assert(err, qt.ErrorMatches, row.message)
		})
	}
}
