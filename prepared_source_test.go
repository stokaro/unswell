package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestOriginalUnitPreservesOriginalByteIdentity(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(preparedOptions())
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(
		"# Guide\r\n\r\nThe cache expires. The cache expires.\r\n\r\nThe **résumé** remains.\r\n")}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	prepared := result.PreparedFeatures.Sources[0]
	c.Assert(prepared.Units, qt.Not(qt.HasLen), 0)
	var repeated []document.Span
	var formatted bool
	for i, unit := range prepared.Units {
		original, err := prepared.OriginalUnit(source, i)
		c.Assert(err, qt.IsNil)
		c.Assert(original.Target, qt.HasLen, len(unit.Binding.Segments))
		c.Assert(original.Context, qt.HasLen, len(unit.Binding.ContextSpans))
		for _, segment := range append(original.Target, original.Context...) {
			c.Assert(segment.Text, qt.Equals, string(source.Bytes[segment.Span.Start:segment.Span.End]))
		}
		if unit.Binding.Kind == "sentence" && strings.Join(segmentTexts(original.Target), "") == "The cache expires." {
			repeated = append(repeated, original.Target[0].Span)
		}
		if strings.Contains(strings.Join(segmentTexts(original.Target), ""), "résumé") {
			formatted = true
			c.Assert(len(original.Target) > 1, qt.IsTrue)
			c.Assert(strings.Join(segmentTexts(original.Target), ""), qt.Not(qt.Contains), "**")
		}
	}
	c.Assert(repeated, qt.HasLen, 2)
	c.Assert(repeated[0], qt.Not(qt.DeepEquals), repeated[1])
	c.Assert(formatted, qt.IsTrue)

	original, err := prepared.OriginalUnit(source, 0)
	c.Assert(err, qt.IsNil)
	want := original.Target[0].Text
	source.Bytes[original.Target[0].Span.Start] = 'X'
	c.Assert(original.Target[0].Text, qt.Equals, want)
}

func TestOriginalUnitRejectsWrongSourceAndInvalidBindings(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(preparedOptions())
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("The résumé remains.\n")}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	prepared := result.PreparedFeatures.Sources[0]
	c.Assert(prepared.Units, qt.Not(qt.HasLen), 0)

	_, err = prepared.OriginalUnit(source, -1)
	c.Assert(err, qt.IsNotNil)
	_, err = prepared.OriginalUnit(source, len(prepared.Units))
	c.Assert(err, qt.IsNotNil)

	changed := source
	changed.Name = "other.md"
	_, err = prepared.OriginalUnit(changed, 0)
	c.Assert(err, qt.IsNotNil)
	changed = source
	changed.Bytes = append([]byte(nil), source.Bytes...)
	changed.Bytes[0] = 'X'
	_, err = prepared.OriginalUnit(changed, 0)
	c.Assert(err, qt.IsNotNil)

	invalid := prepared
	invalid.TargetCount++
	_, err = invalid.OriginalUnit(source, 0)
	c.Assert(err, qt.IsNotNil)
	invalid = prepared
	invalid.Units = append([]unswell.PreparedFeatureUnit(nil), prepared.Units...)
	invalid.Units[0].Binding.Contract = "other"
	_, err = invalid.OriginalUnit(source, 0)
	c.Assert(err, qt.IsNotNil)

	for _, span := range []document.Span{
		{Start: -1, End: 2},
		{Start: 0, End: len(source.Bytes) + 1},
		{Start: 5, End: 6}, // Interior UTF-8 byte in résumé.
	} {
		invalid = prepared
		invalid.Units = append([]unswell.PreparedFeatureUnit(nil), prepared.Units...)
		invalid.Units[0].Binding.Segments = []document.Span{span}
		_, err = invalid.OriginalUnit(source, 0)
		c.Assert(err, qt.IsNotNil)
	}
	invalid = prepared
	invalid.Units = append([]unswell.PreparedFeatureUnit(nil), prepared.Units...)
	invalid.Units[0].Binding.ContextSpans = []document.Span{{Start: 0, End: 3}}
	_, err = invalid.OriginalUnit(source, 0)
	c.Assert(err, qt.IsNotNil)
}

func segmentTexts(segments []unswell.PreparedSourceSegment) []string {
	result := make([]string, len(segments))
	for i, segment := range segments {
		result[i] = segment.Text
	}
	return result
}
