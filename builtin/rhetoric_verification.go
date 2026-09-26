package builtin

import "github.com/stokaro/unswell/document"

// Verification rhetoric promotes a testing activity into an unrestricted quality
// guarantee, or declares that nothing requires trust without naming its scope.
// Concrete invariants, exhaustive proofs, and qualified assertions stay outside
// these surface constructions. The diagnostic requests scope, not deletion.
func verificationAssurance(c frameClause) (rhetoricalFrame, bool) {
	if !c.eligible() || contextualRhetoricScoped(c) || verificationQualified(c.sentence.Tokens) {
		return rhetoricalFrame{}, false
	}
	for i := range c.tokens() {
		if verificationStart(c.tokens(), i) && verificationClaim(c.tokens()[i:]) {
			return localFrame(c, i)
		}
	}
	return rhetoricalFrame{}, false
}

func verificationStart(tokens []document.Token, i int) bool {
	return embeddedClauseStart(tokens, i) || i > 0 && frameWord(tokens[i-1], "so") && embeddedClauseStart(tokens, i-1)
}

func verificationQualified(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "not", "n't", "never", "cannot", "can't", "only", "must", "should", "could", "would",
			"may", "might", "requires", "claim", "for", "under", "within", "assuming", ":") {
			return true
		}
	}
	return false
}

func verificationClaim(tokens []document.Token) bool {
	return passiveQualityGuarantee(tokens) || activeQualityGuarantee(tokens) || unrestrictedTrust(tokens)
}

func activeQualityGuarantee(tokens []document.Token) bool {
	for i := 1; i+1 < len(tokens); i++ {
		if frameWord(tokens[i], "ensures", "ensure", "guarantees", "guarantee") &&
			testingActivity(tokens[:i]) && abstractQuality(tokens[i+1:]) {
			return true
		}
	}
	return false
}

func passiveQualityGuarantee(tokens []document.Token) bool {
	for i := 1; i+3 < len(tokens); i++ {
		if !frameWord(tokens[i], "is", "are", "was", "were") || !abstractQuality(tokens[:i]) {
			continue
		}
		rest := tokens[i+1:]
		if frameWord(rest[0], "further", "also") {
			rest = rest[1:]
		}
		return len(rest) >= 3 && frameWord(rest[0], "ensured", "guaranteed") &&
			frameWord(rest[1], "by") && testingActivity(rest[2:])
	}
	return false
}

func abstractQuality(tokens []document.Token) bool {
	if len(tokens) > 0 && frameWord(tokens[0], "the") {
		tokens = tokens[1:]
	}
	if len(tokens) == 0 || !qualityGuaranteeNoun(tokens[0]) {
		return false
	}
	if len(tokens) == 1 {
		return true
	}
	if len(tokens) >= 3 && frameWord(tokens[1], "of") {
		return positiveRhetoricSubject(tokens[2:])
	}
	return len(tokens) == 3 && frameWord(tokens[1], "and") && qualityGuaranteeNoun(tokens[2])
}

func qualityGuaranteeNoun(token document.Token) bool {
	return frameWord(token, "reliability", "security", "safety", "correctness", "quality")
}

func testingActivity(tokens []document.Token) bool {
	if len(tokens) > 0 && frameWord(tokens[0], "the", "our") {
		tokens = tokens[1:]
	}
	if len(tokens) == 0 || len(tokens) > 5 {
		return false
	}
	last := tokens[len(tokens)-1]
	if !frameWord(last, "tests", "testing", "reviews", "coverage") {
		return false
	}
	for _, token := range tokens[:len(tokens)-1] {
		if !frameWord(token, "rigorous", "extensive", "comprehensive", "thorough", "automated", "continuous",
			"unit", "integration", "regression", "stress", "code", "test") {
			return false
		}
	}
	return true
}

func unrestrictedTrust(tokens []document.Token) bool {
	if len(tokens) != 5 || !frameWord(tokens[3], "on") || !frameWord(tokens[4], "trust", "faith") {
		return false
	}
	return trustQuantifier(tokens[:3])
}

func trustQuantifier(tokens []document.Token) bool {
	return frameWord(tokens[0], "we") && frameWord(tokens[1], "take", "accept") && frameWord(tokens[2], "nothing") ||
		frameWord(tokens[0], "nothing") && frameWord(tokens[1], "is", "was") && frameWord(tokens[2], "taken", "accepted")
}
