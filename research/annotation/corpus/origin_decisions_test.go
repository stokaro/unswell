package corpus_test

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// provenanceSample builds a manifest whose sources span the cohorts the origin
// profile distinguishes. Every source is one Markdown file of two paragraphs.
func provenanceSample() (corpus.Manifest, map[string][]byte) {
	files := map[string][]byte{"LICENSE": []byte("Test-owned source and notice fixture.\n")}
	notice := corpus.Notice{Path: "LICENSE", SHA256: hash(files["LICENSE"]), Bytes: len(files["LICENSE"])}
	manifest := corpus.Manifest{Version: corpus.Version, ID: "provenance-fixture", Seed: "frozen-seed",
		Weights: corpus.Weights{Training: 6000, Development: 1500, Calibration: 1500, FinalTest: 1000},
		Policy:  extract.Policy{}, UnitKinds: []string{"paragraph"}}
	kinds := []struct {
		id       string
		cohort   string
		origin   annotation.Origin
		snapshot corpus.Snapshot
	}{
		{"generated", "controlled", annotation.Origin{Label: "generated", Scope: "document",
			Evidence: "Response r1 of run pilot", GenerationRecord: "records.json#r1"},
			corpus.Snapshot{Date: "2026-09-11", Confidence: "corroborated", Evidence: "Generation record dates the response", Cohort: "controlled"}},
		{"polished", "controlled", annotation.Origin{Label: "human_ai_edited", Scope: "document",
			Evidence: "Response r2 of run pilot", GenerationRecord: "records.json#r2"},
			corpus.Snapshot{Date: "2026-09-11", Confidence: "corroborated", Evidence: "Generation record dates the response", Cohort: "controlled"}},
		{"historical", "historical-2012", annotation.Origin{Label: "unknown", Scope: "repository",
			Evidence: "Dated snapshot; no unit-level authorship record."},
			corpus.Snapshot{Date: "2012-12-28", Confidence: "vcs_only", Evidence: "commit dated 2012-12-28", Cohort: "historical-2012"}},
		{"contemporary", "contemporary", annotation.Origin{Label: "unknown", Scope: "repository",
			Evidence: "Dated snapshot; no unit-level authorship record."},
			corpus.Snapshot{Date: "2026-06-19", Confidence: "corroborated", Evidence: "release dated 2026-06-19", Cohort: "contemporary"}},
	}
	for i, kind := range kinds {
		name := fmt.Sprintf("%s.md", kind.id)
		files[name] = []byte(fmt.Sprintf("# %s\n\nThe cache retries a request after a short wait and records the delay.\n\n"+
			"Callers keep the identifier unchanged between attempts, so the log stays readable.\n", kind.id))
		snapshot := kind.snapshot
		manifest.Sources = append(manifest.Sources, corpus.Source{ID: fmt.Sprintf("s%d", i), Path: name,
			SHA256: hash(files[name]), Bytes: len(files[name]), Format: document.Markdown, ProseLanguage: "en",
			Repository: "test-project", Document: "test-project/" + name, Reference: "fixture:" + name,
			Topic: "cache", Purpose: "Explain cache behavior", Role: "documentation", Origin: kind.origin,
			Rights: annotation.Rights{License: "test-fixture", Evidence: "Test-owned bytes",
				AllowedUses: []string{"annotation", "training", "evaluation"}},
			Notices: []corpus.Notice{notice}, Snapshot: &snapshot})
	}
	return manifest, files
}

func TestOriginDecisionsLabelFromDeclaredProvenance(t *testing.T) {
	c := qt.New(t)
	manifest, files := provenanceSample()
	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, files)
	c.Assert(err, qt.IsNil)

	decisions, err := corpus.OriginDecisions(t.Context(), artifact)
	c.Assert(err, qt.IsNil)
	c.Assert(decisions.Version, qt.Equals, corpus.OriginDecisionsVersion)
	c.Assert(decisions.Basis, qt.Equals, "declared_provenance")
	c.Assert(decisions.Rubric, qt.Equals, annotation.OriginRubric)
	c.Assert(decisions.ProfileSHA256, qt.Equals, annotation.OriginProfileSHA256())
	c.Assert(decisions.RoundSHA256, qt.Equals, artifact.SHA256)
	c.Assert(decisions.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(decisions.Units, qt.HasLen, len(artifact.Units))

	bySource := map[string]corpus.Candidate{}
	for _, candidate := range artifact.Units {
		bySource[candidate.Unit.ID] = candidate
	}
	outcomes := map[string]map[string]int{}
	for _, decision := range decisions.Units {
		candidate := bySource[decision.UnitID]
		c.Assert(decision.Target, qt.DeepEquals, annotation.TargetOf(candidate.Unit))
		c.Assert(decision.Basis, qt.Equals, "declared_provenance")
		key := decision.Reason
		if decision.Label != nil {
			c.Assert(decision.Status, qt.Equals, "resolved")
			key = *decision.Label
		}
		if outcomes[candidate.SourceID] == nil {
			outcomes[candidate.SourceID] = map[string]int{}
		}
		outcomes[candidate.SourceID][key]++
	}
	c.Assert(outcomes["s0"], qt.DeepEquals, map[string]int{"endpoint_generated": 2})
	c.Assert(outcomes["s1"], qt.DeepEquals, map[string]int{"polished_response": 2})
	c.Assert(outcomes["s2"], qt.DeepEquals, map[string]int{"human_snapshot": 2})
	c.Assert(outcomes["s3"], qt.DeepEquals, map[string]int{"contemporary_snapshot": 2})

	joined, err := corpus.JoinDecisions(t.Context(), artifact, decisions, files, []string{"prose-words"})
	c.Assert(err, qt.IsNil)
	c.Assert(joined.Decisions.SHA256, qt.Equals, decisions.SHA256)
	c.Assert(joined.Bindings, qt.HasLen, len(artifact.Units))

	tampered := decisions
	tampered.Units = append([]annotation.EditorialDecision(nil), decisions.Units...)
	tampered.Units[0].Target.TextSHA256 = "0"
	_, err = corpus.JoinDecisions(t.Context(), artifact, tampered, files, []string{"prose-words"})
	c.Assert(err, qt.ErrorMatches, "decision .* does not match a candidate target")

	foreign := decisions
	foreign.RoundSHA256 = "0"
	c.Assert(corpus.MatchDecisionTargets(t.Context(), artifact, foreign), qt.ErrorMatches, "decisions bind a different candidate artifact")
	_, err = corpus.OriginDecisions(t.Context(), corpus.Artifact{})
	c.Assert(err, qt.ErrorMatches, "origin decisions need a sealed candidate artifact")
}
