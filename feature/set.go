package feature

import (
	"context"
	"fmt"

	"github.com/stokaro/unswell/document"
)

// Set holds one immutable measurement per supported block. It retains no input
// document buffers and permits concurrent reads. It is local to an analysis run;
// persisted caches must also compare each unit's Hash and feature contract.
type Set struct {
	blocks map[int]Measurements
}

// NewSet computes supported blocks once within a total token and byte budget.
// Inputs must remain unchanged until it returns. Independent blocks are never
// joined to meet a minimum evidence threshold.
func NewSet(ctx context.Context, blocks []document.Block, identity Identity, limits Limits) (*Set, error) {
	if err := validateLimits(limits); err != nil {
		return nil, err
	}
	if err := validateIdentity(identity); err != nil {
		return nil, err
	}
	if len(blocks) > limits.MaxBlocks {
		return nil, fmt.Errorf("features exceed max_blocks")
	}
	return measureBlocks(ctx, blocks, identity, limits)
}

func measureBlocks(ctx context.Context, blocks []document.Block, identity Identity, limits Limits) (*Set, error) {
	set := &Set{blocks: make(map[int]Measurements)}
	seen := make(map[int]bool)
	visits, bytes := 0, 0
	for _, block := range blocks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if block.ID < 0 || seen[block.ID] {
			return nil, fmt.Errorf("invalid or duplicate feature block ID %d", block.ID)
		}
		seen[block.ID] = true
		if !SupportsBlock(block.Kind) {
			continue
		}
		visits += tokenCount(block)
		bytes += len(block.Text)
		if visits > limits.MaxTokens {
			return nil, ErrTokenLimit
		}
		if bytes > limits.MaxBytes {
			return nil, fmt.Errorf("feature set exceeds total byte budget")
		}
		m, err := Measure(ctx, block, identity, limits)
		if err != nil {
			return nil, err
		}
		set.blocks[block.ID] = m
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return set, nil
}

// Block returns an immutable measurement for a supported block ID.
// An absent set or unknown block is an error, never an empty successful vector.
func (s *Set) Block(id int) (Measurements, error) {
	if s != nil {
		if measurements, ok := s.blocks[id]; ok {
			return measurements, nil
		}
	}
	return Measurements{}, fmt.Errorf("feature block %d was not measured", id)
}

func tokenCount(block document.Block) int {
	total := 0
	for _, sentence := range block.Sentences {
		total += len(sentence.Tokens)
	}
	return total
}
