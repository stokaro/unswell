package reviewbaseline

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

// Run fits the fixed development comparison and writes scores after all fits succeed.
// The input supplies explicit groups and source-bound assistant labels, not human qualification.
func Run(ctx context.Context, reader io.Reader, writer io.Writer) error {
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
	out := output{Version: 1, Basis: "exposed-assistant-development; unit proposals only",
		InputHash: fmt.Sprintf("%x", sha256.Sum256(data)), ModelAlgorithm: model.Algorithm,
		UnitContract: nlp.UnitContract, FeatureContract: feature.UnitContract,
		LexicalContract: feature.LexicalCountContract, Provider: provider, Rows: len(rows)}
	for fold := range 5 {
		for _, kind := range []string{"L", "S", "W", "SW"} {
			result, err := fit(ctx, rows, fold, kind)
			if err != nil {
				return fmt.Errorf("fold %d model %s: %w", fold, kind, err)
			}
			out.Models = append(out.Models, result)
		}
	}
	_, err = commandio.Await(ctx, func() (struct{}, error) { return struct{}{}, json.NewEncoder(writer).Encode(out) })
	return err
}

func decode(data []byte) (input, error) {
	if len(data) > 32<<20 {
		return input{}, fmt.Errorf("baseline input exceeds 32 MiB")
	}
	var in input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&in); err != nil {
		return input{}, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return input{}, fmt.Errorf("baseline input must contain one JSON value")
	}
	if in.Version != 1 || in.Basis != "exposed-assistant-unit-selection" || len(in.Pages) == 0 || len(in.Pages) > 512 {
		return input{}, fmt.Errorf("expected version 1 exposed-assistant input with 1..512 pages")
	}
	slices.SortFunc(in.Pages, func(a, b page) int { return strings.Compare(a.ID, b.ID) })
	return in, validatePages(in.Pages)
}

func validatePages(pages []page) error {
	ids, sources := make(map[string]bool), make(map[string]bool)
	groups := make(map[string]int)
	units := 0
	for _, page := range pages {
		if invalidIdentity(page) || ids[page.ID] || sources[page.Source] {
			return fmt.Errorf("duplicate or invalid page, source, group or fold")
		}
		if fold, ok := groups[page.Group]; ok && fold != page.Fold {
			return fmt.Errorf("one source group crosses folds")
		}
		groups[page.Group], ids[page.ID], sources[page.Source] = page.Fold, true, true
		if page.Hash != fmt.Sprintf("%x", sha256.Sum256([]byte(page.Text))) {
			return fmt.Errorf("source hash mismatch: %s", page.ID)
		}
		units += len(page.Units)
		if units > 100000 {
			return fmt.Errorf("baseline input exceeds 100000 prepared units")
		}
	}
	return nil
}

func invalidIdentity(page page) bool {
	return page.ID == "" || page.Source == "" || page.Group == "" || page.Fold < 0 || page.Fold >= 5
}
