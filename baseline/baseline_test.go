package baseline_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/baseline"
)

func hash(text string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(text))) }

func sampleSnapshot() baseline.Snapshot {
	compatibility := baseline.Compatibility{PolicyHash: hash("policy"), RulesHash: hash("rules"), NLPHash: hash("nlp"),
		FeatureContract: "features-v1", ScoringContract: "scoring-v1", ModelHash: hash("no-model")}
	doc := baseline.Document{Path: "guide.md", Format: "markdown", SourceHash: hash("source"),
		PolicyHash: hash("file-policy"), SuppressionHash: hash("no-directives")}
	identity := baseline.Identity{Path: doc.Path, Kind: "finding", RuleID: "policy.banned-phrases", RuleVersion: "1",
		StructureHash: hash("heading:Retries"), ContentHash: hash("The client may retry 3 times."), EvidenceHash: hash("retry")}
	return baseline.Snapshot{Complete: true, Compatibility: compatibility, Documents: []baseline.Document{doc},
		Candidates: []baseline.Identity{identity}}
}

func twoDocuments() baseline.Snapshot {
	snapshot := sampleSnapshot()
	doc := snapshot.Documents[0]
	doc.Path = "reference.md"
	identity := snapshot.Candidates[0]
	identity.Path = doc.Path
	snapshot.Documents = append(snapshot.Documents, doc)
	snapshot.Candidates = append(snapshot.Candidates, identity)
	return snapshot
}

func TestComparisonRetainsIdentityAcrossUnrelatedSourceChanges(t *testing.T) {
	c := qt.New(t)
	snapshot := sampleSnapshot()
	file, err := baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNil)
	snapshot.Documents[0].SourceHash = hash("unrelated lines were inserted")
	result, err := baseline.Compare(t.Context(), file, snapshot)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Matches, qt.DeepEquals, []baseline.Match{{Fingerprint: file.Entries[0].Fingerprint, State: "existing"}})
	c.Assert(result.Stale, qt.HasLen, 0)
	c.Assert(file.Documents[0].SourceHash, qt.Equals, hash("source"))
}

func TestChangedContentContextAndEvidenceAreNewDebt(t *testing.T) {
	for _, change := range []struct {
		name string
		edit func(*baseline.Identity)
	}{
		{"negation", func(i *baseline.Identity) { i.ContentHash = hash("The client may not retry 3 times.") }},
		{"number", func(i *baseline.Identity) { i.ContentHash = hash("The client may retry 4 times.") }},
		{"structure", func(i *baseline.Identity) { i.StructureHash = hash("heading:Uploads") }},
		{"repeat", func(i *baseline.Identity) { i.EvidenceHash = hash("retry; retry") }},
		{"behavior", func(i *baseline.Identity) { i.RuleVersion = "2" }},
	} {
		t.Run(change.name, func(t *testing.T) {
			c := qt.New(t)
			snapshot := sampleSnapshot()
			file, err := baseline.Create(t.Context(), snapshot)
			c.Assert(err, qt.IsNil)
			change.edit(&snapshot.Candidates[0])
			result, err := baseline.Compare(t.Context(), file, snapshot)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Matches, qt.HasLen, 1)
			c.Assert(result.Matches[0].State, qt.Equals, "new")
			c.Assert(result.Stale, qt.DeepEquals, file.Entries)
		})
	}
}

func TestPartialUpdatePreservesUnobservedDebtAndPrunesObservedStaleEntries(t *testing.T) {
	c := qt.New(t)
	snapshot := twoDocuments()
	file, err := baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNil)
	snapshot.Documents = snapshot.Documents[:1]
	snapshot.Candidates = nil
	result, err := baseline.Compare(t.Context(), file, snapshot)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Stale, qt.DeepEquals, file.Entries[:1])
	c.Assert(result.Unobserved, qt.DeepEquals, file.Entries[1:])
	updated, err := baseline.Update(t.Context(), file, snapshot)
	c.Assert(err, qt.IsNil)
	c.Assert(updated.Entries, qt.DeepEquals, file.Entries[1:])
	c.Assert(updated.Documents, qt.HasLen, 2)
	c.Assert(file.Entries, qt.HasLen, 2)
	updated.Documents[0].Path = "changed.md"
	c.Assert(file.Documents[0].Path, qt.Equals, "guide.md")
	c.Assert(snapshot.Documents[0].Path, qt.Equals, "guide.md")
}

func TestCompatibilityChangesRequireExplicitCoveredUpdate(t *testing.T) {
	for _, change := range []struct {
		name string
		edit func(*baseline.Snapshot)
	}{
		{"policy", func(s *baseline.Snapshot) { s.Compatibility.PolicyHash = hash("new policy") }},
		{"rules", func(s *baseline.Snapshot) { s.Compatibility.RulesHash = hash("new rules") }},
		{"nlp", func(s *baseline.Snapshot) { s.Compatibility.NLPHash = hash("new nlp") }},
		{"model", func(s *baseline.Snapshot) { s.Compatibility.ModelHash = hash("new model") }},
		{"features", func(s *baseline.Snapshot) { s.Compatibility.FeatureContract = "features-v2" }},
		{"scoring", func(s *baseline.Snapshot) { s.Compatibility.ScoringContract = "scoring-v2" }},
		{"format", func(s *baseline.Snapshot) { s.Documents[0].Format = "text" }},
		{"file-policy", func(s *baseline.Snapshot) { s.Documents[0].PolicyHash = hash("new file policy") }},
		{"suppression", func(s *baseline.Snapshot) { s.Documents[0].SuppressionHash = hash("new directive") }},
	} {
		t.Run(change.name, func(t *testing.T) {
			c := qt.New(t)
			snapshot := sampleSnapshot()
			file, err := baseline.Create(t.Context(), snapshot)
			c.Assert(err, qt.IsNil)
			change.edit(&snapshot)
			_, err = baseline.Compare(t.Context(), file, snapshot)
			c.Assert(err, qt.ErrorMatches, ".*explicit update.*")
			updated, err := baseline.Update(t.Context(), file, snapshot)
			c.Assert(err, qt.IsNil)
			_, err = baseline.Compare(t.Context(), updated, snapshot)
			c.Assert(err, qt.IsNil)
		})
	}
	c := qt.New(t)
	snapshot := twoDocuments()
	file, err := baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNil)
	snapshot.Documents = snapshot.Documents[:1]
	snapshot.Candidates = snapshot.Candidates[:1]
	snapshot.Compatibility.PolicyHash = hash("new policy")
	_, err = baseline.Update(t.Context(), file, snapshot)
	c.Assert(err, qt.ErrorMatches, "baseline compatibility update requires observing reference.md")
}

func TestAmbiguousAndIncompleteSnapshotsNeverAcceptDebt(t *testing.T) {
	c := qt.New(t)
	snapshot := sampleSnapshot()
	file, err := baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNil)
	snapshot.Candidates = append(snapshot.Candidates, snapshot.Candidates[0])
	_, err = baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.ErrorMatches, "ambiguous baseline identity.*")
	_, err = baseline.Compare(t.Context(), file, snapshot)
	c.Assert(err, qt.ErrorMatches, "ambiguous baseline identity.*")
	_, err = baseline.Update(t.Context(), file, snapshot)
	c.Assert(err, qt.ErrorMatches, "ambiguous baseline identity.*")
	snapshot = sampleSnapshot()
	snapshot.Complete = false
	_, err = baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.ErrorMatches, "incomplete analysis.*")
	_, err = baseline.Compare(t.Context(), file, snapshot)
	c.Assert(err, qt.ErrorMatches, "incomplete analysis.*")
	_, err = baseline.Update(t.Context(), file, snapshot)
	c.Assert(err, qt.ErrorMatches, "incomplete analysis.*")
}

func TestCanonicalEncodingDoesNotDependOnOrdering(t *testing.T) {
	c := qt.New(t)
	snapshot := twoDocuments()
	file, err := baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNil)
	want, err := baseline.Encode(t.Context(), file)
	c.Assert(err, qt.IsNil)
	slices.Reverse(snapshot.Documents)
	slices.Reverse(snapshot.Candidates)
	other, err := baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNil)
	got, err := baseline.Encode(t.Context(), other)
	c.Assert(err, qt.IsNil)
	c.Assert(string(got), qt.Equals, string(want))
	decoded, err := baseline.Load(t.Context(), got)
	c.Assert(err, qt.IsNil)
	c.Assert(decoded, qt.DeepEquals, file)
}

func TestCancellation(t *testing.T) {
	c := qt.New(t)
	snapshot := sampleSnapshot()
	file, err := baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNil)
	data, err := baseline.Encode(t.Context(), file)
	c.Assert(err, qt.IsNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = baseline.Create(ctx, snapshot)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = baseline.Compare(ctx, file, snapshot)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = baseline.Update(ctx, file, snapshot)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = baseline.Encode(ctx, file)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = baseline.Load(ctx, data)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestInvalidIdentities(t *testing.T) {
	for _, name := range []string{"", ".", "../guide.md", "/guide.md", "C:/guide.md", `docs\guide.md`, "a/../b.md", "a//b.md"} {
		c := qt.New(t)
		identity := sampleSnapshot().Candidates[0]
		identity.Path = name
		_, err := baseline.Fingerprint(identity)
		c.Assert(err, qt.IsNotNil, qt.Commentf("path %q", name))
	}
	c := qt.New(t)
	identity := sampleSnapshot().Candidates[0]
	identity.ContentHash = strings.Repeat("A", 64)
	_, err := baseline.Fingerprint(identity)
	c.Assert(err, qt.IsNotNil)
}
