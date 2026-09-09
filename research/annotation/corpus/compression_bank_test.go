package corpus_test

import (
	"fmt"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func compressionManifest() (corpus.Manifest, map[string][]byte) {
	m, files := sample()
	for i := range m.Sources {
		m.Sources[i].Partition = "training"
		m.Sources[i].Rights.AllowedUses = []string{"annotation", "training"}
	}
	for i, text := range []string{"Service retries after a transport failure.", "Keep the error condition.",
		"A held-out paragraph.", "```go\nconst X = 1\n```\n"} {
		source := m.Sources[0]
		source.ID, source.Path = fmt.Sprintf("d%d", i+2), fmt.Sprintf("extra%d.md", i)
		source.Document, source.Repository = source.Path, source.ID
		source.Reference, source.SHA256, source.Bytes = "fixture:"+source.Path, hash([]byte(text)), len(text)
		if i == 2 {
			source.Partition = "development"
		}
		files[source.Path] = []byte(text)
		m.Sources = append(m.Sources, source)
	}
	m.Sources[1].Related, m.Sources[5].Related = []string{"transitive-bridge"}, []string{"transitive-bridge"}
	return m, files
}

func compressionArtifact(t *testing.T, manifest corpus.Manifest, files map[string][]byte) corpus.Artifact {
	t.Helper()
	c := qt.New(t)
	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, files)
	c.Assert(err, qt.IsNil)
	return artifact
}

func compressionRound(t *testing.T, artifact corpus.Artifact) (*annotation.Round, corpus.CompressionBankOptions) {
	t.Helper()
	c := qt.New(t)
	var seeds []annotation.Unit
	for _, source := range []string{"d0", "d2"} {
		for _, target := range artifact.Units {
			if target.SourceID != source || target.Unit.Kind != "paragraph" {
				continue
			}
			unit := target.Unit
			unit.Origin = annotation.Origin{Label: "human", Scope: "unit", Evidence: "Simulated unit provenance for a tutorial fixture."}
			if source == "d2" {
				unit.Origin.Label, unit.Origin.GenerationRecord = "generated", "fixture:simulated-generation"
			}
			seeds = append(seeds, unit)
			break
		}
	}
	c.Assert(seeds, qt.HasLen, 2)
	options := corpus.CompressionBankOptions{Version: corpus.CompressionBankVersion, Kind: "paragraph", AllowSimulation: true,
		Compression: feature.CompressionOptions{Level: 0, MaxInputBytes: 1 << 20}, Cohorts: []corpus.CompressionCohort{
			{ID: "human", Origin: "human", UnitIDs: []string{seeds[0].ID}},
			{ID: "generated", Origin: "generated", UnitIDs: []string{seeds[1].ID}},
			{ID: "mixed", Origin: "mixed", UnitIDs: []string{seeds[1].ID, seeds[0].ID}},
		}}
	return candidateRound(t, seeds), options
}

func TestCompressionBankPreservesOrderAndReservesConnectedGroups(t *testing.T) {
	c := qt.New(t)
	manifest, files := compressionManifest()
	artifact := compressionArtifact(t, manifest, files)
	round, options := compressionRound(t, artifact)
	bank, err := corpus.BuildCompressionBank(t.Context(), artifact, round, files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(bank.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(bank.Basis, qt.Equals, "simulation")
	c.Assert(bank.Cohorts, qt.HasLen, 3)
	c.Assert(bank.RemainingTrainingUnits, qt.Equals, 1)
	c.Assert(bank.ReservedGroups, qt.HasLen, 2)
	c.Assert(bank.Verification.ArtifactSHA256, qt.Equals, artifact.SHA256)
	mixed := bank.Cohorts[2]
	c.Assert(mixed.Reference, qt.Equals, bank.Cohorts[1].Reference+"\n"+bank.Cohorts[0].Reference)
	c.Assert(mixed.Identity.ReferenceBytes, qt.Equals, len(mixed.Reference)+1)
	c.Assert(mixed.Identity.ReferenceSHA256, qt.Equals, hash([]byte(mixed.Reference+"\n")))
	for _, unit := range mixed.Units {
		c.Assert(unit.Binding.TextSHA256, qt.Equals, hash([]byte(mixed.Reference[unit.Start:unit.End])))
		c.Assert(unit.Source.Segments, qt.DeepEquals, unit.Binding.Segments)
		c.Assert(unit.Rights.AllowedUses, qt.Contains, "training")
	}
	var sources, targets []string
	for _, group := range bank.ReservedGroups {
		c.Assert(group.Partition, qt.Equals, "training")
		sources = append(sources, group.Sources...)
	}
	slices.Sort(sources)
	c.Assert(sources, qt.DeepEquals, []string{"d0", "d1", "d2", "d5"})
	for _, target := range bank.ReservedTargets {
		targets = append(targets, target.SourceID)
	}
	c.Assert(targets, qt.Contains, "d1")
	c.Assert(targets, qt.Not(qt.Contains), "d3")
	c.Assert(targets, qt.Not(qt.Contains), "d4")
	// d5 has no prose, but its transitive source group is still reserved.
	c.Assert(targets, qt.Not(qt.Contains), "d5")
	again, err := corpus.BuildCompressionBank(t.Context(), artifact, round, files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, bank)
}

func TestCompressionBankOwnsItsProvenance(t *testing.T) {
	c := qt.New(t)
	manifest, files := compressionManifest()
	artifact := compressionArtifact(t, manifest, files)
	round, options := compressionRound(t, artifact)
	bank, err := corpus.BuildCompressionBank(t.Context(), artifact, round, files, options)
	c.Assert(err, qt.IsNil)
	bank.Options.Cohorts[0].UnitIDs[0] = "changed"
	bank.Cohorts[0].Units[0].Rights.AllowedUses[0] = "changed"
	bank.Cohorts[0].Units[0].Binding.Segments[0].End++
	bank.ReservedGroups[0].Keys[0] = "changed"
	bank.Verification.Producer.NLP.Capabilities[0] = "changed"
	again, err := corpus.BuildCompressionBank(t.Context(), artifact, round, files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(again.SHA256, qt.Equals, bank.SHA256)
	c.Assert(again.Options.Cohorts[0].UnitIDs[0], qt.Not(qt.Equals), "changed")
	c.Assert(artifact.Plan.Manifest.Sources[0].Format, qt.Equals, document.Markdown)
}
