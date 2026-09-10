package corpus_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func TestCompressionBankRejectsIneligibleSelections(t *testing.T) {
	manifest, files := compressionManifest()
	artifact := compressionArtifact(t, manifest, files)
	for _, row := range []struct {
		name, message string
		edit          func(*corpus.CompressionBankOptions)
	}{
		{"simulation", ".*allow_simulation.*", func(o *corpus.CompressionBankOptions) { o.AllowSimulation = false }},
		{"missing target", ".*training target.*", func(o *corpus.CompressionBankOptions) { o.Cohorts[0].UnitIDs[0] = "missing" }},
		{"wrong kind", ".*selected kind.*", func(o *corpus.CompressionBankOptions) { o.Kind = "sentence" }},
		{"wrong cohort", ".*does not match cohort.*", func(o *corpus.CompressionBankOptions) { o.Cohorts[0].Origin = "generated" }},
		{"one-sided mixed", ".*requires both human and generated.*", func(o *corpus.CompressionBankOptions) {
			o.Cohorts[2].UnitIDs = o.Cohorts[0].UnitIDs
		}},
		{"baseline budget", ".*input-byte budget.*", func(o *corpus.CompressionBankOptions) { o.Compression.MaxInputBytes = 1 }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			round, options := compressionRound(t, artifact)
			row.edit(&options)
			bank, err := corpus.BuildCompressionBank(t.Context(), artifact, round, files, options)
			c.Assert(err, qt.ErrorMatches, row.message)
			c.Assert(bank, qt.DeepEquals, corpus.CompressionBank{})
		})
	}
}

func TestCompressionBankRejectsReservedPartitionsAndPermissions(t *testing.T) {
	for _, name := range []string{"development", "calibration", "final_test", "permission", "prefix limit"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			manifest, files := compressionManifest()
			message := ".*training target.*"
			switch name {
			case "permission":
				manifest.Sources[0].Rights.AllowedUses = []string{"annotation"}
				message = ".*training permission.*"
			case "prefix limit":
				files["readme.md"] = []byte(strings.Repeat("Cache ", 6000) + "stops.")
				manifest.Sources[0].SHA256, manifest.Sources[0].Bytes = hash(files["readme.md"]), len(files["readme.md"])
				message = ".*32KiB prefix limit.*"
			default:
				manifest.Sources[0].Partition, manifest.Sources[1].Partition, manifest.Sources[5].Partition = name, name, name
			}
			artifact := compressionArtifact(t, manifest, files)
			round, options := compressionRound(t, artifact)
			bank, err := corpus.BuildCompressionBank(t.Context(), artifact, round, files, options)
			c.Assert(err, qt.ErrorMatches, message)
			c.Assert(bank, qt.DeepEquals, corpus.CompressionBank{})
		})
	}
}

func TestCompressionBankRejectsUncuratedOrigins(t *testing.T) {
	manifest, files := compressionManifest()
	artifact := compressionArtifact(t, manifest, files)
	_, options := compressionRound(t, artifact)
	for _, label := range []string{"unknown", "mixed", "human_ai_edited", "generated_human_edited"} {
		t.Run(label, func(t *testing.T) {
			c := qt.New(t)
			var units []annotation.Unit
			for _, candidate := range artifact.Units {
				if candidate.Unit.ID == options.Cohorts[0].UnitIDs[0] {
					unit := candidate.Unit
					unit.Origin = annotation.Origin{Label: label, Scope: "unit", Evidence: "Simulated non-endpoint claim.",
						GenerationRecord: "fixture:simulated-edit"}
					units = append(units, unit)
				}
			}
			round := candidateRound(t, units)
			bank, err := corpus.BuildCompressionBank(t.Context(), artifact, round, files, options)
			c.Assert(err, qt.ErrorMatches, ".*curated unit-scoped endpoint claim.*")
			c.Assert(bank, qt.DeepEquals, corpus.CompressionBank{})
		})
	}
}

func TestCompressionBankReproducesSourcesAndHonorsCancellation(t *testing.T) {
	for _, name := range []string{"source", "notice", "target", "round", "canceled"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			manifest, files := compressionManifest()
			artifact := compressionArtifact(t, manifest, files)
			round, options := compressionRound(t, artifact)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			switch name {
			case "source":
				files["readme.md"][0]++
			case "notice":
				files["LICENSE"][0]++
			case "target":
				artifact.Units[0].Unit.Text += " invented"
				artifact.SHA256 = ""
				artifact.SHA256 = hash(encoded(c, artifact))
			case "round":
				round = nil
			case "canceled":
				cancel()
			}
			bank, err := corpus.BuildCompressionBank(ctx, artifact, round, files, options)
			c.Assert(err, qt.IsNotNil)
			c.Assert(bank, qt.DeepEquals, corpus.CompressionBank{})
			if name == "canceled" {
				c.Assert(err, qt.ErrorIs, context.Canceled)
			}
		})
	}
}
