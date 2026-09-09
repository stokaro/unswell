package corpus

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation"
)

func TestBindingRequiresAllSegmentsContextAndUniqueTargets(t *testing.T) {
	for _, name := range []string{"exact", "bounding span", "context", "source", "kind", "missing", "extra", "duplicate"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			unit := annotation.Unit{ID: "u000001", Kind: "fragment", Text: "Cache entry", Context: "Cache entry",
				Source: annotation.Source{SHA256: "source-hash", Segments: []document.Span{{Start: 2, End: 7}, {Start: 12, End: 17}}}}
			artifact := Artifact{Plan: Plan{Manifest: Manifest{Sources: []Source{{ID: "d1", Path: "cache.md"}}}},
				Units: []Candidate{{SourceID: "d1", GroupID: "g1", Partition: "training", Unit: unit}}}
			prepared := unswell.PreparedFeatureUnit{InputHash: "measured-input", Binding: nlp.UnitBinding{Kind: unit.Kind,
				TextSHA256: hashText(unit.Text), ContextSHA256: hashText(unit.Context), Segments: unit.Source.Segments}}
			collection := unswell.PreparedFeatureCollection{Sources: []unswell.PreparedFeatureSource{
				{Path: "cache.md", SourceHash: unit.Source.SHA256, Units: []unswell.PreparedFeatureUnit{prepared}},
			}}
			changePreparedTarget(&collection.Sources[0], name)
			result, err := bindCandidates(t.Context(), artifact, collection)
			if name == "exact" {
				c.Assert(err, qt.IsNil)
				c.Assert(result, qt.DeepEquals, []FeatureBinding{{UnitID: "u000001", SourceID: "d1", Path: "cache.md",
					GroupID: "g1", Partition: "training", FeatureInputHash: "measured-input"}})
			} else {
				c.Assert(err, qt.IsNotNil)
				c.Assert(result, qt.IsNil)
			}
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			_, err = bindCandidates(ctx, artifact, collection)
			c.Assert(err, qt.ErrorIs, context.Canceled)
		})
	}
}

func changePreparedTarget(source *unswell.PreparedFeatureSource, name string) {
	unit := &source.Units[0]
	switch name {
	case "bounding span":
		unit.Binding.Segments = []document.Span{{Start: 2, End: 17}}
	case "context":
		unit.Binding.ContextSHA256 = hashText("Different neighboring context")
	case "source":
		source.SourceHash = "different-source"
	case "kind":
		unit.Binding.Kind = "paragraph"
	case "missing":
		source.Units = nil
	case "duplicate", "extra":
		other := *unit
		if name == "extra" {
			other.Binding.TextSHA256 = hashText("extra")
		}
		source.Units = append(source.Units, other)
	}
}
