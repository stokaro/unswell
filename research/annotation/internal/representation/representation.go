// Package representation exports the engine's mapped units for research audits.
// It does not fit a classifier or turn editorial review labels into probabilities.
package representation

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

type inputPage struct {
	ID     string          `json:"id"`
	Format document.Format `json:"format"`
	Hash   string          `json:"sha256"`
	Text   string          `json:"text"`
}

type input struct {
	Version int         `json:"version"`
	Pages   []inputPage `json:"pages"`
}

type output struct {
	Status           string                             `json:"status"`
	Manifest         unswell.Manifest                   `json:"manifest"`
	Documents        []unswell.DocumentResult           `json:"documents"`
	Errors           []string                           `json:"errors"`
	PreparedFeatures *unswell.PreparedFeatureCollection `json:"prepared_features"`
	OriginalTokens   []tokenSource                      `json:"original_tokens"`
}

// Run exports mapped units from a supplied dataset after a complete engine run.
// Review metadata is ignored by inference and remains in the source-bound input.
func Run(ctx context.Context, reader io.Reader, writer io.Writer) error {
	in, err := decode(ctx, reader)
	if err != nil {
		return err
	}
	out, err := analyze(ctx, in)
	if err != nil {
		return err
	}
	_, err = commandio.Await(ctx, func() (struct{}, error) {
		return struct{}{}, json.NewEncoder(writer).Encode(out)
	})
	return err
}

func decode(ctx context.Context, reader io.Reader) (input, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(reader, (32<<20)+1))
	})
	if err != nil {
		return input{}, err
	}
	if len(data) > 32<<20 {
		return input{}, fmt.Errorf("representation input exceeds 32 MiB")
	}
	var in input
	if err := json.Unmarshal(data, &in); err != nil {
		return input{}, err
	}
	return in, nil
}

func analyze(ctx context.Context, in input) (output, error) {
	sources, err := inputSources(in)
	if err != nil {
		return output{}, err
	}
	provider, err := english.New()
	if err != nil {
		return output{}, err
	}
	engine, err := unswell.New(unswell.Options{
		Config:           []byte("version: 1\nextends: [builtin:technical-v1]\n"),
		Features:         []string{"prose-words"},
		PreparedFeatures: []string{"prose-words"},
		PreparedKinds:    []string{"sentence", "paragraph", "fragment"},
		Jobs:             1,
		NLP:              provider,
	})
	if err != nil {
		return output{}, err
	}
	result, err := engine.AnalyzeAll(ctx, sources)
	if err != nil {
		return output{}, err
	}
	if !complete(result) {
		return output{}, fmt.Errorf("representation analysis is incomplete")
	}
	tokens, err := originalTokens(ctx, sources, engine, provider, result.Features)
	if err != nil {
		return output{}, err
	}
	return output{Status: result.Status, Manifest: result.Manifest, Documents: result.Documents,
		Errors: []string{}, PreparedFeatures: result.PreparedFeatures, OriginalTokens: tokens}, nil
}

func complete(result unswell.RunResult) bool {
	return result.Status == "complete" && result.Manifest.Complete && len(result.Errors) == 0 &&
		len(result.Abstentions) == 0 && len(result.Manifest.SkippedRules) == 0
}

func inputSources(in input) ([]document.Source, error) {
	if in.Version != 1 || len(in.Pages) == 0 || len(in.Pages) > 512 {
		return nil, fmt.Errorf("representation input requires version 1 and 1..512 pages")
	}
	seen := make(map[string]bool)
	sources := make([]document.Source, 0, len(in.Pages))
	for _, page := range in.Pages {
		if page.ID == "" || seen[page.ID] {
			return nil, fmt.Errorf("representation page IDs must be nonempty and unique")
		}
		seen[page.ID] = true
		if page.Hash != fmt.Sprintf("%x", sha256.Sum256([]byte(page.Text))) {
			return nil, fmt.Errorf("source hash mismatch for %s", page.ID)
		}
		sources = append(sources, document.Source{Name: page.ID, Format: page.Format, Bytes: []byte(page.Text)})
	}
	return sources, nil
}
