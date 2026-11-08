package reviewbaseline

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

type inventoryInput struct {
	Version int             `json:"version"`
	Basis   string          `json:"basis"`
	Pages   []inventoryPage `json:"pages"`
}

type inventoryPage struct {
	ID     string          `json:"id"`
	Format document.Format `json:"format"`
	Hash   string          `json:"sha256"`
	Text   string          `json:"text"`
}

type inventoryOutput struct {
	Version   int                   `json:"version"`
	Basis     string                `json:"basis"`
	InputHash string                `json:"input_sha256"`
	Contract  string                `json:"unit_contract"`
	Pages     []inventoryOutputPage `json:"pages"`
}

type inventoryOutputPage struct {
	ID       string                `json:"id"`
	Format   document.Format       `json:"format"`
	Hash     string                `json:"sha256"`
	Bytes    int                   `json:"bytes"`
	Excluded []document.Exclusion  `json:"excluded"`
	Units    []inventoryOutputUnit `json:"units"`
}

type inventoryOutputUnit struct {
	Ordinal  int             `json:"ordinal"`
	Binding  nlp.UnitBinding `json:"binding"`
	Text     string          `json:"text"`
	RawSpan  document.Span   `json:"raw_span"`
	RawSlice string          `json:"raw_slice"`
}

// ExportSourceInventory emits source-bound prose units from pinned, arbitrary
// pages. A sentence is selected where possible; protected boundaries produce
// fragments. It writes nothing until every page and source map is verified.
func ExportSourceInventory(ctx context.Context, reader io.Reader, writer io.Writer) error {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(reader, (32<<20)+1))
	})
	if err != nil {
		return err
	}
	in, err := decodeInventory(data)
	if err != nil {
		return err
	}
	provider, err := english.New()
	if err != nil {
		return err
	}
	engine, err := unswell.New(unswell.Options{Config: []byte("version: 1\nextends: [builtin:technical-v1]\n"), NLP: provider})
	if err != nil {
		return err
	}
	out := inventoryOutput{Version: 1, Basis: in.Basis,
		InputHash: fmt.Sprintf("%x", sha256.Sum256(data)), Contract: nlp.UnitContract,
		Pages: make([]inventoryOutputPage, 0, len(in.Pages))}
	for _, page := range in.Pages {
		prepared, err := prepareInventoryPage(ctx, page, engine, provider)
		if err != nil {
			return fmt.Errorf("%s: %w", page.ID, err)
		}
		out.Pages = append(out.Pages, prepared)
	}
	_, err = commandio.Await(ctx, func() (struct{}, error) { return struct{}{}, json.NewEncoder(writer).Encode(out) })
	return err
}

func decodeInventory(data []byte) (inventoryInput, error) {
	if len(data) > 32<<20 {
		return inventoryInput{}, fmt.Errorf("inventory input exceeds 32 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var in inventoryInput
	if err := decoder.Decode(&in); err != nil {
		return inventoryInput{}, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return inventoryInput{}, fmt.Errorf("inventory input must contain one JSON value")
	}
	if in.Version != 1 || in.Basis != "source-function-inventory" || len(in.Pages) < 1 || len(in.Pages) > 16 {
		return inventoryInput{}, fmt.Errorf("expected version 1 source-function-inventory with 1..16 pages")
	}
	return in, validateInventoryPages(in.Pages)
}

func validateInventoryPages(pages []inventoryPage) error {
	seen := make(map[string]bool, len(pages))
	for _, page := range pages {
		if page.ID == "" || seen[page.ID] {
			return fmt.Errorf("missing or duplicate inventory page ID")
		}
		seen[page.ID] = true
		if page.Hash != fmt.Sprintf("%x", sha256.Sum256([]byte(page.Text))) {
			return fmt.Errorf("source hash mismatch: %s", page.ID)
		}
	}
	return nil
}

func prepareInventoryPage(ctx context.Context, page inventoryPage,
	engine *unswell.Engine, provider nlp.Provider,
) (inventoryOutputPage, error) {
	policy, err := engine.ResolvedPolicy(page.ID)
	if err != nil {
		return inventoryOutputPage{}, err
	}
	doc, err := extract.Parse(ctx, document.Source{Name: page.ID, Format: page.Format, Bytes: []byte(page.Text)}, extract.Options{
		IncludeStructure: true, IncludeQuotes: policy.Analysis.IncludeQuotes,
		MaxBytes: policy.Analysis.MaxFileBytes, MaxBlocks: policy.Analysis.MaxBlocks, Policy: policy.Extraction})
	if err != nil {
		return inventoryOutputPage{}, err
	}
	options := nlp.UnitOptions{Kinds: []string{"sentence", "fragment"},
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences},
		Limits: nlp.UnitLimits{MaxBytes: policy.Analysis.MaxFileBytes, MaxContextBytes: 65536,
			MaxUnits: 10000, MaxTokens: min(policy.Analysis.MaxTokens, 1000000), MaxSegments: 4096}}
	out := inventoryOutputPage{ID: page.ID, Format: page.Format, Hash: doc.Hash, Bytes: len(doc.Source),
		Excluded: doc.Excluded, Units: []inventoryOutputUnit{}}
	for _, block := range doc.Blocks {
		units, err := nlp.PrepareUnits(ctx, block, provider, options)
		if err != nil {
			return inventoryOutputPage{}, err
		}
		for _, unit := range units {
			binding := unit.Binding()
			bounds := document.Bounds(binding.Segments)
			if !bounds.Valid(len(doc.Source)) {
				return inventoryOutputPage{}, fmt.Errorf("invalid source span at unit %d", len(out.Units))
			}
			out.Units = append(out.Units, inventoryOutputUnit{Ordinal: len(out.Units), Binding: binding,
				Text: unit.Block().Text, RawSpan: bounds, RawSlice: string(doc.Source[bounds.Start:bounds.End])})
			if len(out.Units) > 10000 {
				return inventoryOutputPage{}, fmt.Errorf("inventory exceeds 10000 units")
			}
		}
	}
	return out, ctx.Err()
}
