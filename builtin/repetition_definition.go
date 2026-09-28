package builtin

import (
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// Definition benefits use a closed surface grammar with an exact source operand.
// They do not equate arbitrary benefits or resolve general pronoun references.
type definitionBenefit struct {
	source, form string
	occurrence   rule.Occurrence
	ordinal      int
}

type definitionGroup struct {
	source, form string
	ordinal      int
	mixed        bool
	occurrences  []rule.Occurrence
}

type definitionCollector struct {
	view           rule.View
	budget         *repetitionBudget
	observations   *candidateObservations
	claims         []definitionGroup
	ordinal        int
	pending, bound string
}

func (c *definitionCollector) addBlock(block document.Block, emit rule.Emitter) error {
	eligible, continuous := definitionBlock(block)
	if !eligible {
		c.pending, c.bound = "", ""
		c.ordinal += len(block.Sentences)
		if !continuous {
			return c.finish(emit)
		}
		return nil
	}
	for _, sentence := range block.Sentences {
		if err := c.add(block, sentence, emit); err != nil {
			return err
		}
		c.ordinal++
	}
	return nil
}

func definitionBlock(block document.Block) (bool, bool) {
	if block.Excluded {
		return false, false
	}
	if slices.ContainsFunc(block.Context, func(label string) bool { return strings.HasPrefix(label, "container:") }) {
		return false, false
	}
	if block.Kind == "heading" {
		// Only contribution sections can continue an earlier definition benefit.
		return false, slices.Contains([]string{"add a term", "adding a term"}, strings.ToLower(strings.TrimSpace(block.Text)))
	}
	eligible := block.Kind == "paragraph" && !block.Excluded && block.List == nil && len(block.Sentences) > 0 && leadingCondition(block) == ""
	return eligible, eligible
}

func (c *definitionCollector) add(block document.Block, sentence document.Sentence, emit rule.Emitter) error {
	// The claim collector already charges each block and sentence token.
	// Charge extra token and mapping work only for an eligible construction.
	if !definitionCandidate(sentence.Tokens) {
		c.pending, c.bound = "", ""
		return nil
	}
	if err := c.budget.spend(len(sentence.Tokens)); err != nil {
		return err
	}
	words, err := definitionWords(c.view, block, sentence, c.budget)
	if err != nil {
		return err
	}
	if len(words) == 0 {
		c.pending, c.bound = "", ""
		return nil
	}
	source, form := definitionAssertion(words, c.bound)
	c.updateBinding(words)
	if source == "" || sentence.Words < c.view.Parameters.MinWords {
		return nil
	}
	c.observations.advance(block.ID, candidatePrepared)
	current := definitionBenefit{source, form, tokenOccurrence(sentence, 0, len(restrictionTokens(sentence))), c.ordinal}
	return c.collect(current, emit)
}

func (c *definitionCollector) collect(current definitionBenefit, emit rule.Emitter) error {
	kept := c.claims[:0]
	for _, group := range c.claims {
		if err := c.budget.spend(1); err != nil {
			return err
		}
		if current.ordinal-group.ordinal >= c.view.Parameters.WindowSentences {
			if err := c.emit(group, emit); err != nil {
				return err
			}
		} else {
			kept = append(kept, group)
		}
	}
	c.claims = kept
	for i := range c.claims {
		group := &c.claims[i]
		if group.source == current.source {
			group.mixed = group.mixed || group.form != current.form
			group.occurrences = append(group.occurrences, current.occurrence)
			return nil
		}
	}
	c.claims = append(c.claims, definitionGroup{source: current.source, form: current.form, ordinal: current.ordinal,
		occurrences: []rule.Occurrence{current.occurrence}})
	return nil
}

func (c *definitionCollector) finish(emit rule.Emitter) error {
	for _, group := range c.claims {
		if err := c.budget.spend(1); err != nil {
			return err
		}
		if err := c.emit(group, emit); err != nil {
			return err
		}
	}
	c.claims = nil
	return nil
}

func (c *definitionCollector) emit(group definitionGroup, emit rule.Emitter) error {
	if !group.mixed {
		return nil
	}
	blocked, err := c.excludedBetween(group)
	if err != nil || blocked {
		return err
	}
	for _, occurrence := range group.occurrences {
		c.observations.advance(occurrence.BlockID, candidateEvaluated)
	}
	evidence := measured("heuristic", "definition-benefit-restatements", "patterns", 1, 0, 1, group.occurrences)
	evidence.Suggestion = "State the shared-definition benefit once. Keep the source identity and the separate contribution instructions."
	return emit.Emit(evidence)
}

func (c *definitionCollector) excludedBetween(group definitionGroup) (bool, error) {
	first := document.Bounds(group.occurrences[0].Spans).End
	last := document.Bounds(group.occurrences[len(group.occurrences)-1].Spans).Start
	for _, excluded := range c.view.Document.Excluded {
		if err := c.budget.spend(1); err != nil {
			return false, err
		}
		if slices.Contains([]string{"quote", "code"}, excluded.Reason) && excluded.Span.Start >= first && excluded.Span.End <= last {
			return true, nil
		}
	}
	return false, nil
}

func (c *definitionCollector) updateBinding(words []string) {
	// The immediately preceding add instruction must name the same file. A
	// rendering sentence explicitly calls that file a map; unrelated sentences
	// cannot create or carry this alias, except the intervening linking step.
	pending := c.pending
	c.pending = ""
	if definitionAddition(words) {
		c.pending, c.bound = words[3], ""
		return
	}
	if pending != "" && definitionRendering(words) {
		c.bound = pending
		return
	}
	if strings.Join(words, " ") != "then link here from the pages that use it" {
		c.bound = ""
	}
}

func definitionAddition(words []string) bool {
	// Attribute names stay opaque; unknown qualifications cannot bind a map.
	return definitionPattern(words, "add it to $") || definitionPattern(words,
		"add it to $ with its $ , the $ page that teaches it , and a $ where a file in the repository pins the fact")
}

func definitionRendering(words []string) bool {
	const prefix = "the list above renders from that map , so the term appears here without being transcribed"
	return definitionPattern(words, prefix) || definitionPattern(words,
		prefix+" , and $ refuses an entry that names a page which does not exist")
}

// A dollar sign stands for one already validated opaque operand, not prose.
func definitionPattern(words []string, pattern string) bool {
	parts := strings.Fields(pattern)
	if len(words) != len(parts) {
		return false
	}
	for i, expected := range parts {
		if expected == "$" {
			if !definitionOperand(words[i]) {
				return false
			}
		} else if words[i] != expected {
			return false
		}
	}
	return true
}

func definitionAssertion(words []string, bound string) (string, string) {
	const mapped = "the definition stays in the map and is rendered in one place , which is what keeps " +
		"a word from meaning one thing on one page and something else on the next"
	if bound != "" && definitionPattern(words, mapped) {
		return bound, "map-benefit"
	}
	for _, row := range []struct {
		pattern, form string
		source        int
	}{
		{"definitions come from one source , $ , so the same term cannot come to mean two different things on two pages", "single-source", 6},
		{"definitions in $ give each term the same meaning on every page", "same-meaning", 2},
	} {
		if definitionPattern(words, row.pattern) {
			return words[row.source], row.form
		}
	}
	for _, prefix := range []string{"keeping definitions in $", "keeping the definitions in $"} {
		for _, tail := range []string{
			"prevents a term from meaning one thing on one page and something different on another",
			"prevents different pages from assigning different meanings to the same term",
		} {
			if definitionPattern(words, prefix+" "+tail) {
				return words[len(strings.Fields(prefix))-1], "prevent-different-meanings"
			}
		}
	}
	return "", ""
}

func definitionOperand(word string) bool { return strings.HasPrefix(word, "`") }

// Inspect at most three construction tokens before paying for a mapping walk.
func definitionCandidate(tokens []document.Token) bool {
	if len(tokens) < 3 || len(tokens) > 64 {
		return false
	}
	for _, prefix := range [][]string{
		{"definitions", "in"}, {"definitions", "come"}, {"keeping", "definitions"},
		{"keeping", "the", "definitions"}, {"the", "definition"}, {"the", "list"},
		{"add", "it", "to"}, {"then", "link", "here"},
	} {
		if matches(tokens[:len(prefix)], prefix) {
			return true
		}
	}
	return false
}

func definitionWords(view rule.View, block document.Block, sentence document.Sentence, budget *repetitionBudget) ([]string, error) {
	if !eligibleClaimSentence(sentence) {
		return nil, nil
	}
	tokens := restrictionTokens(sentence)
	if view.Exempts(sentence, 0, len(tokens)) {
		return nil, nil
	}
	ok, err := restrictionMappingEligible(view.Document.Source, block, tokens, budget)
	if err != nil || !ok {
		return nil, err
	}
	words := make([]string, len(tokens))
	for i, token := range tokens {
		words[i] = token.Normal
		if token.Protected {
			words[i], err = definitionSource(view.Document.Source, token, budget)
			if err != nil || words[i] == "" {
				return nil, err
			}
		}
	}
	return words, nil
}

func definitionSource(source []byte, token document.Token, budget *repetitionBudget) (string, error) {
	if len(token.Spans) != 1 || !token.Spans[0].Valid(len(source)) {
		return "", nil
	}
	span := token.Spans[0]
	if err := budget.spend(span.End - span.Start); err != nil {
		return "", err
	}
	if span.End-span.Start > 128 {
		return "", nil
	}
	raw := string(source[span.Start:span.End])
	if !definitionSourceText(raw) {
		return "", nil
	}
	return raw, nil
}

func definitionSourceText(raw string) bool {
	if len(raw) < 3 || raw[0] != '`' || raw[len(raw)-1] != '`' {
		return false
	}
	return !strings.ContainsFunc(raw[1:len(raw)-1], func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("._-/", r)
	})
}
