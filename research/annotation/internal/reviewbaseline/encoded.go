package reviewbaseline

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

type encodedRow struct {
	Page   string    `json:"page"`
	Unit   int       `json:"unit"`
	Hash   string    `json:"text_sha256"`
	Tokens int       `json:"tokens"`
	Reason string    `json:"reason,omitempty"`
	Values []float64 `json:"values"`
}

type encodings struct {
	Version     int          `json:"version"`
	InputHash   string       `json:"input_sha256"`
	EncoderHash string       `json:"encoder_sha256"`
	Rows        []encodedRow `json:"rows"`
}

type encodedPrediction struct {
	prediction
	Score  *float64 `json:"raw_score"`
	Reason string   `json:"reason,omitempty"`
}

type encodedFit struct {
	fitted
	Predictions []encodedPrediction `json:"predictions"`
}

type encodedOutput struct {
	output
	EncoderHash    string       `json:"encoder_sha256"`
	EmbeddingsHash string       `json:"embeddings_sha256"`
	Models         []encodedFit `json:"models"`
}

// RunEncoded compares fixed encoder features with optional shared numeric features.
// The sidecar must describe every verified target, including unavailable targets.
// This research tool neither loads an encoder nor creates product model packs.
func RunEncoded(ctx context.Context, reader, vectors io.Reader, writer io.Writer) error {
	data, err := commandio.Await(ctx, func() ([]byte, error) { return io.ReadAll(io.LimitReader(reader, (32<<20)+1)) })
	if err != nil {
		return err
	}
	in, err := decode(data)
	if err != nil {
		return err
	}
	rows, provider, err := prepare(ctx, in)
	if err != nil {
		return err
	}
	inputHash := fmt.Sprintf("%x", sha256.Sum256(data))
	enc, vectorHash, err := readEncodings(ctx, vectors, rows, inputHash)
	if err != nil {
		return err
	}
	out := encodedOutput{output: output{Version: 3, Basis: "exposed-assistant-development; frozen encoder comparison",
		InputHash: inputHash, Provider: provider, Rows: len(rows)}, EncoderHash: enc.EncoderHash, EmbeddingsHash: vectorHash}
	initializeEncodedOutput(&out)
	for fold := range 5 {
		for _, kind := range []string{"E", "ES"} {
			model, err := fitEncoded(ctx, rows, enc.Rows, fold, kind)
			if err != nil {
				return fmt.Errorf("fold %d model %s: %w", fold, kind, err)
			}
			out.Models = append(out.Models, model)
		}
	}
	_, err = commandio.Await(ctx, func() (struct{}, error) { return struct{}{}, json.NewEncoder(writer).Encode(out) })
	return err
}

func readEncodings(ctx context.Context, reader io.Reader, rows []row, inputHash string) (encodings, string, error) {
	raw, err := commandio.Await(ctx, func() ([]byte, error) { return io.ReadAll(io.LimitReader(reader, (128<<20)+1)) })
	if err != nil {
		return encodings{}, "", err
	}
	if len(raw) > 128<<20 {
		return encodings{}, "", fmt.Errorf("encodings exceed 128 MiB")
	}
	enc, err := decodeEncodings(raw, inputHash, len(rows))
	if err != nil {
		return encodings{}, "", err
	}
	for i, r := range rows {
		if err := ctx.Err(); err != nil {
			return encodings{}, "", err
		}
		if err := validateEncoding(r, enc.Rows[i]); err != nil {
			return encodings{}, "", fmt.Errorf("encoding %d: %w", i, err)
		}
	}
	return enc, fmt.Sprintf("%x", sha256.Sum256(raw)), nil
}

func decodeEncodings(raw []byte, inputHash string, count int) (encodings, error) {
	var enc encodings
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&enc); err != nil {
		return encodings{}, err
	}
	if err := d.Decode(new(any)); !errors.Is(err, io.EOF) {
		return encodings{}, fmt.Errorf("encodings must contain one JSON value")
	}
	identity, err := hex.DecodeString(enc.EncoderHash)
	if err != nil || len(identity) != 32 || enc.Version != 1 || enc.InputHash != inputHash || len(enc.Rows) != count {
		return encodings{}, fmt.Errorf("encoding identity or row count mismatch")
	}
	return enc, nil
}

func validateEncoding(r row, e encodedRow) error {
	if e.Page != r.page || e.Unit != r.unit || e.Hash != r.textHash {
		return fmt.Errorf("target binding mismatch")
	}
	if e.Tokens > 256 {
		if e.Reason != "sequence_limit" || e.Values != nil {
			return fmt.Errorf("long target must be unavailable")
		}
		return nil
	}
	if e.Tokens < 3 || e.Reason != "" || len(e.Values) != 128 {
		return fmt.Errorf("invalid available vector")
	}
	return validateUnitVector(e.Values)
}

func validateUnitVector(values []float64) error {
	var norm float64
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("nonfinite vector")
		}
		norm += v * v
	}
	if math.Abs(norm-1) > 1e-4 {
		return fmt.Errorf("vector is not L2-normalized")
	}
	return nil
}
