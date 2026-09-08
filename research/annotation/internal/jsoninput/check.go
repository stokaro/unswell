// Package jsoninput rejects lossy or ambiguous research JSON before decoding.
package jsoninput

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Limits bounds collections during streaming validation, before typed allocation.
// Zero leaves a limit unset. Arrays overrides Array for a named object property;
// its names must be distinct under Unicode case folding.
// These are structural limits, not a promise about peak process memory.
type Limits struct {
	Array, Object int
	Arrays        map[string]int
}

// Check rejects duplicate keys, excessive depth, invalid Unicode, and trailing values.
func Check(ctx context.Context, data []byte, maximum int) error {
	return check(ctx, data, maximum, Limits{})
}

func check(ctx context.Context, data []byte, maximum int, limits Limits) error {
	if len(data) > maximum || !utf8.Valid(data) {
		return fmt.Errorf("annotation input exceeds %d bytes or is not UTF-8", maximum)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	c := checker{ctx: ctx, decoder: decoder, limits: limits}
	if err := c.value(0, ""); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("annotation input must contain one JSON value")
	}
	return validSurrogates(data)
}

// Decode checks collection bounds before allocating the typed destination and
// rejects unknown fields. Field-specific limits also cover case-insensitive aliases.
func Decode(ctx context.Context, data []byte, maximum int, target any, limits Limits) error {
	if err := check(ctx, data, maximum, limits); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

type checker struct {
	ctx     context.Context
	decoder *json.Decoder
	limits  Limits
}

func (c checker) value(depth int, field string) error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	if depth > 32 {
		return fmt.Errorf("annotation JSON nesting exceeds 32")
	}
	token, err := c.decoder.Token()
	if err != nil {
		return err
	}
	switch token {
	case json.Delim('{'):
		return c.object(depth)
	case json.Delim('['):
		return c.array(depth, field)
	}
	return nil
}

func (c checker) array(depth int, field string) error {
	maximum := c.limits.Array
	for name, limit := range c.limits.Arrays {
		if strings.EqualFold(field, name) {
			maximum = limit
			break
		}
	}
	for count := 0; c.decoder.More(); count++ {
		if maximum > 0 && count >= maximum {
			return fmt.Errorf("research JSON array %q exceeds %d entries", field, maximum)
		}
		if err := c.value(depth+1, ""); err != nil {
			return err
		}
	}
	_, err := c.decoder.Token()
	return err
}

func (c checker) object(depth int) error {
	seen := make(map[string]bool)
	for c.decoder.More() {
		if c.limits.Object > 0 && len(seen) >= c.limits.Object {
			return fmt.Errorf("research JSON object exceeds %d properties", c.limits.Object)
		}
		token, err := c.decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok || seen[key] {
			return fmt.Errorf("invalid or duplicate annotation JSON key %q", key)
		}
		seen[key] = true
		if err := c.value(depth+1, key); err != nil {
			return err
		}
	}
	_, err := c.decoder.Token()
	return err
}
