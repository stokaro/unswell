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
	minimum, maximum := 1, MaxSources
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

func reviewSchemas() (*mcp.Tool, *mcp.Tool, error) {
	requestInput, err := jsonschema.For[ReviewRequestInput](nil)
	if err != nil {
		return nil, nil, err
	}
	submitInput, err := jsonschema.For[ReviewSubmitInput](nil)
	if err != nil {
		return nil, nil, err
	}
	for _, input := range []*jsonschema.Schema{requestInput, submitInput} {
		format := input.Properties["source"].Properties["format"]
		for _, name := range document.Formats() {
			format.Enum = append(format.Enum, string(name))
		}
	}
	maximum := MaxReviewFlags
	flags := submitInput.Properties["flags"]
	flags.MaxItems = &maximum
	categories := flags.Items.Properties["categories"]
	for _, name := range ReviewCategories() {
		categories.Items.Enum = append(categories.Items.Enum, name)
	}
	request := tool("unswell_review_request",
		"List the reviewable items of one document with the editorial rubric for the calling assistant.")
	request.InputSchema = requestInput
	submit := tool("unswell_review_submit",
		"Bind the calling assistant's review flags to exact source coordinates. The result is advisory.")
	submit.InputSchema = submitInput
	return request, submit, nil
}
