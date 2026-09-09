package training

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/nlp"
)

// LexicalOptions defines training-only vocabulary selection. MaxFeatures shares
// the numerical model's 128-column budget. MinTargets counts distinct fitted
// training targets, not independent documents. Counts use log1p before scaling.
type LexicalOptions struct {
	Counts      feature.LexicalOptions `json:"counts"`
	MaxFeatures int                    `json:"max_features"`
	MinTargets  int                    `json:"min_targets"`
}

// VocabularyTerm binds one explicit source-derived key to its column and frequency.
type VocabularyTerm struct {
	Key     string `json:"key"`
	Targets int    `json:"targets"`
}

// Vocabulary is an explicit developer artifact containing source-derived text.
// Terms are ordered by column ID; selection ranks training target frequency,
// breaking ties by key. Zero occurrences of known terms remain observed zeros.
type Vocabulary struct {
	Contract        string           `json:"contract"`
	Options         LexicalOptions   `json:"options"`
	TrainingTargets int              `json:"training_targets"`
	Terms           []VocabularyTerm `json:"terms"`
	SHA256          string           `json:"sha256,omitempty"`
}

func (o LexicalOptions) validate() error {
	if err := o.Counts.Validate(); err != nil {
		return err
	}
	if o.MaxFeatures < 1 || o.MaxFeatures > model.MaxFeatures || o.MinTargets < 1 || o.MinTargets > 10000 {
		return fmt.Errorf("invalid lexical vocabulary size or target frequency")
	}
	return nil
}

func lexicalColumnID(key string) string { return fmt.Sprintf("ngram.%x", sha256.Sum256([]byte(key))) }

func lexicalColumns(v Vocabulary, kind string) []feature.Descriptor {
	result := make([]feature.Descriptor, len(v.Terms))
	for i, term := range v.Terms {
		result[i] = feature.Descriptor{ID: lexicalColumnID(term.Key), Version: "1", Family: "lexical-ngram",
			Type: "number", Unit: "log1p-count", Scope: kind, Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences},
			Formula:       "log1p of the frozen vocabulary term count in the prepared target.",
			Normalization: feature.LexicalCountContract,
			Limitations:   "A descriptive lexical measurement; unknown keys are ignored and no editorial quality is implied.",
			Missing:       "Zero is an observed absence of this known term; invalid or incomplete input is an error."}
	}
	return result
}

func freezeVocabulary(ctx context.Context, counts map[string]int, rows int, options LexicalOptions) (Vocabulary, error) {
	terms := make([]VocabularyTerm, 0, len(counts))
	for key, targets := range counts {
		if targets >= options.MinTargets {
			terms = append(terms, VocabularyTerm{Key: key, Targets: targets})
		}
	}
	slices.SortFunc(terms, func(a, b VocabularyTerm) int {
		if a.Targets != b.Targets {
			return b.Targets - a.Targets
		}
		return strings.Compare(a.Key, b.Key)
	})
	terms = slices.Clone(terms[:min(len(terms), options.MaxFeatures)])
	if len(terms) == 0 {
		return Vocabulary{}, fmt.Errorf("training partition has no eligible lexical vocabulary")
	}
	slices.SortFunc(terms, func(a, b VocabularyTerm) int { return strings.Compare(lexicalColumnID(a.Key), lexicalColumnID(b.Key)) })
	v := Vocabulary{Contract: feature.LexicalCountContract, Options: options, TrainingTargets: rows, Terms: terms}
	hash, err := hashJSON(v)
	if err != nil {
		return Vocabulary{}, err
	}
	v.SHA256 = hash
	return v, ctx.Err()
}

func validateVocabulary(v Vocabulary) error {
	if err := v.Options.validate(); err != nil {
		return err
	}
	if v.Contract != feature.LexicalCountContract || v.TrainingTargets < 1 || v.TrainingTargets > 10000 ||
		len(v.Terms) == 0 || len(v.Terms) > v.Options.MaxFeatures {
		return fmt.Errorf("invalid lexical vocabulary contract or dimensions")
	}
	if err := validateVocabularyTerms(v); err != nil {
		return err
	}
	want := v.SHA256
	v.SHA256 = ""
	hash, err := hashJSON(v)
	if err != nil || hash != want {
		return fmt.Errorf("lexical vocabulary digest mismatch")
	}
	return nil
}

func validateVocabularyTerms(v Vocabulary) error {
	last := ""
	for _, term := range v.Terms {
		id := lexicalColumnID(term.Key)
		if len(term.Key) > 1<<16 || !feature.ValidLexicalKey(term.Key, v.Options.Counts) || id <= last ||
			term.Targets < v.Options.MinTargets || term.Targets > v.TrainingTargets {
			return fmt.Errorf("invalid lexical vocabulary key, order, or frequency")
		}
		last = id
	}
	return nil
}
