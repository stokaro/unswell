package annotation

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// MaxBytes bounds an administrative round before JSON decoding.
const MaxBytes = 32 << 20

//go:embed schema.json
var schemaBytes []byte

// Load checks the versioned schema and cross-record invariants without I/O.
// It does not attest to the truth of provenance or human-participation declarations.
func Load(ctx context.Context, data []byte) (*Round, error) {
	if len(data) > MaxBytes || !utf8.Valid(data) {
		return nil, fmt.Errorf("annotation input exceeds %d bytes or is not UTF-8", MaxBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueValue(ctx, decoder, 0); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("annotation input must contain one JSON value")
	}
	if err := validSurrogates(data); err != nil {
		return nil, err
	}
	if err := validateSchema(data); err != nil {
		return nil, err
	}
	var round Round
	if err := json.Unmarshal(data, &round.data); err != nil {
		return nil, err
	}
	if err := round.data.validate(ctx); err != nil {
		return nil, err
	}
	return &round, nil
}

func validateSchema(data []byte) error {
	var schemaValue, value any
	if err := json.Unmarshal(schemaBytes, &schemaValue); err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("urn:unswell:annotation:v1", schemaValue); err != nil {
		return err
	}
	schema, err := compiler.Compile("urn:unswell:annotation:v1")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if err := schema.Validate(value); err != nil {
		return fmt.Errorf("annotation schema: %w", err)
	}
	return nil
}

func uniqueValue(ctx context.Context, decoder *json.Decoder, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 32 {
		return fmt.Errorf("annotation JSON nesting exceeds 32")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	switch token {
	case json.Delim('{'):
		return uniqueObject(ctx, decoder, depth)
	case json.Delim('['):
		for decoder.More() {
			if err := uniqueValue(ctx, decoder, depth+1); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
	}
	return err
}

func uniqueObject(ctx context.Context, decoder *json.Decoder, depth int) error {
	seen := make(map[string]bool)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok || seen[key] {
			return fmt.Errorf("invalid or duplicate annotation JSON key %q", key)
		}
		seen[key] = true
		if err := uniqueValue(ctx, decoder, depth+1); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}
