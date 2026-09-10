package report

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// SARIFSchemaVersion identifies the pinned specification schema this package
// validates against. The copy is the published schema for that version; it is
// not a claim that a consumer accepts every document the schema allows.
const SARIFSchemaVersion = "2.1.0"

// The schema is embedded so validation needs no network. Source:
// https://json.schemastore.org/sarif-2.1.0.json, the location the reports
// declare, with SHA-256 7c9688f0a1c4a4e1649ecc78521087e664729c1dff56ee8212ff195c7b16132a.
//
//go:embed sarif-schema.json
var sarifSchema []byte

var compiledSARIF = sync.OnceValues(func() (*jsonschema.Schema, error) {
	var value any
	if err := json.Unmarshal(sarifSchema, &value); err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(offlineLoader{})
	if err := compiler.AddResource("urn:sarif:2.1.0", value); err != nil {
		return nil, err
	}
	return compiler.Compile("urn:sarif:2.1.0")
})

// ValidateSARIF checks one report against the pinned SARIF 2.1.0 schema without
// reading any external resource. A valid document is well formed for consumers;
// it does not establish that a specific consumer imports it.
func ValidateSARIF(data []byte) error {
	schema, err := compiledSARIF()
	if err != nil {
		return err
	}
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("sarif report is not valid JSON: %w", err)
	}
	if err := schema.Validate(document); err != nil {
		return fmt.Errorf("sarif report does not match the %s schema: %w", SARIFSchemaVersion, err)
	}
	return nil
}
