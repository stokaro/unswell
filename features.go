package unswell

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
)

func (e *Engine) sharedFeatures(ctx context.Context, doc *document.Document) (*feature.Set, error) {
	if !e.needsSharedFeatures() {
		return nil, nil
	}
	set, err := feature.NewSet(ctx, doc.Blocks, e.featureIdentity(doc), feature.Limits{
		MaxTokens: e.policy.Analysis.MaxCandidates, MaxUniqueWords: e.policy.Analysis.MaxCandidates,
		MaxBytes: e.policy.Analysis.MaxFileBytes, MaxBlocks: e.policy.Analysis.MaxBlocks,
	})
	if errors.Is(err, feature.ErrTokenLimit) {
		return nil, fmt.Errorf("shared feature checks exceed max_candidates: %w", err)
	}
	return set, err
}

func (e *Engine) needsSharedFeatures() bool {
	if len(e.featureIDs) > 0 {
		return true
	}
	for _, descriptor := range e.descriptors {
		if e.policy.Rules[descriptor.ID].Enabled && descriptor.SharedFeatures {
			return true
		}
	}
	return false
}

func (e *Engine) featureIdentity(doc *document.Document) feature.Identity {
	identity := e.nlp.Identity()
	identity.Capabilities = slices.Clone(identity.Capabilities)
	return feature.Identity{NLP: identity, Capabilities: slices.Clone(e.capabilities), Source: doc.Hash,
		Policy: e.policy.Hash, Vocabulary: e.policy.Hash, Preprocessing: "extracted-provider-tokens-v1"}
}
