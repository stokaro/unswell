package builtin

import "github.com/stokaro/unswell/document"

// Attribution before a candidate can introduce reported speech across a colon.
// Attribution after its clause boundary cannot qualify an earlier judgment.
func rhetoricAttributed(c frameClause) bool {
	for _, token := range c.sentence.Tokens[:c.end] {
		if frameWord(token, "according", "claims", "claimed", "says", "said", "argues", "argued",
			"rejects", "refutes", "disputes", "reported", "quoted") {
			return true
		}
	}
	return false
}

func rhetoricQuoteOpen(c frameClause) bool {
	open := false
	for _, token := range c.sentence.Tokens[:c.start] {
		if quotedClaim([]document.Token{token}) {
			open = !open
		}
	}
	return open
}

func rhetoricQuoted(c frameClause) bool {
	// Closed quotations in an earlier clause do not turn the writer's following
	// judgment into a quote; an unmatched quote may still contain the candidate.
	return rhetoricQuoteOpen(c) || quotedClaim(c.tokens())
}

func evaluationScoped(c frameClause) bool {
	if rhetoricAttributed(c) {
		return true
	}
	for i, token := range c.tokens() {
		if token.Tag == "CD" && !indefiniteOne(c.tokens(), i) ||
			frameWord(token, "if", "when", "unless", "whenever", "except", "without", "measured", "survey", "surveyed",
				"benchmark", "benchmarks", "median", "percentile", "respondents") {
			return true
		}
	}
	return false
}
