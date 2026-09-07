package report

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/stokaro/unswell"
)

//go:embed schema.json
var savedSchema []byte

var compiledSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	var value any
	if err := json.Unmarshal(savedSchema, &value); err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(offlineLoader{})
	if err := compiler.AddResource("urn:unswell:saved-result", value); err != nil {
		return nil, err
	}
	return compiler.Compile("urn:unswell:saved-result")
})

type offlineLoader struct{}

// Load refuses all external resources; the result schema is entirely embedded.
func (offlineLoader) Load(location string) (any, error) {
	return nil, fmt.Errorf("external schema resource is disabled: %s", location)
}

func validateSaved(data []byte, result unswell.RunResult) error {
	schema, err := compiledSchema()
	if err != nil {
		return err
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := schema.Validate(value); err != nil {
		return fmt.Errorf("invalid saved result: %w", err)
	}
	return validateCompletion(result)
}

func validateCompletion(result unswell.RunResult) error {
	complete := result.Status == "complete"
	if result.Manifest.Complete != complete || (complete && len(result.Errors) != 0) {
		return fmt.Errorf("inconsistent analysis completion")
	}
	if result.Gate.Passed && (!complete || (!result.Manifest.NoGate && len(result.Gate.Reasons) != 0)) {
		return fmt.Errorf("inconsistent policy decision")
	}
	return nil
}
