package server

import (
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell/document"
)

func toolSchemas() (*mcp.Tool, *mcp.Tool, error) {
	input, err := jsonschema.For[CheckInput](nil)
	if err != nil {
		return nil, nil, err
	}
	minimum, maximum := 1, 256
	sources := input.Properties["sources"]
	sources.MinItems, sources.MaxItems = &minimum, &maximum
	format := sources.Items.Properties["format"]
	for _, name := range document.Formats() {
		format.Enum = append(format.Enum, string(name))
	}
	description, err := jsonschema.For[Description](nil)
	if err != nil {
		return nil, nil, err
	}
	// A nil override map is part of the core policy's JSON representation.
	// The SDK's inferred map schema otherwise rejects that valid null value.
	languages := description.Properties["policy"].Properties["extraction"].Properties["languages"]
	languages.Type = ""
	languages.Types = []string{"object", "null"}
	check := tool("unswell_check", "Check supplied prose or source bytes with the fixed Unswell policy.")
	check.InputSchema = input
	describe := tool("unswell_describe", "Inspect supported formats, checked contexts, rules, and the effective policy.")
	describe.OutputSchema = description
	return check, describe, nil
}
