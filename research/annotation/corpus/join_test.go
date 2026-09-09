package corpus_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func candidateFixture(t *testing.T) (corpus.Artifact, map[string][]byte) {
	t.Helper()
	c := qt.New(t)
	manifest, files := sample()
	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, files)
	c.Assert(err, qt.IsNil)
	return artifact, files
}

func candidateRound(t *testing.T, units []annotation.Unit) *annotation.Round {
	t.Helper()
	c := qt.New(t)
	data, err := os.ReadFile("../testdata/tutorial.json")
	c.Assert(err, qt.IsNil)
	var round map[string]any
	c.Assert(json.Unmarshal(data, &round), qt.IsNil)
	round["units"], round["judgments"], round["adjudications"] = units, []any{}, []any{}
	loaded, err := annotation.Load(t.Context(), encoded(c, round))
	c.Assert(err, qt.IsNil)
	return loaded
}

func TestJoinKeepsEveryCandidateAndMissingDecisions(t *testing.T) {
	c := qt.New(t)
	artifact, files := candidateFixture(t)
	round := candidateRound(t, []annotation.Unit{artifact.Units[0].Unit})
	got, err := corpus.Join(t.Context(), artifact, round, files, []string{"prose-words", "noun-token-ratio"})
	c.Assert(err, qt.IsNil)
	c.Assert(got.Version, qt.Equals, corpus.JoinedVersion)
	c.Assert(got.Status, qt.Equals, "verified_targets_with_measured_features")
	c.Assert(got.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(got.Verification.ArtifactSHA256, qt.Equals, artifact.SHA256)
	c.Assert(got.Decisions.Basis, qt.Equals, "simulation")
	c.Assert(got.Decisions.Units, qt.HasLen, 1)
	c.Assert(got.Decisions.Units[0].Reason, qt.Equals, "missing_judgments")
	c.Assert(got.Decisions.Units[0].Label, qt.IsNil)
	c.Assert(got.Features.Requested, qt.DeepEquals, []string{"noun-token-ratio", "prose-words"})
	c.Assert(got.Bindings, qt.HasLen, len(artifact.Units))
	byHash := make(map[string]float64)
	for _, source := range got.Features.Sources {
		c.Assert(source.IncludeQuotes, qt.IsFalse)
		c.Assert(source.IncludeStructure, qt.IsFalse)
		c.Assert(source.PolicyHash, qt.Not(qt.Equals), got.Decisions.ProfileSHA256)
		for _, unit := range source.Units {
			for _, value := range unit.Values {
				if value.ID == "prose-words" {
					c.Assert(value.Number, qt.IsNotNil)
					byHash[source.Path+unit.InputHash] = *value.Number
				}
			}
		}
	}
	for i, binding := range got.Bindings {
		candidate := artifact.Units[i]
		c.Assert(binding.UnitID, qt.Equals, candidate.Unit.ID)
		c.Assert(binding.SourceID, qt.Equals, candidate.SourceID)
		c.Assert(binding.GroupID, qt.Equals, candidate.GroupID)
		c.Assert(binding.Partition, qt.Equals, candidate.Partition)
		c.Assert(byHash[binding.Path+binding.FeatureInputHash], qt.Equals, float64(candidate.Words))
	}
	again, err := corpus.Join(t.Context(), artifact, round, files, []string{"noun-token-ratio", "prose-words"})
	c.Assert(err, qt.IsNil)
	c.Assert(got, qt.DeepEquals, again)
	wantHash := got.SHA256
	got.SHA256 = ""
	c.Assert(hash(encoded(c, got)), qt.Equals, wantHash)
	// Test binaries can omit module dependencies; the NLP identity is always present.
	for i := range got.Verification.Producer.Dependencies {
		got.Verification.Producer.Dependencies[i].Version = "changed"
	}
	c.Assert(artifact.Pipeline.Dependencies, qt.DeepEquals, again.Verification.Producer.Dependencies)
	c.Assert(got.Verification.Producer.NLP.Capabilities, qt.Not(qt.HasLen), 0)
	got.Verification.Producer.NLP.Capabilities[0] = "changed"
	c.Assert(artifact.Pipeline.NLP.Capabilities, qt.DeepEquals, again.Verification.Producer.NLP.Capabilities)
	got.Features.Sources[0].Units[0].Binding.Segments[0].Start++
	c.Assert(again.Features.Sources[0].Units[0].Binding.Segments, qt.Not(qt.DeepEquals), got.Features.Sources[0].Units[0].Binding.Segments)
}

func TestJoinRejectsSourcesAndChangedCorpus(t *testing.T) {
	for _, name := range []string{"source", "notice", "candidate", "partition", "pipeline"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			artifact, files := candidateFixture(t)
			round := candidateRound(t, []annotation.Unit{artifact.Units[0].Unit})
			switch name {
			case "source":
				files["readme.md"][0]++
			case "notice":
				files["LICENSE"][0]++
			case "candidate":
				artifact.Units[0].Unit.Text = "Invented replacement."
			case "partition":
				artifact.Units[0].Partition = "final_test"
			case "pipeline":
				artifact.Pipeline.NLP.Model += "-changed"
			}
			artifact.SHA256 = ""
			artifact.SHA256 = hash(encoded(c, artifact))
			result, err := corpus.Join(t.Context(), artifact, round, files, []string{"prose-words"})
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, corpus.JoinedArtifact{})
		})
	}
}

func TestJoinSelectionAndCancellation(t *testing.T) {
	c := qt.New(t)
	artifact, files := candidateFixture(t)
	unit := artifact.Units[0].Unit
	round := candidateRound(t, []annotation.Unit{unit})
	for _, features := range [][]string{nil, {"unknown"}, {"prose-words", "prose-words"}, {"activation/syntax.noun-stack"}} {
		result, err := corpus.Join(t.Context(), artifact, round, files, features)
		c.Assert(err, qt.IsNotNil)
		c.Assert(result, qt.DeepEquals, corpus.JoinedArtifact{})
	}
	unit.Source.RepositoryID = "other-project"
	changed := candidateRound(t, []annotation.Unit{unit})
	_, err := corpus.Join(t.Context(), artifact, changed, files, []string{"prose-words"})
	c.Assert(err, qt.ErrorMatches, ".*does not match.*")
	_, err = corpus.Join(t.Context(), artifact, nil, files, []string{"prose-words"})
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := corpus.Join(ctx, artifact, round, files, []string{"prose-words"})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, corpus.JoinedArtifact{})
}

func TestJoinPreservesEmptySourcesAndExtractionExclusions(t *testing.T) {
	c := qt.New(t)
	manifest, files := sample()
	manifest.Policy = extract.Policy{Languages: map[document.Format]extract.LanguagePolicy{
		document.Go: {Contexts: []string{}},
	}}
	manifest.UnitKinds = []string{"fragment"}
	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, files)
	c.Assert(err, qt.IsNil)
	round := candidateRound(t, []annotation.Unit{artifact.Units[0].Unit})
	got, err := corpus.Join(t.Context(), artifact, round, files, []string{"prose-words"})
	c.Assert(err, qt.IsNil)
	c.Assert(got.Features.Sources, qt.HasLen, 2)
	c.Assert(got.Features.Sources[1].Path, qt.Equals, "sample.go")
	c.Assert(got.Features.Sources[1].TargetCount, qt.Equals, 0)
	c.Assert(got.Features.Sources[1].Units, qt.HasLen, 0)
	c.Assert(got.Features.Kinds, qt.Not(qt.Contains), "paragraph")
}

func TestJoinPreservesInheritedContextsAndExceptionDefaults(t *testing.T) {
	for _, row := range []struct {
		name   string
		policy extract.Policy
		units  int
	}{
		{"global defaults", extract.Policy{}, 3},
		{"comments", extract.Policy{Languages: map[document.Format]extract.LanguagePolicy{
			document.Go: {Contexts: []string{"comment"}},
		}}, 2},
		{"exception", extract.Policy{Exceptions: []extract.Exception{{ID: "omit-string", Paths: []string{"sample.go"},
			Kinds: []string{"string"}, Reason: "Exercise an exception with omitted optional selectors."}}}, 2},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			manifest, files := sample()
			manifest.Policy = row.policy
			plan, err := corpus.MakePlan(t.Context(), manifest)
			c.Assert(err, qt.IsNil)
			artifact, err := corpus.Build(t.Context(), plan, files)
			c.Assert(err, qt.IsNil)
			round := candidateRound(t, []annotation.Unit{artifact.Units[0].Unit})
			got, err := corpus.Join(t.Context(), artifact, round, files, []string{"prose-words"})
			c.Assert(err, qt.IsNil)
			c.Assert(got.Features.Sources[1].TargetCount, qt.Equals, row.units)
		})
	}
}
