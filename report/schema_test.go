package report_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func TestOfficialSARIFSchema(t *testing.T) {
	c := qt.New(t)
	data, err := os.ReadFile("testdata/sarif-schema-2.1.0.json")
	c.Assert(err, qt.IsNil)
	var schemaData any
	c.Assert(json.Unmarshal(data, &schemaData), qt.IsNil)
	compiler := jsonschema.NewCompiler()
	c.Assert(compiler.AddResource("sarif.json", schemaData), qt.IsNil)
	schema, err := compiler.Compile("sarif.json")
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Config: []byte("version: 1\nextends: [builtin:strict]\n"), IncludeSource: true})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "docs/😀 sample.md", Format: document.Markdown,
		Bytes: []byte("# 😀\n\nIt is **important** to note that the client can retry.\r\n")})
	c.Assert(err, qt.IsNil)
	var buffer bytes.Buffer
	c.Assert(report.Write(&buffer, "sarif", result, report.Options{}), qt.IsNil)
	var value any
	c.Assert(json.Unmarshal(buffer.Bytes(), &value), qt.IsNil)
	c.Assert(schema.Validate(value), qt.IsNil)
	// The official schema must reject an impossible line number in a result.
	invalid := bytes.Replace(buffer.Bytes(), []byte(`"startLine": 3`), []byte(`"startLine": 0`), 1)
	c.Assert(bytes.Equal(invalid, buffer.Bytes()), qt.IsFalse)
	c.Assert(json.Unmarshal(invalid, &value), qt.IsNil)
	c.Assert(schema.Validate(value), qt.IsNotNil)
}
