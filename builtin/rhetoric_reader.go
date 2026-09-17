package builtin

import "github.com/stokaro/unswell/document"

func readerPrevalence(tokens []document.Token) bool {
	if readerMajority(tokens) {
		return true
	}
	for i := 1; i < min(len(tokens)-2, 10); i++ {
		if !frameWord(tokens[i], "is", "are", "was", "were") || !positiveRhetoricSubject(tokens[:i]) {
			continue
		}
		rest := tokens[i+1:]
		return readerPreference(rest) || majorityPractice(rest) || majoritySafety(rest)
	}
	return false
}

func readerPreference(tokens []document.Token) bool {
	return len(tokens) == 5 && frameWord(tokens[0], "almost", "nearly") && frameWord(tokens[1], "always", "never") &&
		frameWord(tokens[2], "what") && frameWord(tokens[3], "you", "users", "people") &&
		frameWord(tokens[4], "want", "need", "prefer")
}

func majorityPractice(tokens []document.Token) bool {
	return len(tokens) >= 4 && frameWord(tokens[0], "what") && readerMajority(tokens[1:])
}

func majoritySafety(tokens []document.Token) bool {
	return len(tokens) == 4 && frameWord(tokens[0], "safe", "suitable", "ideal") && frameWord(tokens[1], "for") &&
		frameWord(tokens[2], "most") && frameWord(tokens[3], "runs", "workloads", "users", "people", "environments")
}

func readerMajority(tokens []document.Token) bool {
	if len(tokens) < 3 || !frameWord(tokens[0], "most") {
		return false
	}
	i := 1
	if frameWord(tokens[i], "production") {
		i++
	}
	if i+1 >= len(tokens) || !frameWord(tokens[i], "people", "users", "operators", "environments") ||
		!frameWord(tokens[i+1], "want", "need", "prefer", "do", "run", "use") {
		return false
	}
	for _, token := range tokens[i+2:] {
		if frameWord(token, "only", "not", "n't", "because", "before", "until", "survey", "surveyed", "respondents") || token.Tag == "CD" {
			return false
		}
	}
	return true
}

// Popularity and reader-preference claims are distinct from mechanisms. A
// recommendation following a colon cannot establish their claimed prevalence.
func readerAssurance(c frameClause) (rhetoricalFrame, bool) {
	if len(c.sentence.Tokens) > 96 || question(c.sentence) {
		return rhetoricalFrame{}, false
	}
	for i := range c.tokens() {
		candidate, ok := localRhetoricCandidate(c, i)
		if !ok {
			continue
		}
		if frame, found := readerCandidate(c, candidate, i); found {
			return frame, true
		}
	}
	return rhetoricalFrame{}, false
}

func readerCandidate(c, candidate frameClause, start int) (rhetoricalFrame, bool) {
	tokens := candidate.tokens()
	boundary := embeddedClauseStart(c.tokens(), start)
	majority := boundary && (readerPrevalence(tokens) || readerResult(tokens))
	if start > 0 && frameWord(c.tokens()[start-1], "what", "policy") {
		majority = majority || readerMajority(tokens)
	}
	if majority && !candidateRhetoricScoped(candidate) {
		return localFrame(candidate, 0)
	}
	if boundary {
		return boundedReaderJudgment(candidate)
	}
	return rhetoricalFrame{}, false
}

func boundedReaderJudgment(candidate frameClause) (rhetoricalFrame, bool) {
	if end := readerJudgmentEnd(candidate.tokens()); end > 0 {
		candidate.end = candidate.start + end
		if !candidateRhetoricScoped(candidate) {
			return localFrame(candidate, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func readerJudgmentEnd(tokens []document.Token) int {
	for i := 1; i < min(len(tokens)-2, 8); i++ {
		if frameWord(tokens[i], "is", "was") && positiveRhetoricSubject(tokens[:i]) {
			if end := nearlyAlwaysRight(tokens[i+1:]); end > 0 {
				return i + 1 + end
			}
		}
	}
	return 0
}
