// Package jsoninput rejects lossy or ambiguous research JSON before decoding.
package jsoninput

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

// Check rejects duplicate keys, excessive depth, invalid Unicode, and trailing values.
func Check(ctx context.Context, data []byte, maximum int) error {
	if len(data) > maximum || !utf8.Valid(data) {
		return fmt.Errorf("annotation input exceeds %d bytes or is not UTF-8", maximum)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueValue(ctx, decoder, 0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("annotation input must contain one JSON value")
	}
	return validSurrogates(data)
}

// Decode also rejects unknown fields throughout a typed destination.
func Decode(ctx context.Context, data []byte, maximum int, target any) error {
	if err := Check(ctx, data, maximum); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
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
