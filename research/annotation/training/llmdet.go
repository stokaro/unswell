package training

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// LLMDetPackVersion identifies a manifest of local LLMDet dictionary tables
// and the tokenizer files that produce their token IDs.
const LLMDetPackVersion = "unswell-llmdet-pack-v1"

// MaxLLMDetPackBytes bounds a pack manifest.
const MaxLLMDetPackBytes = 1 << 20

// LLMDetModel names one model's table and tokenizer files, relative to the
// manifest, with the digest of each.
type LLMDetModel struct {
	Model            string `json:"model"`
	Table            string `json:"table"`
	TableSHA256      string `json:"table_sha256"`
	Vocabulary       string `json:"vocabulary"`
	VocabularySHA256 string `json:"vocabulary_sha256"`
	Merges           string `json:"merges"`
	MergesSHA256     string `json:"merges_sha256"`
}

// LLMDetPack is a manifest of dictionary tables. Its digest covers the model
// list and every file digest, so an artifact fitted on it names the exact
// tables and tokenizers. The tables themselves stay outside the repository.
type LLMDetPack struct {
	Version string        `json:"version"`
	SHA256  string        `json:"sha256,omitempty"`
	Models  []LLMDetModel `json:"models"`
	root    string
}

// LoadLLMDetPack reads a manifest, resolves its files against the manifest's
// directory, and checks every file digest before any table is read.
func LoadLLMDetPack(ctx context.Context, path string) (LLMDetPack, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- The operator explicitly selects the local manifest.
	if err != nil {
		return LLMDetPack{}, err
	}
	var pack LLMDetPack
	if err := jsoninput.Decode(ctx, data, MaxLLMDetPackBytes, &pack, jsoninput.Limits{Array: 16, Object: 8}); err != nil {
		return LLMDetPack{}, err
	}
	if pack.Version != LLMDetPackVersion || len(pack.Models) == 0 || len(pack.Models) > 16 {
		return LLMDetPack{}, fmt.Errorf("LLMDet pack requires the pack version and 1 to 16 models")
	}
	pack.root = filepath.Dir(path)
	if err := pack.verifyModels(ctx); err != nil {
		return LLMDetPack{}, err
	}
	want := pack.SHA256
	pack.SHA256 = ""
	encoded, err := json.Marshal(pack)
	if err != nil {
		return LLMDetPack{}, err
	}
	sum := sha256.Sum256(encoded)
	pack.SHA256 = hex.EncodeToString(sum[:])
	if want != "" && want != pack.SHA256 {
		return LLMDetPack{}, fmt.Errorf("LLMDet pack digest mismatch")
	}
	return pack, nil
}

func (p LLMDetPack) verifyModels(ctx context.Context) error {
	seen := make(map[string]bool, len(p.Models))
	for _, model := range p.Models {
		if model.Model == "" || seen[model.Model] {
			return fmt.Errorf("LLMDet pack models require distinct names")
		}
		seen[model.Model] = true
		for _, file := range []struct{ name, digest string }{{model.Table, model.TableSHA256},
			{model.Vocabulary, model.VocabularySHA256}, {model.Merges, model.MergesSHA256}} {
			if err := p.verify(ctx, file.name, file.digest); err != nil {
				return fmt.Errorf("LLMDet model %s: %w", model.Model, err)
			}
		}
	}
	return nil
}

func (p LLMDetPack) path(name string) (string, error) {
	if name == "" || filepath.IsAbs(name) || !filepath.IsLocal(name) {
		return "", fmt.Errorf("pack file %q must be a local relative path", name)
	}
	return filepath.Join(p.root, name), nil
}

func (p LLMDetPack) verify(ctx context.Context, name, digest string) error {
	if !validDigest(digest) {
		return fmt.Errorf("file %q requires a SHA-256 digest", name)
	}
	resolved, err := p.path(name)
	if err != nil {
		return err
	}
	file, err := os.Open(resolved) // #nosec G304 -- The manifest the operator selected names this local file.
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err := errors.Join(err, file.Close()); err != nil {
		return err
	}
	if hex.EncodeToString(hash.Sum(nil)) != digest {
		return fmt.Errorf("file %q does not match its digest", name)
	}
	return ctx.Err()
}

func (p LLMDetPack) read(name string) ([]byte, error) {
	resolved, err := p.path(name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(resolved) // #nosec G304 -- The manifest the operator selected names this local file.
}

// RunLLMDet fits the LLMDet baseline. Every target of the fitted kind is
// tokenized with each model's tokenizer and measured against that model's
// dictionary table. Each model gives a proxy perplexity and a context
// coverage. Options.Features may add prepared features to the same rows.
func RunLLMDet(ctx context.Context, candidates corpus.Artifact, round *annotation.Round, files map[string][]byte,
	options Options, pack LLMDetPack,
) (Artifact, error) {
	if err := llmdetGuards(ctx, candidates, options, pack); err != nil {
		return Artifact{}, err
	}
	joined, decisions, err := referenceRoundInputs(ctx, candidates, round, files, options.Features)
	if err != nil {
		return Artifact{}, err
	}
	return fitReference(ctx, candidates, files, decisions, options, llmdetSource(options.Kind, pack), joined)
}

// RunLLMDetDecisions fits the LLMDet baseline from a prepared decision set,
// such as the cohort labels of the pattern protocol.
func RunLLMDetDecisions(ctx context.Context, candidates corpus.Artifact, decisions annotation.DecisionSet,
	files map[string][]byte, options Options, pack LLMDetPack,
) (Artifact, error) {
	if err := llmdetGuards(ctx, candidates, options, pack); err != nil {
		return Artifact{}, err
	}
	joined, err := referenceDecisionInputs(ctx, candidates, decisions, files, options.Features)
	if err != nil {
		return Artifact{}, err
	}
	return fitReference(ctx, candidates, files, decisions, options, llmdetSource(options.Kind, pack), joined)
}

func llmdetGuards(ctx context.Context, candidates corpus.Artifact, options Options, pack LLMDetPack) error {
	if pack.Version != LLMDetPackVersion || !validDigest(pack.SHA256) || len(pack.Models) == 0 {
		return fmt.Errorf("LLMDet training requires a loaded table pack")
	}
	if err := validateReferenceFeatures(options.Features); err != nil {
		return err
	}
	if err := validateOptions(ctx, candidates, options); err != nil {
		return err
	}
	return zeroPolicyForRules(options, "llmdet_tables")
}

func validateLLMDetArtifact(a Artifact) error {
	if a.Identity.FeatureSource != "llmdet_tables" {
		if a.Identity.LLMDetPackHash != "" {
			return fmt.Errorf("artifacts without LLMDet tables cannot name a table pack")
		}
		return nil
	}
	if !validDigest(a.Identity.LLMDetPackHash) || a.Identity.Context != "prepared_piece" ||
		a.Identity.Preprocessing != llmdetPreprocessing {
		return fmt.Errorf("LLMDet artifact representation mismatch")
	}
	references := referenceColumns(a)
	if references == 0 || references%2 != 0 {
		return fmt.Errorf("LLMDet artifact requires paired reference columns")
	}
	return nil
}

// ModelNames lists the pack's models in manifest order.
func (p LLMDetPack) ModelNames() []string {
	names := make([]string, 0, len(p.Models))
	for _, model := range p.Models {
		names = append(names, model.Model)
	}
	return slices.Clone(names)
}
