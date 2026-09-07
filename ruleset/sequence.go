package ruleset

import (
	"fmt"
	"slices"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

type gapSpec struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

type tokenSpec struct {
	Value  string   `yaml:"value"`
	POS    string   `yaml:"pos"`
	Not    bool     `yaml:"not"`
	Target bool     `yaml:"target"`
	Gap    *gapSpec `yaml:"gap"`
}

type sequenceSpec struct {
	Type          string      `yaml:"type"`
	At            string      `yaml:"at"`
	CaseSensitive bool        `yaml:"case_sensitive"`
	Tokens        []tokenSpec `yaml:"tokens"`
}

type sequenceMatcher struct {
	at        string
	sensitive bool
	targeted  bool
	atoms     []tokenSpec
}

func compileSequence(c *compiler, node yaml.Node, _ int) (matcher, error) {
	var spec sequenceSpec
	if err := decodeNode(node, &spec); err != nil {
		return nil, err
	}
	if err := validateAt(spec.At); err != nil {
		return nil, err
	}
	if len(spec.Tokens) == 0 || len(spec.Tokens) > 32 {
		return nil, fmt.Errorf("sequence requires 1 to 32 token or gap entries")
	}
	c.nodes += len(spec.Tokens)
	if c.nodes > 512 {
		return nil, fmt.Errorf("sequence exceeds 512 matcher nodes")
	}
	m := &sequenceMatcher{at: spec.At, sensitive: spec.CaseSensitive, atoms: spec.Tokens}
	for i := range m.atoms {
		if err := m.validateAtom(c, i); err != nil {
			return nil, fmt.Errorf("sequence token %d: %w", i+1, err)
		}
		m.targeted = m.targeted || m.atoms[i].Target
	}
	return m, nil
}

func (m *sequenceMatcher) validateAtom(c *compiler, index int) error {
	atom := &m.atoms[index]
	if atom.Gap != nil {
		return m.validateGap(index)
	}
	if atom.Value == "" && atom.POS == "" {
		return fmt.Errorf("token needs value or pos")
	}
	if atom.Value != "" {
		values := patternTokens(atom.Value, m.sensitive)
		if len(atom.Value) > 1000 || len(values) != 1 {
			return fmt.Errorf("value must tokenize to one token of at most 1000 bytes")
		}
		atom.Value = values[0]
	}
	if atom.POS != "" {
		if !validPOS(atom.POS) {
			return fmt.Errorf("unknown Penn Treebank tag %q", atom.POS)
		}
		c.require(nlp.POS)
	}
	return nil
}

func (m *sequenceMatcher) validateGap(index int) error {
	atom := m.atoms[index]
	if !m.gapPosition(index) {
		return fmt.Errorf("gaps must occur between tokens and cannot be adjacent")
	}
	if atom.Value != "" || atom.POS != "" || atom.Not || atom.Target {
		return fmt.Errorf("gap cannot have token constraints or a target")
	}
	if atom.Gap.Min < 0 || atom.Gap.Max < atom.Gap.Min || atom.Gap.Max > 32 {
		return fmt.Errorf("gap must satisfy 0 <= min <= max <= 32")
	}
	return nil
}

func (m *sequenceMatcher) gapPosition(index int) bool {
	return index > 0 && index < len(m.atoms)-1 && m.atoms[index-1].Gap == nil
}

func validPOS(tag string) bool {
	return slices.Contains([]string{
		"CC", "CD", "DT", "EX", "FW", "IN", "JJ", "JJR", "JJS", "LS", "MD", "NN", "NNS", "NNP", "NNPS",
		"PDT", "POS", "PRP", "PRP$", "RB", "RBR", "RBS", "RP", "SYM", "TO", "UH", "VB", "VBD", "VBG",
		"VBN", "VBP", "VBZ", "WDT", "WP", "WP$", "WRB", "#", "$", ".", ",", ":", "``", "''", "-LRB-", "-RRB-",
	}, tag)
}

type sequenceResult struct {
	end     int
	targets []int
	ok      bool
}

type sequenceWalk struct {
	matcher  *sequenceMatcher
	eval     *evaluation
	sentence *document.Sentence
	start    int
	memo     map[[2]int]sequenceResult
}

func (m *sequenceMatcher) evaluate(e *evaluation, current unit) (matchResult, error) {
	var result matchResult
	for _, view := range current {
		if err := m.sentence(e, view.sentence, &result); err != nil {
			return matchResult{}, err
		}
	}
	result.occurrences = uniqueOccurrences(result.occurrences)
	return result, nil
}

func (m *sequenceMatcher) sentence(e *evaluation, sentence *document.Sentence, result *matchResult) error {
	for start := range sentence.Tokens {
		if m.at == "start" && start != 0 {
			break
		}
		walk := sequenceWalk{matcher: m, eval: e, sentence: sentence, start: start, memo: make(map[[2]int]sequenceResult)}
		matched, err := walk.visit(0, start)
		if err != nil {
			return err
		}
		if matched.ok {
			if err := addOccurrence(result, walk.occurrence(matched)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *sequenceWalk) visit(atomIndex, position int) (sequenceResult, error) {
	if err := w.eval.spend(1); err != nil {
		return sequenceResult{}, err
	}
	key := [2]int{atomIndex, position}
	if result, ok := w.memo[key]; ok {
		return result, nil
	}
	result, err := w.next(atomIndex, position)
	if err == nil {
		w.memo[key] = result
	}
	return result, err
}

func (w *sequenceWalk) next(index, position int) (sequenceResult, error) {
	if index == len(w.matcher.atoms) {
		ok, err := positioned(w.eval, w.matcher.at, w.sentence, w.start, position)
		return sequenceResult{end: position, ok: ok}, err
	}
	atom := w.matcher.atoms[index]
	if atom.Gap != nil {
		return w.gap(index, position, *atom.Gap)
	}
	if position >= len(w.sentence.Tokens) {
		return sequenceResult{}, nil
	}
	ok, err := w.matcher.accepts(atom, w.sentence.Tokens[position])
	if err != nil || !ok {
		return sequenceResult{}, err
	}
	result, err := w.visit(index+1, position+1)
	if result.ok && atom.Target {
		result.targets = append([]int{position}, result.targets...)
	}
	return result, err
}

func (m *sequenceMatcher) accepts(atom tokenSpec, token document.Token) (bool, error) {
	if token.Protected {
		return false, nil
	}
	if atom.POS != "" && token.Tag == "" {
		return false, fmt.Errorf("POS capability returned a token without a tag")
	}
	value := token.Normal
	if m.sensitive {
		value = token.Text
	}
	ok := (atom.Value == "" || atom.Value == value) && (atom.POS == "" || atom.POS == token.Tag)
	if atom.Not {
		ok = !ok
	}
	return ok, nil
}

func (w *sequenceWalk) gap(index, position int, gap gapSpec) (sequenceResult, error) {
	for skip := 0; skip <= gap.Max && position+skip <= len(w.sentence.Tokens); skip++ {
		if skip > 0 && w.sentence.Tokens[position+skip-1].Protected {
			break
		}
		if skip < gap.Min {
			continue
		}
		result, err := w.visit(index+1, position+skip)
		if err != nil || result.ok {
			return result, err
		}
	}
	return sequenceResult{}, nil
}

func (w *sequenceWalk) occurrence(result sequenceResult) rule.Occurrence {
	if !w.matcher.targeted {
		return tokenOccurrence(w.sentence, w.start, result.end)
	}
	occurrence := rule.Occurrence{BlockID: w.sentence.BlockID, SentenceID: w.sentence.ID}
	for _, target := range result.targets {
		occurrence.Spans = append(occurrence.Spans, w.sentence.Tokens[target].Spans...)
	}
	return occurrence
}
