package feature_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
)

func TestActivationsPreserveApplicabilityAndMaximum(t *testing.T) {
	c := qt.New(t)
	builder, err := feature.NewActivations(4, 4)
	c.Assert(err, qt.IsNil)
	c.Assert(builder.Observe(feature.BlockObservation{BlockID: 0, Status: "evaluated"}), qt.IsNil)
	c.Assert(builder.Record(0, 250), qt.IsNil)
	c.Assert(builder.Record(0, 750), qt.IsNil)
	c.Assert(builder.Record(0, 750), qt.IsNil)
	c.Assert(builder.Observe(feature.BlockObservation{BlockID: 1, Status: "evaluated"}), qt.IsNil)
	c.Assert(builder.Observe(feature.BlockObservation{BlockID: 2, Status: "inapplicable", Reason: "insufficient_words"}), qt.IsNil)
	values, err := builder.Values("example.rule", false)
	c.Assert(err, qt.IsNil)
	c.Assert(*values[0].Number, qt.Equals, 0.75)
	c.Assert(*values[1].Number, qt.Equals, float64(0))
	c.Assert(values[2].Number, qt.IsNil)
	c.Assert(values[2].Reason, qt.Equals, "inapplicable/insufficient_words")
	c.Assert(values[3].Number, qt.IsNil)
	c.Assert(values[3].Reason, qt.Equals, "applicability_unknown")
	*values[0].Number = 99
	again, err := builder.Values("example.rule", false)
	c.Assert(err, qt.IsNil)
	c.Assert(*again[0].Number, qt.Equals, 0.75)
}

func TestActivationsRequireDeclaredObservations(t *testing.T) {
	c := qt.New(t)
	builder, err := feature.NewActivations(1, 1)
	c.Assert(err, qt.IsNil)
	c.Assert(builder.Record(0, 1000), qt.IsNil)
	values, err := builder.Values("example.rule", false)
	c.Assert(err, qt.IsNil)
	c.Assert(*values[0].Number, qt.Equals, float64(1))
	_, err = builder.Values("example.rule", true)
	c.Assert(err, qt.ErrorMatches, "missing declared block observation for 0")
	c.Assert(builder.Observe(feature.BlockObservation{BlockID: 0, Status: "evaluated"}), qt.IsNotNil)
	_, err = builder.Values("example.rule", false)
	c.Assert(err, qt.IsNotNil)
}

func TestActivationsLatchInvalidOrContradictoryObservations(t *testing.T) {
	for _, row := range []struct {
		name string
		run  func(*feature.ActivationBuilder) error
	}{
		{"invalid ID", func(b *feature.ActivationBuilder) error { return b.Observe(feature.BlockObservation{BlockID: 1}) }},
		{"invalid status", func(b *feature.ActivationBuilder) error { return b.Observe(feature.BlockObservation{Status: "maybe"}) }},
		{"missing reason", func(b *feature.ActivationBuilder) error {
			return b.Observe(feature.BlockObservation{Status: "inapplicable"})
		}},
		{"observed reason", func(b *feature.ActivationBuilder) error {
			return b.Observe(feature.BlockObservation{Status: "evaluated", Reason: "unexpected"})
		}},
		{"negative activation", func(b *feature.ActivationBuilder) error { return b.Record(0, -1) }},
		{"excess activation", func(b *feature.ActivationBuilder) error { return b.Record(0, 1001) }},
		{"invalid evidence ID", func(b *feature.ActivationBuilder) error { return b.Record(-1, 1) }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			builder, err := feature.NewActivations(1, 1)
			c.Assert(err, qt.IsNil)
			c.Assert(row.run(builder), qt.IsNotNil)
			c.Assert(builder.Record(0, 0), qt.IsNotNil)
			values, err := builder.Values("example.rule", false)
			c.Assert(err, qt.IsNotNil)
			c.Assert(values, qt.IsNil)
		})
	}
}

func TestActivationsRejectDuplicateAndContradictoryAccounting(t *testing.T) {
	for _, evidenceFirst := range []bool{false, true} {
		c := qt.New(t)
		builder, err := feature.NewActivations(1, 1)
		c.Assert(err, qt.IsNil)
		inapplicable := feature.BlockObservation{Status: "inapplicable", Reason: "no_prose_words"}
		if evidenceFirst {
			c.Assert(builder.Record(0, 500), qt.IsNil)
			c.Assert(builder.Observe(inapplicable), qt.IsNotNil)
		} else {
			c.Assert(builder.Observe(inapplicable), qt.IsNil)
			c.Assert(builder.Record(0, 500), qt.IsNotNil)
		}
	}
	c := qt.New(t)
	builder, err := feature.NewActivations(1, 1)
	c.Assert(err, qt.IsNil)
	c.Assert(builder.Observe(feature.BlockObservation{Status: "evaluated"}), qt.IsNil)
	c.Assert(builder.Observe(feature.BlockObservation{Status: "evaluated"}), qt.IsNotNil)
	for _, counts := range [][2]int{{-1, 1}, {1, 0}, {2, 1}} {
		_, err := feature.NewActivations(counts[0], counts[1])
		c.Assert(err, qt.IsNotNil)
	}
	for _, reason := range []string{"", "No words", "1word", "source\ntext", "unicode_é"} {
		c.Assert(feature.ValidApplicabilityReason(reason), qt.IsFalse)
	}
}
