package llmdet

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// BPE is the byte-level byte-pair tokenizer of the GPT-2 family, read from
// a vocabulary and an ordered merge list. It reproduces the reference
// pre-tokenization without a backtracking regular expression, so the same
// text yields the same token IDs as the published tokenizer files.
type BPE struct {
	vocab   map[string]int
	ranks   map[[2]string]int
	symbols [256]string
	mutex   sync.Mutex
	cache   map[string][]int
}

const (
	// MaxBPEBytes bounds one vocabulary or merge file.
	MaxBPEBytes = 16 << 20
	// MaxBPEText bounds one text handed to Encode.
	MaxBPEText  = 1 << 20
	maxBPECache = 65536
	maxBPEVocab = 1 << 20
)

// LoadBPE reads a vocabulary (token to ID) and a merge list (one pair per
// line after the version comment) and checks that every merge names known
// symbols.
func LoadBPE(ctx context.Context, vocabulary, merges []byte) (*BPE, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(vocabulary) > MaxBPEBytes || len(merges) > MaxBPEBytes {
		return nil, fmt.Errorf("tokenizer files exceed %d bytes", MaxBPEBytes)
	}
	vocab, err := parseVocabulary(vocabulary)
	if err != nil {
		return nil, err
	}
	ranks, err := parseMerges(ctx, merges, vocab)
	if err != nil {
		return nil, err
	}
	b := &BPE{vocab: vocab, ranks: ranks, cache: make(map[string][]int)}
	b.symbols = byteSymbols()
	return b, nil
}

func parseVocabulary(vocabulary []byte) (map[string]int, error) {
	var vocab map[string]int
	decoder := json.NewDecoder(bytes.NewReader(vocabulary))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&vocab); err != nil {
		return nil, fmt.Errorf("tokenizer vocabulary: %w", err)
	}
	if len(vocab) == 0 || len(vocab) > maxBPEVocab {
		return nil, fmt.Errorf("tokenizer vocabulary requires 1 to %d entries", maxBPEVocab)
	}
	seen := make(map[int]bool, len(vocab))
	for token, id := range vocab {
		if token == "" || id < 0 || id >= maxBPEVocab || seen[id] || !utf8.ValidString(token) {
			return nil, fmt.Errorf("tokenizer vocabulary has an invalid or repeated entry")
		}
		seen[id] = true
	}
	return vocab, nil
}

func parseMerges(ctx context.Context, merges []byte, vocab map[string]int) (map[[2]string]int, error) {
	ranks := make(map[[2]string]int)
	scanner := bufio.NewScanner(bytes.NewReader(merges))
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pair, err := parseMergeLine(line, vocab)
		if err != nil {
			return nil, err
		}
		if _, exists := ranks[pair]; exists {
			return nil, fmt.Errorf("merge %q is repeated", line)
		}
		ranks[pair] = len(ranks)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(ranks) == 0 {
		return nil, fmt.Errorf("tokenizer requires at least one merge")
	}
	return ranks, nil
}

func parseMergeLine(line string, vocab map[string]int) ([2]string, error) {
	left, right, found := strings.Cut(line, " ")
	if !found || left == "" || right == "" || strings.Contains(right, " ") {
		return [2]string{}, fmt.Errorf("merge line %q is not one pair", line)
	}
	if _, exists := vocab[left+right]; !exists {
		return [2]string{}, fmt.Errorf("merge %q produces a token outside the vocabulary", line)
	}
	return [2]string{left, right}, nil
}

// byteSymbols maps every byte to the reference's printable symbol: the
// printable Latin ranges stand for themselves and every other byte takes
// the next code point from 256 upward.
func byteSymbols() [256]string {
	var symbols [256]string
	printable := func(b int) bool {
		return (b >= '!' && b <= '~') || (b >= 0xa1 && b <= 0xac) || (b >= 0xae && b <= 0xff)
	}
	next := 256
	for b := range 256 {
		if printable(b) {
			symbols[b] = string(rune(b))
			continue
		}
		symbols[b] = string(rune(next))
		next++
	}
	return symbols
}

// Encode returns the token IDs of a text. A text the vocabulary cannot
// represent is an error, never a silent unknown token.
func (b *BPE) Encode(ctx context.Context, text string) ([]int, error) {
	if b == nil {
		return nil, fmt.Errorf("missing tokenizer")
	}
	if len(text) > MaxBPEText || !utf8.ValidString(text) {
		return nil, fmt.Errorf("tokenizer text requires valid UTF-8 within %d bytes", MaxBPEText)
	}
	var ids []int
	for _, piece := range pretokenize(text) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		tokens, err := b.piece(piece)
		if err != nil {
			return nil, err
		}
		ids = append(ids, tokens...)
	}
	return ids, nil
}

func (b *BPE) piece(piece string) ([]int, error) {
	b.mutex.Lock()
	cached, exists := b.cache[piece]
	b.mutex.Unlock()
	if exists {
		return cached, nil
	}
	symbols := make([]string, 0, len(piece))
	for i := range len(piece) {
		symbols = append(symbols, b.symbols[piece[i]])
	}
	merged := b.merge(symbols)
	ids := make([]int, 0, len(merged))
	for _, symbol := range merged {
		id, known := b.vocab[symbol]
		if !known {
			return nil, fmt.Errorf("tokenizer vocabulary lacks %q", symbol)
		}
		ids = append(ids, id)
	}
	b.mutex.Lock()
	if len(b.cache) >= maxBPECache {
		b.cache = make(map[string][]int)
	}
	b.cache[piece] = ids
	b.mutex.Unlock()
	return ids, nil
}

// merge applies the lowest-ranked adjacent pair until no pair has a rank,
// as the reference does.
func (b *BPE) merge(symbols []string) []string {
	for len(symbols) > 1 {
		pair, found := b.lowestPair(symbols)
		if !found {
			break
		}
		joined := make([]string, 0, len(symbols))
		for i := 0; i < len(symbols); i++ {
			if i+1 < len(symbols) && symbols[i] == pair[0] && symbols[i+1] == pair[1] {
				joined = append(joined, pair[0]+pair[1])
				i++
				continue
			}
			joined = append(joined, symbols[i])
		}
		symbols = joined
	}
	return symbols
}

// lowestPair finds the adjacent pair with the lowest merge rank.
func (b *BPE) lowestPair(symbols []string) ([2]string, bool) {
	best, at := -1, -1
	for i := 0; i+1 < len(symbols); i++ {
		rank, exists := b.ranks[[2]string{symbols[i], symbols[i+1]}]
		if exists && (best < 0 || rank < best) {
			best, at = rank, i
		}
	}
	if at < 0 {
		return [2]string{}, false
	}
	return [2]string{symbols[at], symbols[at+1]}, true
}

// pretokenize splits a text as the reference pattern does. A piece is a
// contraction, or an optional space with a run of letters, digits, or other
// marks. A run of whitespace leaves its last character to the next piece.
// A lone whitespace character is a piece of its own.
func pretokenize(text string) []string {
	runes := []rune(text)
	var pieces []string
	for i := 0; i < len(runes); {
		if end, ok := contraction(runes, i); ok {
			pieces = append(pieces, string(runes[i:end]))
			i = end
			continue
		}
		if end, ok := prefixedRun(runes, i); ok {
			pieces = append(pieces, string(runes[i:end]))
			i = end
			continue
		}
		end := i
		for end < len(runes) && isSpace(runes[end]) {
			end++
		}
		if end < len(runes) && end-1 > i {
			end--
		}
		pieces = append(pieces, string(runes[i:end]))
		i = end
	}
	return pieces
}

var contractions = []string{"'s", "'t", "'re", "'ve", "'m", "'ll", "'d"}

func contraction(runes []rune, i int) (int, bool) {
	if runes[i] != '\'' {
		return 0, false
	}
	rest := string(runes[i:min(len(runes), i+3)])
	for _, candidate := range contractions {
		if strings.HasPrefix(rest, candidate) {
			return i + len([]rune(candidate)), true
		}
	}
	return 0, false
}

// prefixedRun matches an optional space followed by a run of one class:
// letters, digits, or marks that are neither whitespace nor letters nor
// digits.
func prefixedRun(runes []rune, i int) (int, bool) {
	start := i
	if runes[start] == ' ' && start+1 < len(runes) {
		start++
	}
	if start >= len(runes) {
		return 0, false
	}
	class := runeClass(runes[start])
	if class == classSpace {
		return 0, false
	}
	end := start
	for end < len(runes) && runeClass(runes[end]) == class {
		end++
	}
	return end, true
}

const (
	classSpace = iota
	classLetter
	classNumber
	classOther
)

func runeClass(r rune) int {
	switch {
	case isSpace(r):
		return classSpace
	case unicode.IsLetter(r):
		return classLetter
	case unicode.IsNumber(r):
		return classNumber
	}
	return classOther
}

// isSpace is the reference's whitespace class: Unicode white space and the
// information separators U+001C to U+001F.
func isSpace(r rune) bool {
	return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f)
}
