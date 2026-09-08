package builtin

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func observeBlock(view rule.View, block document.Block, reason string) error {
	observation := feature.BlockObservation{BlockID: block.ID, Status: "evaluated"}
	if reason != "" {
		observation.Status, observation.Reason = "inapplicable", reason
	}
	return view.Observe(observation)
}

func observedProseWord(view rule.View, sentence document.Sentence, index int) bool {
	return view.Observer != nil && surfaceProseWord(sentence.Tokens[index]) && !view.Exempts(sentence, index, index+1)
}

func tokenBlockReason(block document.Block, evaluated, hasPatterns bool) string {
	switch {
	case !hasPatterns:
		return "no_patterns"
	case len(block.Sentences) == 0:
		return "no_sentences"
	case !evaluated:
		return "no_eligible_tokens"
	default:
		return ""
	}
}

func proseMinimumReason(block document.Block, minimum int) string {
	switch {
	case !proseBlock(block):
		return "unsupported_unit"
	case block.Words == 0:
		return "no_prose_words"
	case block.Words < minimum:
		return "insufficient_words"
	default:
		return ""
	}
}
