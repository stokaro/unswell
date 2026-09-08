package unswell

import (
	"context"
	"errors"
	"fmt"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
)

func (e *Engine) sharedFeatures(ctx context.Context, doc *document.Document) (*feature.Set, error) {
	for _, descriptor := range e.descriptors {
		if e.policy.Rules[descriptor.ID].Enabled && descriptor.SharedFeatures {
			set, err := feature.NewSet(ctx, doc.Blocks, feature.Identity{
				NLP: e.nlp.Identity(), Capabilities: e.capabilities, Source: doc.Hash,
				Policy: e.policy.Hash, Vocabulary: e.policy.Hash, Preprocessing: "extracted-provider-tokens-v1",
			}, feature.Limits{
				MaxTokens: e.policy.Analysis.MaxCandidates, MaxUniqueWords: e.policy.Analysis.MaxCandidates,
				MaxBytes: e.policy.Analysis.MaxFileBytes, MaxBlocks: e.policy.Analysis.MaxBlocks,
			})
			if errors.Is(err, feature.ErrTokenLimit) {
				return nil, fmt.Errorf("shared feature checks exceed max_candidates: %w", err)
			}
			return set, err
		}
	}
	return nil, nil
}
