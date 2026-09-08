package baseline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Load decodes bounded JSON, rejecting duplicate keys, nulls, unknown fields,
// incompatible versions, ambiguous debt, and altered identity hashes.
func Load(ctx context.Context, data []byte) (File, error) {
	if err := ctx.Err(); err != nil {
		return File{}, err
	}
	if len(data) == 0 || len(data) > MaxBytes {
		return File{}, fmt.Errorf("baseline JSON must contain 1 to %d bytes", MaxBytes)
	}
	if err := checkJSON(ctx, json.NewDecoder(bytes.NewReader(data)), 0); err != nil {
		return File{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var file File
	if err := decoder.Decode(&file); err != nil {
		return File{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return File{}, fmt.Errorf("baseline must contain exactly one JSON object")
	}
	if _, err := fileIndex(ctx, file); err != nil {
		return File{}, err
	}
	return ordered(file), nil
}

// Encode validates and serializes owned, sorted copies, without timestamps.
func Encode(ctx context.Context, file File) ([]byte, error) {
	if _, err := fileIndex(ctx, file); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(ordered(file), "", "  ")
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(data)+1 > MaxBytes {
		return nil, fmt.Errorf("encoded baseline exceeds %d bytes", MaxBytes)
	}
	return append(data, '\n'), nil
}

func checkJSON(ctx context.Context, decoder *json.Decoder, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 32 {
		return fmt.Errorf("baseline JSON nesting exceeds 32")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return fmt.Errorf("baseline JSON cannot contain null")
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter == '{' {
		return checkJSONObject(ctx, decoder, depth)
	}
	if delimiter != '[' {
		return fmt.Errorf("unexpected baseline JSON delimiter")
	}
	for decoder.More() {
		if err := checkJSON(ctx, decoder, depth+1); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func checkJSONObject(ctx context.Context, decoder *json.Decoder, depth int) error {
	seen := make(map[string]bool)
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := key.(string)
		if !ok || seen[name] {
			return fmt.Errorf("duplicate or invalid baseline JSON key %q", key)
		}
		seen[name] = true
		if err := checkJSON(ctx, decoder, depth+1); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}
