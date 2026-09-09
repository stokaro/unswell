package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

// LexicalCountContract identifies word and character counts on prepared targets.
const LexicalCountContract = "unswell-lexical-counts-v1"

// LexicalOptions selects inclusive n-gram orders. A zero pair disables a family.
// Word orders are 1..3 and character orders are 1..6. At least one is required.
type LexicalOptions struct {
	WordMin int `json:"word_min"`
	WordMax int `json:"word_max"`
	CharMin int `json:"char_min"`
	CharMax int `json:"char_max"`
}

// Validate rejects unsupported or empty lexical representations.
func (o LexicalOptions) Validate() error {
	if !lexicalRange(o.WordMin, o.WordMax, 3) || !lexicalRange(o.CharMin, o.CharMax, 6) ||
		(o.WordMax == 0 && o.CharMax == 0) {
		return fmt.Errorf("invalid lexical n-gram orders")
	}
	return nil
}

func lexicalRange(low, high, maximum int) bool {
	return (low == 0 && high == 0) || (low >= 1 && high >= low && high <= maximum)
}

// LexicalTerm is an observed count. Keys explicitly contain source-derived text:
// w: followed by a JSON array of normalized words, or c: followed by Unicode text.
type LexicalTerm struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

// CountLexical returns owned counts in key order, using the existing prepared
// tokens and sentences. Word sequences stop at punctuation and sentence boundaries;
// no stop words, negations, numbers, or identifiers are removed. Character sequences
// use document.Normalize and collapsed whitespace within each sentence, splitting
// at protected boundaries. Counts are descriptive and do not imply a defect.
// Limits cap input bytes/tokens, unique keys, retained key bytes, and occurrences
// (at most MaxBytes*6). Errors return no partial counts.
func CountLexical(ctx context.Context, unit nlp.PreparedUnit, identity Identity, limits Limits,
	options LexicalOptions,
) ([]LexicalTerm, error) {
	block, err := validateLexicalUnit(ctx, unit, identity, limits, options)
	if err != nil {
		return nil, err
	}
	counter := lexicalCounter{ctx: ctx, limits: limits, options: options, counts: make(map[string]int)}
	for _, sentence := range block.Sentences {
		if err := counter.words(sentence.Tokens); err != nil {
			return nil, err
		}
		if err := counter.characters(sentence.Text); err != nil {
			return nil, err
		}
	}
	result := make([]LexicalTerm, 0, len(counter.counts))
	for key, count := range counter.counts {
		result = append(result, LexicalTerm{Key: key, Count: count})
	}
	slices.SortFunc(result, func(a, b LexicalTerm) int { return strings.Compare(a.Key, b.Key) })
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func validateLexicalUnit(ctx context.Context, unit nlp.PreparedUnit, identity Identity, limits Limits,
	options LexicalOptions,
) (document.Block, error) {
	if err := options.Validate(); err != nil {
		return document.Block{}, err
	}
	if unit.Binding().Contract != nlp.UnitContract {
		return document.Block{}, fmt.Errorf("invalid prepared lexical unit")
	}
	if err := matchUnitNLP(unit, identity); err != nil {
		return document.Block{}, err
	}
	block := unit.Block()
	if err := validateInputs(ctx, block, identity, limits); err != nil {
		return document.Block{}, err
	}
	if !slices.Contains(identity.Capabilities, nlp.Tokens) || !slices.Contains(identity.Capabilities, nlp.Sentences) {
		return document.Block{}, fmt.Errorf("lexical counts require tokens and sentences")
	}
	return block, nil
}

type lexicalCounter struct {
	ctx                context.Context
	limits             Limits
	options            LexicalOptions
	counts             map[string]int
	bytes, occurrences int64
}

func (c *lexicalCounter) add(key string) error {
	if err := c.ctx.Err(); err != nil {
		return err
	}
	c.occurrences++
	if c.occurrences > int64(c.limits.MaxBytes)*6 {
		return fmt.Errorf("lexical occurrence budget exceeded")
	}
	if c.counts[key] == 0 {
		c.bytes += int64(len(key))
		if len(c.counts) == c.limits.MaxUniqueWords || c.bytes > int64(c.limits.MaxBytes)*16 {
			return fmt.Errorf("lexical unique-key or retained-byte budget exceeded")
		}
	}
	c.counts[key]++
	return nil
}

func (c *lexicalCounter) words(tokens []document.Token) error {
	for start := range tokens {
		if err := c.wordStart(tokens[start:]); err != nil {
			return err
		}
	}
	return c.ctx.Err()
}

func (c *lexicalCounter) wordStart(tokens []document.Token) error {
	words := make([]string, 0, c.options.WordMax)
	for _, token := range tokens[:min(len(tokens), c.options.WordMax)] {
		if !token.Word || token.Protected {
			break
		}
		words = append(words, token.Normal)
		if len(words) < c.options.WordMin {
			continue
		}
		key, err := json.Marshal(words)
		if err != nil {
			return err
		}
		if err := c.add("w:" + string(key)); err != nil {
			return err
		}
	}
	return c.ctx.Err()
}

func (c *lexicalCounter) characters(text string) error {
	if c.options.CharMax == 0 {
		return c.ctx.Err()
	}
	for piece := range strings.SplitSeq(text, "\x00") {
		runes := []rune(strings.Join(strings.Fields(document.Normalize(piece)), " "))
		for start := range runes {
			for order := c.options.CharMin; order <= c.options.CharMax && start+order <= len(runes); order++ {
				if err := c.add("c:" + string(runes[start:start+order])); err != nil {
					return err
				}
			}
		}
	}
	return c.ctx.Err()
}

// ValidLexicalKey checks a frozen vocabulary entry against its representation.
func ValidLexicalKey(key string, options LexicalOptions) bool {
	if options.Validate() != nil || !utf8.ValidString(key) || strings.ContainsRune(key, 0) {
		return false
	}
	if value, ok := strings.CutPrefix(key, "c:"); ok {
		return validLexicalCharacters(value, options)
	}
	value, ok := strings.CutPrefix(key, "w:")
	return ok && validLexicalWords(value, options)
}

func validLexicalCharacters(value string, options LexicalOptions) bool {
	n := utf8.RuneCountInString(value)
	return n >= options.CharMin && n <= options.CharMax && n != 0 && document.Normalize(value) == value &&
		!strings.Contains(value, "  ") && !strings.ContainsFunc(value, func(r rune) bool { return r != ' ' && unicode.IsSpace(r) })
}

func validLexicalWords(value string, options LexicalOptions) bool {
	var words []string
	if json.Unmarshal([]byte(value), &words) != nil || len(words) == 0 ||
		len(words) < options.WordMin || len(words) > options.WordMax {
		return false
	}
	for _, word := range words {
		if !document.IsWord(word) || document.Normalize(word) != word || strings.ContainsRune(word, 0) {
			return false
		}
	}
	canonical, err := json.Marshal(words)
	return err == nil && string(canonical) == value
}
