// Package nlp defines the backend contract without exposing vendor types.
package nlp

import (
	"context"

	"github.com/stokaro/unswell/document"
)

// Capability is a separately implemented linguistic representation.
type Capability string

// Capability names distinguish surface analysis from dependency parsing.
const (
	Tokens       Capability = "tokens"
	Sentences    Capability = "sentences"
	POS          Capability = "pos"
	Chunks       Capability = "chunks"
	Lemmas       Capability = "lemmas"
	Dependencies Capability = "dependencies"
	Entities     Capability = "entities"
)

// Identity records a pinned provider and its supported feature contract.
type Identity struct {
	Name              string       `json:"name"`
	Version           string       `json:"version"`
	Model             string       `json:"model"`
	ModelHash         string       `json:"model_sha256"`
	SentenceModelHash string       `json:"sentence_model_sha256"`
	License           string       `json:"license"`
	Capabilities      []Capability `json:"capabilities"`
}

// Provider must support concurrent calls and must not mutate its input.
type Provider interface {
	Identity() Identity
	Analyze(context.Context, document.MappedText, []Capability) ([]document.Sentence, error)
}
