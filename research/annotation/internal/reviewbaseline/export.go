package reviewbaseline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

type textRow struct {
	Page string `json:"page"`
	Unit int    `json:"unit"`
	Hash string `json:"text_sha256"`
	Text string `json:"text"`
}

type textExport struct {
	Version   int       `json:"version"`
	InputHash string    `json:"input_sha256"`
	Contract  string    `json:"unit_contract"`
	Rows      []textRow `json:"rows"`
}

// ExportText emits exact eligible text for a separately pinned research encoder.
// It verifies every source and binding before writing. Labels are not exported.
func ExportText(ctx context.Context, reader io.Reader, writer io.Writer) error {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(reader, (32<<20)+1))
	})
	if err != nil {
		return err
	}
	in, err := decode(data)
	if err != nil {
		return err
	}
	rows, _, err := prepare(ctx, in)
	if err != nil {
		return err
	}
	out := textExport{Version: 1, InputHash: fmt.Sprintf("%x", sha256.Sum256(data)), Contract: nlp.UnitContract}
	for _, row := range rows {
		out.Rows = append(out.Rows, textRow{Page: row.page, Unit: row.unit, Hash: row.textHash, Text: row.text})
	}
	_, err = commandio.Await(ctx, func() (struct{}, error) { return struct{}{}, json.NewEncoder(writer).Encode(out) })
	return err
}
