package corpus

// White-box tests: Inject ambiguous rule bindings and change each component of their identity;
// JoinRules computes its own feature collection instead of accepting these intermediate states.

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation"
)

func ruleBindingFixture() (Artifact, unswell.FeatureCollection) {
	segments := []document.Span{{Start: 3, End: 8}, {Start: 12, End: 17}}
	unit := annotation.Unit{ID: "u000001", Kind: "fragment", Text: "Cache entry", Context: "Cache entry",
		Source: annotation.Source{SHA256: "source-hash", Segments: segments}}
	artifact := Artifact{Plan: Plan{Manifest: Manifest{Sources: []Source{{ID: "d1", Path: "cache.go"}}}},
		Units: []Candidate{{SourceID: "d1", GroupID: "g1", Partition: "training", Unit: unit}}}
	block := unswell.FeatureUnit{UnitID: 0, InputHash: "token-input", Binding: &unswell.FeatureBlockBinding{
		Contract: unswell.FeatureBlockBindingContract, TextSHA256: hashText(" Cache entry"),
		Segments:      []document.Span{{Start: 2, End: 8}, {Start: 12, End: 17}},
		TrimmedSHA256: hashText(unit.Text), TrimmedSegments: segments}}
	collection := unswell.FeatureCollection{Sources: []unswell.FeatureSource{
		{Path: "cache.go", SourceHash: unit.Source.SHA256, Units: []unswell.FeatureUnit{block}},
	}}
	return artifact, collection
}

func TestRuleBindingRequiresCompleteTextContextAndSegments(t *testing.T) {
	for _, name := range []string{"exact", "bounding span", "context", "partial text", "source", "missing"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			artifact, collection := ruleBindingFixture()
			unit := &artifact.Units[0].Unit
			switch name {
			case "bounding span":
				unit.Source.Segments = []document.Span{{Start: 3, End: 17}}
			case "context":
				unit.Context = "Different neighboring text."
			case "partial text":
				unit.Text, unit.Context = "Cache", "Cache"
			case "source":
				unit.Source.SHA256 = "different-source"
			case "missing":
				collection.Sources[0].Units = nil
			}
			result, err := bindRuleCandidates(t.Context(), artifact, collection)
			c.Assert(err, qt.IsNil)
			c.Assert(result, qt.HasLen, 1)
			if name == "exact" {
				c.Assert(result[0].BlockID, qt.IsNotNil)
				c.Assert(*result[0].BlockID, qt.Equals, 0)
				c.Assert(result[0].InputHash, qt.HasLen, 64)
				c.Assert(result[0].Reason, qt.Equals, "")
			} else {
				c.Assert(result[0].BlockID, qt.IsNil)
				c.Assert(result[0].InputHash, qt.Equals, "")
				c.Assert(result[0].Reason, qt.Equals, "no_complete_block_match")
			}
		})
	}
}

func TestRuleBindingRejectsMissingOrAmbiguousContracts(t *testing.T) {
	for _, name := range []string{"missing", "future", "duplicate"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			artifact, collection := ruleBindingFixture()
			source := &collection.Sources[0]
			switch name {
			case "missing":
				source.Units[0].Binding = nil
			case "future":
				source.Units[0].Binding.Contract = "future"
			case "duplicate":
				source.Units = append(source.Units, source.Units[0])
			}
			result, err := bindRuleCandidates(t.Context(), artifact, collection)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.IsNil)
		})
	}
}

func TestRuleBindingIdentityIncludesActualInputAndPolicy(t *testing.T) {
	for _, name := range []string{"original text", "original mapping", "policy", "ruleset", "NLP", "context"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			artifact, collection := ruleBindingFixture()
			before, err := bindRuleCandidates(t.Context(), artifact, collection)
			c.Assert(err, qt.IsNil)
			source := &collection.Sources[0]
			switch name {
			case "original text":
				source.Units[0].Binding.TextSHA256 = hashText("  Cache entry")
			case "original mapping":
				source.Units[0].Binding.Segments[0].Start = 1
			case "policy":
				source.PolicyHash = "changed"
			case "ruleset":
				source.RulesetHash = "changed"
			case "NLP":
				source.NLP.Version = "changed"
			case "context":
				source.Units[0].ContextHash = "changed"
			}
			after, err := bindRuleCandidates(t.Context(), artifact, collection)
			c.Assert(err, qt.IsNil)
			c.Assert(after[0].BlockID, qt.IsNotNil)
			c.Assert(after[0].InputHash, qt.Not(qt.Equals), before[0].InputHash)
		})
	}
}
