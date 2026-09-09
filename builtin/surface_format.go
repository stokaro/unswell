package builtin

import (
	"context"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func emDashDensity(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	for _, block := range view.Document.Blocks {
		if !proseBlock(block) || block.Words == 0 || block.Words < view.Parameters.MinWords {
			if err := observeBlock(view, block, proseMinimumReason(block, view.Parameters.MinWords)); err != nil {
				return err
			}
			continue
		}
		if err := emitDashDensity(m, block, emit); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func emitDashDensity(m *editorialMatcher, block document.Block, emit rule.Emitter) error {
	count, occurrences, err := emDashes(m, block)
	if err != nil {
		return err
	}
	if err := observeBlock(m.view, block, ""); err != nil {
		return err
	}
	p := m.view.Parameters
	value := float64(count) * 100 / float64(block.Words)
	if count <= p.AllowedOccurrences || value <= float64(p.Onset) {
		return nil
	}
	return emit.Emit(rule.Evidence{Kind: "heuristic", Activation: metricActivation(value, p.Onset, p.Saturation),
		Occurrences: occurrences, Metrics: []rule.Metric{
			{Name: "em-dash-density", Value: value, Unit: "em-dashes/100-prose-words",
				Onset: float64(p.Onset), Saturation: float64(p.Saturation)},
			{Name: "em-dash-count", Value: float64(count), Unit: "em-dashes"},
			{Name: "prose-denominator", Value: float64(block.Words), Unit: "words"},
		}})
}

func emDashes(m *editorialMatcher, block document.Block) (int, []rule.Occurrence, error) {
	var occurrences []rule.Occurrence
	mapping := newMappedRangeIndex(block)
	count := 0
	for at, r := range block.Text {
		if err := m.spend(); err != nil {
			return 0, nil, err
		}
		if r == '—' {
			count++
			occurrences = append(occurrences, mapping.occurrences(at, at+len("—"))...)
		}
	}
	return count, occurrences, nil
}

type shortList struct {
	context              document.ListContext
	section, first, last int
	blocks               []document.Block
	eligible             bool
}

func listFragmentation(ctx context.Context, view rule.View, emit rule.Emitter) error {
	observations := newListObservations(view)
	m := newEditorialMatcher(ctx, view)
	groups, err := shortLists(m)
	if err != nil {
		return err
	}
	groups = slices.DeleteFunc(groups, func(group shortList) bool {
		return !group.eligible || len(group.blocks) != group.context.Items
	})
	if observations != nil {
		for _, group := range groups {
			observations.listGroup(group, candidatePrepared)
		}
	}
	if err := emitShortLists(m, groups, emit, observations); err != nil {
		return err
	}
	return observations.finish(ctx, view)
}

func shortLists(m *editorialMatcher) ([]shortList, error) {
	var groups []shortList
	ids := make(map[int]int)
	section := 0
	var reference []string
	gaps := newProseGapIndex(m.view.Document.Excluded)
	var previous *document.Block
	for at := range m.view.Document.Blocks {
		block := &m.view.Document.Blocks[at]
		if err := m.spend(); err != nil {
			return nil, err
		}
		if listSectionBoundary(&gaps, previous, block) {
			section++
		}
		previous = block
		reference = listReferenceHeading(*block, reference)
		if !grammarListBlock(block) {
			continue
		}
		i, exists := ids[block.List.ID]
		if !exists {
			i = len(groups)
			ids[block.List.ID] = i
			groups = append(groups, shortList{context: *block.List, section: section, first: at,
				eligible: eligibleList(*block.List, m.view.Parameters) && !listReferenceScope(block.Context, reference)})
		}
		ok, err := shortFragment(m, *block)
		if err != nil {
			return nil, err
		}
		group := &groups[i]
		group.eligible = group.eligible && ok && group.section == section
		group.blocks = append(group.blocks, *block)
		group.last = at
	}
	return groups, nil
}

func grammarListBlock(block *document.Block) bool {
	return block.Kind == "list-item" && block.List != nil
}

func listSectionBoundary(gaps *proseGapIndex, previous, block *document.Block) bool {
	return gaps.between(previous, block) || (block.Kind != "paragraph" && block.Kind != "list-item")
}

func eligibleList(list document.ListContext, p rule.Parameters) bool {
	return list.Depth == 1 && !list.Ordered && !list.Task && !list.Complex && list.Items > 0 && list.Items <= p.MaxListItems
}

func listReferenceHeading(block document.Block, reference []string) []string {
	if block.Kind == "heading" {
		return nextSummaryScope(block, reference, referenceHeadings())
	}
	return reference
}

func referenceHeadings() []string {
	return []string{"reference", "api reference", "usage", "installation", "procedure", "procedures", "steps", "requirements",
		"configuration", "options", "parameters"}
}

func listReferenceScope(context, reference []string) bool {
	if len(reference) == 0 {
		return false
	}
	for _, label := range reference {
		if strings.HasPrefix(label, "heading-") && !slices.Contains(context, label) {
			return false
		}
	}
	return true
}

func shortFragment(m *editorialMatcher, block document.Block) (bool, error) {
	if block.Words == 0 || block.Words > m.view.Parameters.MaxItemWords || strings.ContainsAny(block.Text, ":?!.") {
		return false, nil
	}
	for _, sentence := range block.Sentences {
		if procedureSentence(sentence) {
			return false, nil
		}
		for i, token := range sentence.Tokens {
			if err := m.spend(); err != nil {
				return false, err
			}
			if protectedFragmentToken(token, i) || m.view.Exempts(sentence, i, i+1) {
				return false, nil
			}
		}
	}
	return true, nil
}

func procedureSentence(sentence document.Sentence) bool {
	return len(sentence.Tokens) > 0 && (sentence.Tokens[0].Tag == "VB" || proceduralOpening(sentence.Tokens[0].Normal))
}

func protectedFragmentToken(token document.Token, index int) bool {
	return token.Protected || protectedIdentifier(token, index) || token.Tag == "CD" || (token.Word && !surfaceProseWord(token))
}

func proceduralOpening(word string) bool {
	return slices.Contains([]string{"add", "apply", "build", "check", "choose", "click", "close", "configure", "create", "delete",
		"edit", "enable", "ensure", "enter", "install", "keep", "open", "press", "read", "remove", "replace", "restart", "return",
		"run", "save", "select", "set", "start", "stop", "type", "use", "verify", "wait"}, word)
}

func emitShortLists(m *editorialMatcher, groups []shortList, emit rule.Emitter, observations *candidateObservations) error {
	p := m.view.Parameters
	for start := 0; start < len(groups); {
		if err := m.spend(); err != nil {
			return err
		}
		end := start
		for end < len(groups) && groups[end].section == groups[start].section && groups[end].last-groups[start].first < p.WindowBlocks {
			observations.listGroup(groups[end], candidateEvaluated)
			end++
		}
		if end-start <= p.AllowedOccurrences {
			start++
			continue
		}
		var occurrences []rule.Occurrence
		items, words := 0, 0
		for _, group := range groups[start:end] {
			for _, block := range group.blocks {
				items++
				words += block.Words
				occurrences = append(occurrences, blockOccurrences(block)...)
			}
		}
		evidence := measured("heuristic", "short-unordered-lists", "lists", end-start, p.AllowedOccurrences, p.SaturationOccurrences, occurrences)
		evidence.Metrics = append(evidence.Metrics, rule.Metric{Name: "short-list-items", Value: float64(items), Unit: "items"},
			rule.Metric{Name: "list-prose-words", Value: float64(words), Unit: "words"})
		if err := emit.Emit(evidence); err != nil {
			return err
		}
		start = end
	}
	return m.ctx.Err()
}
