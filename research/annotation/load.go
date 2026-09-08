package annotation

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// MaxBytes bounds an administrative round before JSON decoding.
const MaxBytes = 32 << 20

//go:embed schema.json
var schemaBytes []byte

// Load checks the versioned schema and cross-record invariants without I/O.
// It does not attest to the truth of provenance or human-participation declarations.
func Load(ctx context.Context, data []byte) (*Round, error) {
	if err := jsoninput.Check(ctx, data, MaxBytes); err != nil {
		return nil, err
	}
	if err := jsoninput.Schema(data, schemaBytes, "urn:unswell:annotation:v1"); err != nil {
		return nil, fmt.Errorf("annotation schema: %w", err)
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
