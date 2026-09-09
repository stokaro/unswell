package feature

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/stokaro/unswell/nlp"
)

// CompressionContract fixes framing, UTF-8 bytes, separators, and delta formulas.
const CompressionContract = "unswell-compression-increments-v1"

// CompressionOptions pins a level (0..9) and a 1..1MiB input-byte work budget.
// Level zero is a stored-block numerical control, not the default compression level.
type CompressionOptions struct {
	Level         int `json:"level"`
	MaxInputBytes int `json:"max_input_bytes"`
}

// CompressionIdentity binds an explicit reference and the actual compressor.
// ReferenceSHA256 includes the appended LF. It does not attest corpus permissions.
type CompressionIdentity struct {
	Contract        string             `json:"contract"`
	GoVersion       string             `json:"go_version"`
	ReferenceSHA256 string             `json:"reference_sha256"`
	ReferenceBytes  int                `json:"reference_bytes"`
	Options         CompressionOptions `json:"options"`
}

// Compression owns an immutable reference prefix and its baseline size.
// Callers establish the reference's training split, order, and permitted uses.
type Compression struct {
	prefix   string
	identity CompressionIdentity
	seedSize int
	baseSize int
}

// CompressionCounts retains byte sizes before any ratio or boundary abstention.
type CompressionCounts struct {
	TargetBytes        int `json:"target_bytes"`
	SeedBytes          int `json:"seed_compressed_bytes"`
	SeedTargetBytes    int `json:"seed_target_compressed_bytes"`
	ControlBytes       int `json:"control_compressed_bytes"`
	ControlTargetBytes int `json:"control_target_compressed_bytes"`
}

// CompressionResult owns values and complete target bindings, without source text.
// Hash binds target/context, source/policy/NLP identities, reference, and formulas.
type CompressionResult struct {
	Hash     string              `json:"hash"`
	Identity CompressionIdentity `json:"identity"`
	Binding  nlp.UnitBinding     `json:"binding"`
	Counts   CompressionCounts   `json:"counts"`
	Values   []Value             `json:"values"`
}

// NewCompression validates an explicit UTF-8 reference and compresses its baseline.
// The appended LF counts toward the 32KiB prefix limit. Nothing is truncated.
func NewCompression(ctx context.Context, reference string, options CompressionOptions) (*Compression, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	prefix, err := compressionPrefix(reference, options)
	if err != nil {
		return nil, err
	}
	sizer, err := newCompressionSizer(ctx, options.Level)
	if err != nil {
		return nil, err
	}
	seed, err := sizer.size(prefix)
	if err != nil {
		return nil, err
	}
	control, err := sizer.size("\n")
	if err != nil {
		return nil, err
	}
	identity := CompressionIdentity{Contract: CompressionContract, GoVersion: runtime.Version(),
		ReferenceSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(prefix))), ReferenceBytes: len(prefix), Options: options}
	return &Compression{prefix: prefix, identity: identity, seedSize: seed, baseSize: control}, nil
}

// Identity returns the effective reference and compressor identity by value.
func (m *Compression) Identity() CompressionIdentity {
	if m == nil {
		return CompressionIdentity{}
	}
	return m.identity
}

// Measure uses one shared prepared target with its original source mapping.
// Missing targets, incompatible NLP, invalid limits, and cancellation are errors.
// Nonpositive compressed-size increments have no numeric feature value.
func (m *Compression) Measure(ctx context.Context, unit nlp.PreparedUnit, identity Identity, limits Limits) (CompressionResult, error) {
	if err := ctx.Err(); err != nil {
		return CompressionResult{}, err
	}
	if m == nil || m.prefix == "" {
		return CompressionResult{}, fmt.Errorf("compression requires an initialized reference")
	}
	text, err := compressionTarget(ctx, unit, identity, limits)
	if err != nil {
		return CompressionResult{}, err
	}
	counts, err := m.counts(ctx, text)
	if err != nil {
		return CompressionResult{}, err
	}
	result := CompressionResult{Identity: m.identity, Binding: unit.Binding(), Counts: counts, Values: compressionValues(counts)}
	identity.Capabilities = orderedCapabilities(identity.Capabilities)
	identity.NLP.Capabilities = orderedCapabilities(identity.NLP.Capabilities)
	data, err := json.Marshal(struct {
		Result CompressionResult
		Inputs Identity
	}{result, identity})
	if err != nil {
		return CompressionResult{}, err
	}
	result.Hash = fmt.Sprintf("%x", sha256.Sum256(data))
	if err := ctx.Err(); err != nil {
		return CompressionResult{}, err
	}
	return result, nil
}
