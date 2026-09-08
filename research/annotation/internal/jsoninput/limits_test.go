package jsoninput_test

import (
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

type observedDecode struct{ calls int }

func (d *observedDecode) UnmarshalJSON([]byte) error {
	d.calls++
	return nil
}

func TestBoundsPrecedeTypedDecoding(t *testing.T) {
	for _, row := range []struct{ name, input string }{
		{"array", `{"values":[{},{},{}]}`},
		{"nested", `{"values":{"nested":[{},{},{}]}}`},
		{"object", `{"a":0,"b":0,"c":0}`},
		{"field alias", `{"ITEMS":[1,2]}`},
		{"Unicode field alias", `{"itemſ":[1,2]}`},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			var destination observedDecode
			limits := jsoninput.Limits{Array: 2, Object: 2, Arrays: map[string]int{"items": 1}}
			err := jsoninput.Decode(t.Context(), []byte(row.input), 1024, &destination, limits)
			c.Assert(err, qt.ErrorMatches, "research JSON .* exceeds .* entries|research JSON object exceeds .* properties")
			c.Assert(destination.calls, qt.Equals, 0)
		})
	}
}

func TestCollectionBoundariesAndOverrides(t *testing.T) {
	c := qt.New(t)
	var destination map[string]json.RawMessage
	limits := jsoninput.Limits{Array: 2, Object: 2, Arrays: map[string]int{"keys": 3}}
	c.Assert(jsoninput.Decode(t.Context(), []byte(`{"values":[1,2],"keys":[1,2,3]}`), 1024, &destination, limits), qt.IsNil)
	c.Assert(string(destination["keys"]), qt.Equals, "[1,2,3]")
	// The existing annotation checker retains its schema-defined collection limits.
	c.Assert(jsoninput.Check(t.Context(), []byte(`{"values":[1,2,3]}`), 1024), qt.IsNil)
}
