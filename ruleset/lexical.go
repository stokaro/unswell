package ruleset

import (
	"fmt"
	"regexp"
	"regexp/syntax"
	"slices"
	"strings"

	"github.com/jdkato/prose/v3/tokenize"
	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type phraseSpec struct {
	Type          string   `yaml:"type"`
	At            string   `yaml:"at"`
	CaseSensitive bool     `yaml:"case_sensitive"`
	Values        []string `yaml:"values"`
}

type phraseMatcher struct {
	patterns  [][]string
	at        string
	sensitive bool
}

func compilePhrase(_ *compiler, node yaml.Node, _ int) (matcher, error) {
	var spec phraseSpec
	if err := decodeNode(node, &spec); err != nil {
		return nil, err
	}
	if err := validateAt(spec.At); err != nil {
		return nil, err
	}
	if len(spec.Values) == 0 || len(spec.Values) > 100 {
		return nil, fmt.Errorf("phrase and token-set matchers require 1 to 100 values")
	}
	m := &phraseMatcher{at: spec.At, sensitive: spec.CaseSensitive}
	for _, value := range spec.Values {
		pattern, err := compileValue(value, spec)
		if err != nil {
			return nil, err
		}
		m.patterns = append(m.patterns, pattern)
	}
	return m, nil
}

func compileValue(value string, spec phraseSpec) ([]string, error) {
	if len(value) == 0 || len(value) > 1000 || strings.ContainsRune(value, 0) {
		return nil, fmt.Errorf("phrase values require 1 to 1000 bytes without protected boundaries")
	}
	pattern := patternTokens(value, spec.CaseSensitive)
	if len(pattern) == 0 || spec.Type == "token-set" && len(pattern) != 1 {
		return nil, fmt.Errorf("token-set values must tokenize to exactly one token; phrases cannot be empty")
	}
	return pattern, nil
}

func patternTokens(text string, sensitive bool) []string {
	if !sensitive {
		text = document.Normalize(text)
	}
	var result []string
	for _, token := range tokenize.New().Tokenize(text) {
		result = append(result, token.Text)
	}
	return result
}

func validateAt(at string) error {
	if !slices.Contains([]string{"", "any", "start", "end"}, at) {
		return fmt.Errorf("at must be any, start, or end")
	}
	return nil
}

func positioned(e *evaluation, at string, sentence *document.Sentence, start, end int) (bool, error) {
	if at == "start" {
		return start == 0, nil
	}
	if at != "end" {
		return true, nil
	}
	requiredEnd, err := e.sentenceEnd(sentence)
	return end >= requiredEnd, err
}

func (e *evaluation) sentenceEnd(sentence *document.Sentence) (int, error) {
	if e.ends == nil {
		e.ends = make(map[*document.Sentence]int)
	}
	if end, ok := e.ends[sentence]; ok {
		return end, nil
	}
	end := len(sentence.Tokens)
	for end > 0 {
		if err := e.spend(1); err != nil {
			return 0, err
		}
		token := sentence.Tokens[end-1]
		if token.Word || token.Protected {
			break
		}
		end--
	}
	e.ends[sentence] = end
	return end, nil
}

func (m *phraseMatcher) evaluate(e *evaluation, current unit) (matchResult, error) {
	var result matchResult
	for _, view := range current {
		for start := range view.sentence.Tokens {
			for _, pattern := range m.patterns {
				if err := m.matchAt(e, view.sentence, start, pattern, &result); err != nil {
					return matchResult{}, err
				}
			}
		}
	}
	result.occurrences = uniqueOccurrences(result.occurrences)
	return result, nil
}

func (m *phraseMatcher) matchAt(
	e *evaluation, sentence *document.Sentence, start int, pattern []string, result *matchResult,
) error {
	if err := e.spend(1); err != nil {
		return err
	}
	end := start + len(pattern)
	if end > len(sentence.Tokens) {
		return nil
	}
	position, err := positioned(e, m.at, sentence, start, end)
	if err != nil || !position {
		return err
	}
	for i, value := range pattern {
		if err := e.spend(1); err != nil {
			return err
		}
		token := sentence.Tokens[start+i]
		text := token.Normal
		if m.sensitive {
			text = token.Text
		}
		if token.Protected || text != value {
			return nil
		}
	}
	return addOccurrence(result, tokenOccurrence(sentence, start, end))
}

type regexSpec struct {
	Type          string `yaml:"type"`
	Pattern       string `yaml:"pattern"`
	CaseSensitive bool   `yaml:"case_sensitive"`
}

type regexMatcher struct{ pattern *regexp.Regexp }

func compileRegex(_ *compiler, node yaml.Node, _ int) (matcher, error) {
	var spec regexSpec
	if err := decodeNode(node, &spec); err != nil {
		return nil, err
	}
	if len(spec.Pattern) == 0 || len(spec.Pattern) > 4096 {
		return nil, fmt.Errorf("regex requires 1 to 4096 pattern bytes")
	}
	pattern := spec.Pattern
	if !spec.CaseSensitive {
		pattern = "(?i:" + pattern + ")"
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("RE2 pattern: %w", err)
	}
	tree, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return nil, err
	}
	if nullableRegex(tree) {
		return nil, fmt.Errorf("every regex alternative must consume text; zero-width matches are unsupported")
	}
	return &regexMatcher{pattern: compiled}, nil
}

func nullableRegex(tree *syntax.Regexp) bool {
	switch tree.Op {
	case syntax.OpEmptyMatch, syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText, syntax.OpEndText,
		syntax.OpWordBoundary, syntax.OpNoWordBoundary, syntax.OpStar, syntax.OpQuest:
		return true
	case syntax.OpCapture, syntax.OpPlus:
		return nullableRegex(tree.Sub[0])
	case syntax.OpRepeat:
		return tree.Min == 0 || nullableRegex(tree.Sub[0])
	case syntax.OpConcat, syntax.OpAlternate:
		all := tree.Op == syntax.OpConcat
		for _, child := range tree.Sub {
			if nullableRegex(child) != all {
				return !all
			}
		}
		return all
	case syntax.OpNoMatch, syntax.OpLiteral, syntax.OpCharClass, syntax.OpAnyCharNotNL, syntax.OpAnyChar:
		return false
	}
	return false
}

func (m *regexMatcher) evaluate(e *evaluation, current unit) (matchResult, error) {
	var result matchResult
	for _, view := range current {
		if err := m.sentence(e, view, &result); err != nil {
			return matchResult{}, err
		}
	}
	return result, nil
}

func (m *regexMatcher) sentence(e *evaluation, view sentenceView, result *matchResult) error {
	tokens := view.sentence.Tokens
	if len(tokens) == 0 {
		return nil
	}
	start, end := tokens[0].Start, tokens[len(tokens)-1].End
	if start < 0 || end > len(view.block.Text) || end < start {
		return fmt.Errorf("invalid mapped token range")
	}
	for fragment := range strings.SplitSeq(view.block.Text[start:end], "\x00") {
		if err := m.fragment(e, view, fragment, start, result); err != nil {
			return err
		}
		start += len(fragment) + 1
	}
	return nil
}

func (m *regexMatcher) fragment(e *evaluation, view sentenceView, text string, offset int, result *matchResult) error {
	if err := e.spend(1 + len(text)/64); err != nil {
		return err
	}
	matches := m.pattern.FindAllStringIndex(text, min(e.remaining+1, 10001))
	for _, match := range matches {
		if err := e.spend(1); err != nil {
			return err
		}
		if match[0] == match[1] {
			continue
		}
		occurrence := rule.Occurrence{BlockID: view.block.ID, SentenceID: view.sentence.ID,
			Spans: view.block.Spans(offset+match[0], offset+match[1])}
		if err := addOccurrence(result, occurrence); err != nil {
			return err
		}
	}
	return e.ctx.Err()
}
