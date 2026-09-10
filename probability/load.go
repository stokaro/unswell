package probability

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/nlp"
)

// Load decodes bounded JSON, rejecting duplicate keys, nulls, unknown fields,
// unsupported contracts, inconsistent declarations, and an altered digest.
// A corrupt or incompatible pack is an error, never a silent abstention.
func Load(ctx context.Context, data []byte) (*Pack, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > MaxBytes {
		return nil, fmt.Errorf("probability pack JSON must contain 1 to %d bytes", MaxBytes)
	}
	if err := checkJSON(ctx, json.NewDecoder(bytes.NewReader(data)), 0); err != nil {
		return nil, err
	}
	file, err := decodeFile(data)
	if err != nil {
		return nil, err
	}
	if err := validate(file); err != nil {
		return nil, err
	}
	return compile(ctx, file)
}

// Digest returns the pack digest of a validated file with its own hash omitted.
// It identifies bytes and declarations; it is not an attestation.
func Digest(file File) (string, error) {
	file.SHA256 = ""
	data, err := json.Marshal(file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func decodeFile(data []byte) (File, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var file File
	if err := decoder.Decode(&file); err != nil {
		return File{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return File{}, fmt.Errorf("probability pack must contain exactly one JSON object")
	}
	return file, nil
}

func compile(ctx context.Context, file File) (*Pack, error) {
	classifier, err := model.NewLogistic(model.Parameters{Means: file.Logistic.Means, Scales: file.Logistic.Scales,
		Weights: file.Logistic.Weights, Intercept: file.Logistic.Intercept})
	if err != nil {
		return nil, err
	}
	calibration, err := model.NewIsotonic(model.IsotonicParameters{Scores: file.Calibration.Scores,
		Responses: file.Calibration.Responses})
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &Pack{file: cloneFile(file), classifier: classifier, calibration: calibration}, nil
}

// File returns a detached copy of the validated declarations and parameters.
func (p *Pack) File() File { return cloneFile(p.file) }

// SHA256 returns the verified pack digest.
func (p *Pack) SHA256() string { return p.file.SHA256 }

// Kind returns the single unit kind this pack estimates.
func (p *Pack) Kind() string { return p.file.Kind }

// Accepted reports the author's acceptance declaration. This package cannot
// verify a corpus, an evaluation, or a product decision behind it.
func (p *Pack) Accepted() bool { return p.file.DeclaredStatus == "accepted" }

// Columns returns the required measurements in exact evaluation order.
func (p *Pack) Columns() []feature.Descriptor { return cloneColumns(p.file.Contract.Columns) }

func cloneFile(file File) File {
	file.Contract.Columns = cloneColumns(file.Contract.Columns)
	file.Contract.Capabilities = slices.Clone(file.Contract.Capabilities)
	file.Contract.NLP.Capabilities = slices.Clone(file.Contract.NLP.Capabilities)
	file.Logistic = &Logistic{Means: slices.Clone(file.Logistic.Means), Scales: slices.Clone(file.Logistic.Scales),
		Weights: slices.Clone(file.Logistic.Weights), Intercept: file.Logistic.Intercept}
	file.Calibration.Scores = slices.Clone(file.Calibration.Scores)
	file.Calibration.Responses = slices.Clone(file.Calibration.Responses)
	return file
}

func cloneColumns(columns []feature.Descriptor) []feature.Descriptor {
	result := slices.Clone(columns)
	for i := range result {
		result[i].Requires = slices.Clone(result[i].Requires)
	}
	return result
}

func capabilityCovered(available []nlp.Capability, required nlp.Capability) bool {
	return slices.Contains(available, required)
}

func checkJSON(ctx context.Context, decoder *json.Decoder, depth int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if depth > 32 {
		return fmt.Errorf("probability pack JSON nesting exceeds 32")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return fmt.Errorf("probability pack JSON cannot contain null")
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter == '{' {
		return checkJSONObject(ctx, decoder, depth)
	}
	if delimiter != '[' {
		return fmt.Errorf("unexpected probability pack JSON delimiter")
	}
	return checkJSONArray(ctx, decoder, depth)
}

func checkJSONArray(ctx context.Context, decoder *json.Decoder, depth int) error {
	for decoder.More() {
		if err := checkJSON(ctx, decoder, depth+1); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
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
			return fmt.Errorf("duplicate or invalid probability pack JSON key %q", key)
		}
		seen[name] = true
		if err := checkJSON(ctx, decoder, depth+1); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}
