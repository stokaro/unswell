package feature

import (
	"context"
	"fmt"
	"strings"

	"github.com/stokaro/unswell/document"
)

// NgramOptions selects the existing 3-8-word candidate sizes and limits attempted
// token visits. Punctuation and protected boundaries consume visits too.
type NgramOptions struct {
	MinWords, MaxWords, MaxVisits int
}

// Ngram contains an owned normalized key and half-open input token indices.
// The caller maps these indices through original token spans; gaps are not filled.
type Ngram struct {
	Start, End int
	Key        string
}

// ScanNgrams streams candidates in start/end token order without retaining input
// buffers. It returns the logical visits spent, including on error. Input tokens
// must remain unchanged during the call. Callback errors stop the scan unchanged;
// consumers must discard partial results on any error. The callback applies term
// and protected-signature policy before grouping; keys explicitly expose text.
func ScanNgrams(ctx context.Context, tokens []document.Token, options NgramOptions, limits SequenceLimits,
	yield func(Ngram) error) (int, error) {
	if err := validateSequence(ctx, tokens, 0, limits); err != nil {
		return 0, err
	}
	if options.MinWords < 3 || options.MaxWords > 8 || options.MinWords > options.MaxWords ||
		options.MaxVisits < 0 || options.MaxVisits > 1<<30 || yield == nil {
		return 0, fmt.Errorf("invalid n-gram options or callback")
	}
	scanner := ngramScanner{ctx: ctx, tokens: tokens, options: options, yield: yield}
	for start := range tokens {
		if err := scanner.from(start); err != nil {
			return scanner.visits, err
		}
	}
	return scanner.visits, ctx.Err()
}

type ngramScanner struct {
	ctx     context.Context
	tokens  []document.Token
	options NgramOptions
	yield   func(Ngram) error
	visits  int
}

func (s *ngramScanner) from(start int) error {
	var words []string
	for end := start; end < min(len(s.tokens), start+s.options.MaxWords); end++ {
		if err := s.spend(); err != nil {
			return err
		}
		token := s.tokens[end]
		if !token.Word || token.Protected {
			break
		}
		words = append(words, token.Normal)
		if len(words) >= s.options.MinWords && ngramContent(words) {
			if err := s.yield(Ngram{Start: start, End: end + 1, Key: strings.Join(words, " ")}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ngramScanner) spend() error {
	if err := s.ctx.Err(); err != nil {
		return err
	}
	if s.visits == s.options.MaxVisits {
		return ErrTokenLimit
	}
	s.visits++
	return nil
}

func ngramContent(words []string) bool {
	first := ""
	for _, word := range words {
		if InformativeWord(word) {
			if first != "" && word != first {
				return true
			}
			first = word
		}
	}
	return false
}
