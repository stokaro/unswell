package builtin

import (
	"context"
	"strconv"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func numberedFraming(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	observations := newCandidateObservations(view, numberedProseBlock)
	gaps := newProseGapIndex(view.Document.Excluded)
	var previous *document.Block
	for i := range view.Document.Blocks {
		if err := m.spend(); err != nil {
			return err
		}
		block := &view.Document.Blocks[i]
		blocked := gaps.between(previous, block)
		if !blocked && adjacentNumberedBlocks(view.Document, previous, block) {
			observations.advance(previous.ID, candidateEvaluated)
			observations.advance(block.ID, candidateEvaluated)
			if err := emitNumberedFraming(view, *previous, *block, emit); err != nil {
				return err
			}
		}
		previous = block
	}
	return observations.finish(ctx, view)
}

func numberedProseBlock(block document.Block) bool {
	return !block.Excluded && headingProseBlock(block)
}

func adjacentNumberedBlocks(doc *document.Document, before, after *document.Block) bool {
	return before != nil && !before.Excluded && !after.Excluded &&
		before.Kind == "paragraph" && after.Kind == "heading" && adjacentProse(doc, before, after, true)
}

func emitNumberedFraming(view rule.View, paragraph, heading document.Block, emit rule.Emitter) error {
	frame, ok := findNumberedFraming(paragraph, heading)
	if !ok || frameExempt(view, frame) || view.Parameters.AllowedOccurrences > 0 {
		return nil
	}
	evidence := measured("heuristic", "counted-introduction-heading", "patterns", 1,
		view.Parameters.AllowedOccurrences, view.Parameters.SaturationOccurrences,
		[]rule.Occurrence{frameOccurrence(frame.parts[0]), frameOccurrence(frame.parts[1])})
	evidence.Suggestion = "Name the configuration or task in the introduction and heading. Keep a count when it states a real requirement; " +
		"this rule does not check the count against code or prove that it is wrong."
	return emit.Emit(evidence)
}

func findNumberedFraming(paragraph, heading document.Block) (rhetoricalFrame, bool) {
	left, leftOK := numberedOpening(paragraph, 96, 0)
	right, rightOK := numberedOpening(heading, 4, 1)
	if !leftOK || !rightOK || len(heading.Sentences) != 1 || !left.eligible() {
		return rhetoricalFrame{}, false
	}
	count, noun, rest := countedFramePrefix(right.tokens())
	if count == 0 || len(rest) != 0 {
		return rhetoricalFrame{}, false
	}
	if !matchingCountedLead(left, count, noun) {
		return rhetoricalFrame{}, false
	}
	return rhetoricalFrame{parts: []frameClause{left, right}}, true
}

func matchingCountedLead(left frameClause, count int, noun string) bool {
	other, subject, tail := countedFramePrefix(left.tokens())
	return other == count && subject == noun && len(tail) >= 2 && !fixedCountClaim(left.sentence.Tokens)
}

func numberedOpening(block document.Block, limit, ordinal int) (frameClause, bool) {
	if len(block.Sentences) == 0 || len(block.Sentences[0].Tokens) > limit || question(block.Sentences[0]) {
		return frameClause{}, false
	}
	parts := frameClauses(block.Sentences[0], ordinal)
	if len(parts) == 0 || rhetoricQuoted(parts[0]) || rhetoricAttributed(parts[0]) {
		return frameClause{}, false
	}
	return parts[0], true
}

// A bare count heading must repeat the lead's generic content noun. A title
// that names its task, and a count of concrete technical objects, do not qualify.
func countedFramePrefix(tokens []document.Token) (int, string, []document.Token) {
	if len(tokens) > 0 && frameWord(tokens[0], "the", "these", "just", "only") {
		tokens = tokens[1:]
	}
	if len(tokens) < 2 || tokens[0].Protected || !frameWord(tokens[1], "lines", "steps", "rules", "points", "things", "ways", "tips") {
		return 0, "", nil
	}
	n := framingCount(tokens[0].Normal)
	if n < 2 || n > 99 {
		return 0, "", nil
	}
	return n, tokens[1].Normal, tokens[2:]
}

func framingCount(text string) int {
	for i, word := range []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine",
		"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen", "twenty"} {
		if text == word {
			return i
		}
	}
	n, err := strconv.Atoi(text)
	if err != nil || strconv.Itoa(n) != text {
		return 0
	}
	return n
}

func fixedCountClaim(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "must", "shall", "required", "requires", "require", "exactly", "fixed", "mandatory", "not", "never", "n't") {
			return true
		}
	}
	return false
}
