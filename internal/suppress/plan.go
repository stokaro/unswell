package suppress

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
)

// Target identifies a complete structural unit permitted by a directive.
type Target struct {
	Scope string
	ID    int
	Span  document.Span
	prose document.Span
}

// Entry retains the directive, its structural targets, and matched raw findings.
type Entry struct {
	Kind, Reason string
	Rules        []string
	Span         document.Span
	End          *document.Span
	Targets      []Target
	Findings     []string
	UsedRules    []string
}

// Plan belongs to one analysis call; matching updates its audit records and budget.
type Plan struct {
	Entries     []Entry
	byRule      map[string][]int
	budget      int
	options     Options
	targetCount int
	blocks      []resolvedBlock
}

// Build validates source directives and binds them after sentence segmentation.
func Build(ctx context.Context, doc document.Document, catalog map[string]bool, options Options) (*Plan, error) {
	p := &Plan{byRule: make(map[string][]int), budget: options.MaxCandidates, options: options}
	if err := ctx.Err(); err != nil {
		return p, err
	}
	if len(doc.Directives) > 1000 {
		return p, fmt.Errorf("source exceeds 1000 suppression directives")
	}
	if len(doc.Directives) == 0 {
		return p, nil
	}
	if err := p.resolveBlocks(ctx, doc.Blocks); err != nil {
		return p, err
	}
	if err := p.bindDirectives(ctx, doc, catalog, options); err != nil {
		return p, err
	}
	slices.SortFunc(p.Entries, func(a, b Entry) int { return a.Span.Start - b.Span.Start })
	return p, p.index(ctx)
}

func (p *Plan) bindDirectives(ctx context.Context, doc document.Document, catalog map[string]bool, options Options) error {
	raw := slices.Clone(doc.Directives)
	slices.SortFunc(raw, func(a, b document.Directive) int { return a.Span.Start - b.Span.Start })
	var stack []directive
	previous := 0
	for _, value := range raw {
		if err := p.work(ctx); err != nil {
			return err
		}
		if !value.Span.Valid(len(doc.Source)) || value.Span.Start < previous {
			return fmt.Errorf("invalid or overlapping directive range")
		}
		previous = value.Span.End
		d, err := parse(value, catalog, options)
		if err != nil {
			return at(value.Span, err)
		}
		stack, err = p.bind(ctx, doc, d, stack)
		if err != nil {
			return at(value.Span, err)
		}
	}
	if len(stack) != 0 {
		return at(stack[len(stack)-1].span, fmt.Errorf("suppression region has no matching enable"))
	}
	return nil
}

func at(span document.Span, err error) error {
	return fmt.Errorf("suppression at bytes [%d,%d): %w", span.Start, span.End, err)
}

func (p *Plan) bind(ctx context.Context, doc document.Document, d directive, stack []directive) ([]directive, error) {
	if d.kind == "disable" {
		return append(stack, d), nil
	}
	var end *document.Span
	if d.kind == "enable" {
		if len(stack) == 0 || !slices.Equal(stack[len(stack)-1].ids, d.ids) {
			return stack, fmt.Errorf("enable must match the innermost region's rule IDs")
		}
		closing := d.span
		end = &closing
		d = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
	}
	targets, err := p.targets(ctx, doc, d, end)
	if err != nil {
		return stack, err
	}
	p.targetCount += len(targets)
	if p.targetCount > 10000 {
		return stack, fmt.Errorf("suppression plan exceeds 10000 structural targets")
	}
	p.Entries = append(p.Entries, Entry{Kind: d.kind, Reason: d.reason, Rules: d.ids, Span: d.span, End: end, Targets: targets})
	return stack, nil
}

func (p *Plan) work(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p.budget--
	if p.budget < 0 {
		return fmt.Errorf("suppression resolution exceeds max_candidates")
	}
	return nil
}

func (p *Plan) targets(ctx context.Context, doc document.Document, d directive, end *document.Span) ([]Target, error) {
	var result []Target
	for _, block := range p.blocks {
		if err := p.work(ctx); err != nil {
			return nil, err
		}
		if !eligibleBlock(block, doc.Format, d.kind) {
			continue
		}
		selected, err := p.blockTargets(ctx, block, d, end)
		if err != nil {
			return nil, err
		}
		result = append(result, selected...)
		if len(result) > 10000 {
			return nil, fmt.Errorf("suppression exceeds 10000 structural targets")
		}
		if len(result) != 0 && nextUnit(d.kind) {
			return result, nil
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("suppression has no complete eligible target")
	}
	return result, nil
}

func eligibleBlock(block resolvedBlock, format document.Format, kind string) bool {
	return block.words > 0 && (format != document.Go || block.kind == "comment" || kind == "disable-file")
}

func nextUnit(kind string) bool {
	return kind == "disable-next-sentence" || kind == "disable-next-block"
}

func (p *Plan) blockTargets(ctx context.Context, block resolvedBlock, d directive, end *document.Span) ([]Target, error) {
	var result []Target
	for _, unit := range targetUnits(block, d.kind, d.span.End, end) {
		if err := p.work(ctx); err != nil {
			return nil, err
		}
		if d.kind != "disable-file" && (unit.prose.Start < d.span.End || end != nil && unit.prose.End > end.Start) {
			continue
		}
		result = append(result, unit)
		if nextUnit(d.kind) {
			break
		}
	}
	return result, nil
}

func targetUnits(block resolvedBlock, kind string, start int, end *document.Span) []Target {
	wholeBlock := block.paragraph.prose.Start >= start && (end == nil || block.paragraph.prose.End <= end.Start)
	if kind != "disable-next-sentence" && (kind != "disable" || wholeBlock) {
		return []Target{block.paragraph}
	}
	return block.sentences
}
