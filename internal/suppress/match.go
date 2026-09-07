package suppress

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
)

func (p *Plan) index(ctx context.Context) error {
	for index, entry := range p.Entries {
		for _, id := range entry.Rules {
			for _, previous := range p.byRule[id] {
				if err := p.disjoint(ctx, entry, p.Entries[previous], id); err != nil {
					return err
				}
			}
			p.byRule[id] = append(p.byRule[id], index)
		}
	}
	return nil
}

func (p *Plan) disjoint(ctx context.Context, left, right Entry, id string) error {
	for _, a := range left.Targets {
		for _, b := range right.Targets {
			if err := p.work(ctx); err != nil {
				return err
			}
			if a.Span.Start < b.Span.End && b.Span.Start < a.Span.End {
				return at(left.Span, fmt.Errorf("overlapping suppression targets for rule %s", id))
			}
		}
	}
	return nil
}

// Match permits a finding only when every evidence segment is covered. It returns
// the contributing entry indexes and records use only after full coverage succeeds.
func (p *Plan) Match(ctx context.Context, ruleID, findingID string, spans []document.Span) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(spans) == 0 || len(p.byRule[ruleID]) == 0 {
		return nil, nil
	}
	var matched []int
	for _, span := range spans {
		index, err := p.covering(ctx, ruleID, span)
		if err != nil || index < 0 {
			return nil, err
		}
		if !slices.Contains(matched, index) {
			matched = append(matched, index)
		}
	}
	slices.Sort(matched)
	p.markUsed(matched, ruleID, findingID)
	return matched, nil
}

func (p *Plan) markUsed(matched []int, ruleID, findingID string) {
	for _, index := range matched {
		entry := &p.Entries[index]
		if !slices.Contains(entry.Findings, findingID) {
			entry.Findings = append(entry.Findings, findingID)
		}
		if !slices.Contains(entry.UsedRules, ruleID) {
			entry.UsedRules = append(entry.UsedRules, ruleID)
			slices.Sort(entry.UsedRules)
		}
	}
}

func (p *Plan) covering(ctx context.Context, ruleID string, span document.Span) (int, error) {
	if span.Start < 0 || span.End <= span.Start {
		return -1, fmt.Errorf("invalid suppression evidence range")
	}
	for _, index := range p.byRule[ruleID] {
		for _, target := range p.Entries[index].Targets {
			if err := p.work(ctx); err != nil {
				return -1, err
			}
			if span.Start >= target.Span.Start && span.End <= target.Span.End {
				return index, nil
			}
		}
	}
	return -1, nil
}

// ValidateUse rejects each permission that did not cover a complete raw finding.
func (p *Plan) ValidateUse() error {
	if !p.options.RejectUnused {
		return nil
	}
	for _, entry := range p.Entries {
		for _, id := range entry.Rules {
			if !slices.Contains(entry.UsedRules, id) {
				return at(entry.Span, fmt.Errorf("unused suppression for rule %s", id))
			}
		}
	}
	return nil
}
