package jsoninput

import (
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Schema validates bytes against a fixed local JSON Schema resource.
func Schema(data, schemaBytes []byte, identity string) error {
	var schemaValue, value any
	if err := json.Unmarshal(schemaBytes, &schemaValue); err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(identity, schemaValue); err != nil {
		return err
	}
	schema, err := compiler.Compile(identity)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := schema.Validate(value); err != nil {
		return err
	}
	return nil
}
