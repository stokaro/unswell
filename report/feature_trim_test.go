package report

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestTrimmedBindingsRetainInteriorSourceSegments(t *testing.T) {
	original := []document.Span{{Start: 1, End: 5}, {Start: 9, End: 14}, {Start: 19, End: 23}}
	for _, row := range []struct {
		name    string
		trimmed []document.Span
		valid   bool
	}{
		{"empty", nil, true},
		{"unchanged", original, true},
		{"outer trim", []document.Span{{Start: 3, End: 5}, {Start: 9, End: 14}, {Start: 19, End: 21}}, true},
		{"single interval", []document.Span{{Start: 10, End: 12}}, true},
		{"gap", []document.Span{{Start: 6, End: 8}}, false},
		{"bounding span", []document.Span{{Start: 1, End: 23}}, false},
		{"dropped interior", []document.Span{{Start: 1, End: 5}, {Start: 19, End: 23}}, false},
		{"interior trim", []document.Span{{Start: 1, End: 5}, {Start: 10, End: 14}, {Start: 19, End: 23}}, false},
		{"shortened interior", []document.Span{{Start: 1, End: 4}, {Start: 9, End: 14}}, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			err := validateTrimmedBinding(original, row.trimmed)
			c.Assert(err == nil, qt.Equals, row.valid)
		})
	}
}
