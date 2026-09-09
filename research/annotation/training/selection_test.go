package training

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func TestCalibrationRequiresTrainingRightsOnlyWhenFitted(t *testing.T) {
	c := qt.New(t)
	input := fixtureInput(t)
	input.manifest.Sources[2].Rights.AllowedUses = []string{"annotation", "evaluation"}
	candidates, round := input.compile(t)
	result, err := Run(t.Context(), candidates, round, input.files, fittingOptions())
	c.Assert(err, qt.ErrorMatches, ".*training permission.*")
	c.Assert(result, qt.DeepEquals, Artifact{})
	options := fittingOptions()
	options.Calibration = "none"
	result, err = Run(t.Context(), candidates, round, input.files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Calibration, qt.IsNil)
	c.Assert(result.Partitions[2].Excluded["reserved_partition"], qt.Equals, 4)
}

func TestSelectionRejectsRepresentationChanges(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*unswell.PreparedFeatureSource)
	}{
		{"NLP", func(s *unswell.PreparedFeatureSource) { s.NLP.Version = "different" }},
		{"capabilities", func(s *unswell.PreparedFeatureSource) { s.Capabilities = nil }},
		{"policy", func(s *unswell.PreparedFeatureSource) { s.PolicyHash = "different" }},
		{"vocabulary", func(s *unswell.PreparedFeatureSource) { s.VocabularyHash = "different" }},
		{"extraction", func(s *unswell.PreparedFeatureSource) { s.ExtractionPolicyHash = "different" }},
		{"preparation", func(s *unswell.PreparedFeatureSource) { s.PreparationHash = "different" }},
		{"quotes", func(s *unswell.PreparedFeatureSource) { s.IncludeQuotes = !s.IncludeQuotes }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			input := fixtureInput(t)
			candidates, round := input.compile(t)
			options := fittingOptions()
			joined, err := corpus.Join(t.Context(), candidates, round, input.files, options.Features)
			c.Assert(err, qt.IsNil)
			row.edit(&joined.Features.Sources[1])
			_, err = selectRows(t.Context(), candidates.Plan, joined, options)
			c.Assert(err, qt.ErrorMatches, ".*incompatible feature, policy, or NLP identities.*")
		})
	}
}

func TestUnresolvedDecisionsDoNotBecomeNegativeRows(t *testing.T) {
	c := qt.New(t)
	input := fixtureInput(t)
	candidates, round := input.compile(t)
	options := fittingOptions()
	joined, err := corpus.Join(t.Context(), candidates, round, input.files, options.Features)
	c.Assert(err, qt.IsNil)
	for i := range joined.Decisions.Units {
		decision := &joined.Decisions.Units[i]
		if decision.UnitID == "u000002" {
			decision.Status, decision.Reason, decision.Label = "unresolved", "uncertain", nil
		}
	}
	selected, err := selectRows(t.Context(), candidates.Plan, joined, options)
	c.Assert(err, qt.IsNil)
	c.Assert(selected.training, qt.HasLen, 1)
	c.Assert(selected.training[0].Label, qt.Equals, 1)
	c.Assert(selected.partitions[0].Excluded["decision/uncertain"], qt.Equals, 1)
	_, err = binaryLabel(annotation.EditorialDecision{Status: "resolved"})
	c.Assert(err, qt.IsNotNil)
}
