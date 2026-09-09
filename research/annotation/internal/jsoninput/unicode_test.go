package jsoninput_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

func TestUnicodeEscapes(t *testing.T) {
	cases := []struct {
		name, value string
		valid       bool
	}{
		{"pair", `"\ud83d\ude00"`, true},
		{"literal slash", `"\\ud800"`, true},
		{"ordinary", `"\u0061"`, true},
		{"low alone", `"\udc00"`, false},
		{"high then ordinary", `"\ud800\u0061"`, false},
		{"high at end", `"\ud800"`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(jsoninput.Check(t.Context(), []byte(tc.value), 1024) == nil, qt.Equals, tc.valid)
		})
	}
}
