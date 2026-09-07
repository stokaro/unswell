package report_test

import (
	"bytes"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func TestSavedResultRejectsMissingContractAndFalseCompletion(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{
		Name: "sample.txt", Format: document.Plain, Bytes: []byte("The client retries after a connection failure."),
	})
	c.Assert(err, qt.IsNil)
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"missing manifest", func(value map[string]any) { delete(value, "manifest") }},
		{"missing findings", func(value map[string]any) { delete(value, "findings") }},
		{"incomplete manifest", func(value map[string]any) { value["manifest"].(map[string]any)["complete"] = false }},
		{"unknown field", func(value map[string]any) { value["unexpected"] = true }},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			encoded, err := json.Marshal(result)
			c.Assert(err, qt.IsNil)
			var value map[string]any
			c.Assert(json.Unmarshal(encoded, &value), qt.IsNil)
			row.mutate(value)
			encoded, err = json.Marshal(value)
			c.Assert(err, qt.IsNil)
			_, err = report.Read(bytes.NewReader(encoded))
			c.Assert(err, qt.IsNotNil)
		})
	}
}
