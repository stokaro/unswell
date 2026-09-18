package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// Contextual constructions use only bounded, affirmative source clauses. A
// stated condition, measurement or quotation keeps their interpretation open.
func contextualRhetoricScoped(c frameClause) bool {
	if len(c.sentence.Tokens) > 96 || quotedClaim(c.sentence.Tokens) {
		return true
	}
	for i, token := range c.sentence.Tokens {
		if (token.Tag == "CD" && !indefiniteOne(c.sentence.Tokens, i)) ||
			frameWord(token, "if", "when", "unless", "because", "whenever", "except",
				"without", "according", "measured", "benchmark", "benchmarks", "median", "percentile",
				"claims", "claimed", "says", "said", "argues", "argued", "rejects", "refutes", "disputes") {
			return true
		}
	}
	return false
}

// The tagger labels the pronoun in "is one that ..." as a number. This
// construction identifies an antecedent rather than a measured quantity.
func indefiniteOne(tokens []document.Token, i int) bool {
	return i > 0 && i+1 < len(tokens) && frameWord(tokens[i], "one") &&
		frameWord(tokens[i-1], "is", "was", "are", "were") &&
		frameWord(tokens[i+1], "that", "which", "who", "nobody", "noone", "everyone", "everybody")
}

func contextualEvaluation(tokens []document.Token) bool {
	for i := 1; i < min(len(tokens)-2, 12); i++ {
		if !frameWord(tokens[i], "is", "are", "was", "were") || !positiveRhetoricSubject(tokens[:i]) {
			continue
		}
		rest := tokens[i+1:]
		return cognitiveWorth(rest) || (deicticEvaluationSubject(tokens[:i]) && wholeAbstractValue(rest))
	}
	return false
}

func cognitiveWorth(tokens []document.Token) bool {
	if len(tokens) < 2 || !frameWord(tokens[0], "worth") ||
		!frameWord(tokens[1], "knowing", "noting", "remembering", "understanding", "mentioning") {
		return false
	}
	return len(tokens) == 2 || frameWord(tokens[2], "before", "in", "ahead", "now", "here")
}

func deicticEvaluationSubject(tokens []document.Token) bool {
	if len(tokens) == 1 {
		return frameWord(tokens[0], "this", "that", "it", "which")
	}
	return len(tokens) == 2 && frameWord(tokens[0], "this", "that", "the") &&
		frameWord(tokens[1], "distinction", "difference", "separation", "contrast", "boundary", "behavior", "choice")
}

func wholeAbstractValue(tokens []document.Token) bool {
	if len(tokens) < 3 || !frameWord(tokens[0], "the") || !frameWord(tokens[1], "whole", "entire") ||
		!frameWord(tokens[2], "value", "point", "purpose", "benefit") {
		return false
	}
	if len(tokens) == 3 {
		return true
	}
	return abstractValueObject(tokens[3:])
}

func abstractValueObject(tokens []document.Token) bool {
	if !frameWord(tokens[0], "of") {
		return false
	}
	return cognitiveReading(tokens[1:]) || len(tokens) == 3 && frameWord(tokens[1], "the", "this", "that") &&
		frameWord(tokens[2], "verb", "approach", "choice", "feature", "option", "design", "decision", "distinction", "difference")
}

func contextualDocumentFrame(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if contextualRhetoricScoped(c) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := range tokens {
		if !embeddedClauseStart(tokens, i) {
			continue
		}
		if contextualDocumentPredicate(tokens[i:]) {
			return localFrame(c, i)
		}
	}
	if frameWord(tokens[0], "it", "they") && editorialOwnership(tokens[1:]) && nearbyDocumentAntecedent(clauses, index) {
		return localFrame(c, 0)
	}
	return rhetoricalFrame{}, false
}

func contextualDocumentPredicate(tokens []document.Token) bool {
	subject := documentSubjectEnd(tokens)
	return documentPresentation(tokens) || subject > 0 && (documentMeaning(tokens[subject:]) || editorialOwnership(tokens[subject:]))
}

func documentPresentation(tokens []document.Token) bool {
	start := presentationSubjectEnd(tokens)
	if start == 0 || start+2 >= len(tokens) {
		return false
	}
	rest := tokens[start:]
	return documentDedication(rest) ||
		frameWord(rest[0], "looks", "look") && frameWord(rest[1], "at") ||
		frameWord(rest[0], "will") && frameWord(rest[1], "detail", "describe", "explain")
}

func presentationSubjectEnd(tokens []document.Token) int {
	if len(tokens) >= 7 && frameWord(tokens[0], "the") && frameWord(tokens[1], "rest", "remainder") &&
		frameWord(tokens[2], "of") {
		if end := documentSubjectEnd(tokens[3:]); end > 0 {
			return 3 + end
		}
	}
	return documentSubjectEnd(tokens)
}

func documentDedication(tokens []document.Token) bool {
	return frameWord(tokens[0], "is", "was") && frameWord(tokens[1], "dedicated", "devoted") && frameWord(tokens[2], "to")
}

func nearbyDocumentAntecedent(clauses []frameClause, index int) bool {
	if index == 0 {
		return false
	}
	previous, current := clauses[index-1], clauses[index]
	return previous.eligible() && previous.sentence.BlockID == current.sentence.BlockID &&
		previous.ordinal+1 == current.ordinal && documentAnnouncement(previous.tokens()) &&
		!contextualRhetoricScoped(previous) && !interveningActor(previous.tokens())
}

func interveningActor(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, ",", "but", "while", "whereas") {
			return true
		}
	}
	return false
}

func documentMeaning(tokens []document.Token) bool {
	if len(tokens) < 4 || !frameWord(tokens[0], "is", "was") || !frameWord(tokens[1], "what") {
		return false
	}
	for _, token := range tokens[2:min(len(tokens), 10)] {
		if frameWord(token, "means", "mean", "explains", "explained", "describes", "covers") {
			return true
		}
	}
	return false
}

func editorialOwnership(tokens []document.Token) bool {
	if len(tokens) < 3 || !frameWord(tokens[0], "owns", "own", "owned") || !frameWord(tokens[1], "the", "this", "that") {
		return false
	}
	return frameWord(tokens[2], "sequence", "detail", "details", "explanation", "instructions", "walkthrough") &&
		(len(tokens) == 3 || frameWord(tokens[3], ","))
}

func contextualAssuranceFrame(c frameClause) (rhetoricalFrame, bool) {
	if contextualRhetoricScoped(c) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := range tokens {
		if embeddedClauseStart(tokens, i) && extremeQualityPredicate(tokens[i:]) {
			return localFrame(c, i)
		}
		if universalReaderStart(tokens, i) && universalReaderClaim(tokens[i:]) {
			return localFrame(c, i)
		}
	}
	return rhetoricalFrame{}, false
}

func extremeQualityPredicate(tokens []document.Token) bool {
	for i := 1; i < min(len(tokens)-2, 12); i++ {
		if !positiveRhetoricSubject(tokens[:i]) {
			continue
		}
		if frameWord(tokens[i], "is", "are", "was", "were") && intensifiedQuality(tokens[i+1:]) {
			return true
		}
		if frameWord(tokens[i], "offers", "offer", "provides", "provide", "exposes", "expose", "delivers", "deliver",
			"brings", "bring", "gives", "give") && superlativeQuality(tokens[i+1:]) {
			return true
		}
	}
	return false
}

func intensifiedQuality(tokens []document.Token) bool {
	if len(tokens) > 2 && frameWord(tokens[0], "also") {
		tokens = tokens[1:]
	}
	return len(tokens) >= 2 && frameWord(tokens[0], "ridiculously", "unbelievably", "incredibly", "amazingly") &&
		frameWord(tokens[1], "extensible", "powerful", "flexible", "easy", "fast", "reliable", "robust", "effective", "efficient")
}

func superlativeQuality(tokens []document.Token) bool {
	if len(tokens) > 2 && frameWord(tokens[0], "a", "an", "the") {
		tokens = tokens[1:]
	}
	if len(tokens) < 2 || !frameWord(tokens[0], "unprecedented", "unparalleled", "unmatched", "unrivaled", "ultimate") {
		return false
	}
	i := 1
	if len(tokens) > 3 && frameWord(tokens[i], "level", "degree", "amount") && frameWord(tokens[i+1], "of") {
		i += 2
	}
	return frameWord(tokens[i], "control", "flexibility", "power", "performance", "safety", "reliability", "efficiency",
		"productivity", "scalability", "precision", "speed")
}

func universalReaderStart(tokens []document.Token, i int) bool {
	return embeddedClauseStart(tokens, i) || (i > 0 && !tokens[i-1].Protected &&
		(tokens[i-1].Tag == "NN" || tokens[i-1].Tag == "NNS" || indefiniteOne(tokens, i-1) ||
			(i > 1 && frameWord(tokens[i-1], "that", "which", "who") && indefiniteOne(tokens, i-2))))
}

func universalReaderClaim(tokens []document.Token) bool {
	if len(tokens) < 3 {
		return false
	}
	i := readerPredicateStart(tokens)
	if i == 0 || i >= len(tokens) || !frameWord(tokens[i], "know", "knows", "known", "understand", "understands", "understood",
		"check", "checked", "verify", "verified", "want", "wants", "need", "needs", "act") {
		return false
	}
	for _, token := range tokens[i+1:] {
		if frameWord(token, "before", "after", "until", "not", "only", "n't") || strings.HasPrefix(token.Tag, "CD") {
			return false
		}
	}
	return true
}

func readerPredicateStart(tokens []document.Token) int {
	i := 1
	switch {
	case frameWord(tokens[0], "nobody", "everyone", "everybody"):
	case frameWord(tokens[0], "no") && frameWord(tokens[1], "one"):
		i = 2
	default:
		return 0
	}
	if frameWord(tokens[i], "can", "could") {
		i++
		if i < len(tokens) && frameWord(tokens[i], "have") {
			i++
		}
	}
	return i
}
