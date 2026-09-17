package builtin

import (
	"strconv"
	"strings"

	"github.com/stokaro/unswell/document"
)

// claimIdentity copies prose bytes while checking their source map. A separate
// quoted serialization would traverse every prose token a second time.
func claimIdentity(source []byte, block document.Block, tokens []document.Token, budget *repetitionBudget) (string, bool, error) {
	start, end := tokens[0].Start, tokens[len(tokens)-1].End
	if start < 0 || end > len(block.Map) || start >= end {
		return "", false, nil
	}
	walk := claimIdentityWalk{source: source, mapped: block.Map, budget: budget,
		position: start, previous: block.Map[start].Start, previousStart: block.Map[start].Start}
	opaque := false
	for _, token := range tokens {
		if err := budget.spend(1); err != nil {
			return "", false, err
		}
		ok, err := walk.token(token)
		if err != nil || !ok {
			return "", false, err
		}
		opaque = opaque || token.Protected
	}
	return walk.key.String(), opaque, nil
}

type claimIdentityWalk struct {
	source             []byte
	mapped             []document.Span
	budget             *repetitionBudget
	key                strings.Builder
	position, previous int
	previousStart      int
}

func (w *claimIdentityWalk) token(token document.Token) (bool, error) {
	value := token.Text
	if token.Protected {
		atom, err := opaqueClaimAtom(w.source, token, w.budget)
		if err != nil || atom == "" {
			return false, err
		}
		value = atom
	}
	if !w.validToken(token) {
		return false, nil
	}
	if ok, err := w.advance(token.Start, ""); err != nil || !ok {
		return false, err
	}
	// Length prefixes preserve token boundaries without scanning or escaping the
	// value. The marker distinguishes a code operand from the same prose bytes.
	w.key.WriteString(strconv.FormatBool(token.Protected))
	w.key.WriteByte(':')
	w.key.WriteString(strconv.Itoa(len(value)))
	w.key.WriteByte(':')
	if token.Protected {
		w.key.WriteString(value)
		value = ""
	}
	return w.advance(token.End, value)
}

func (w *claimIdentityWalk) validToken(token document.Token) bool {
	return token.Start >= w.position && token.End <= len(w.mapped) && token.End > token.Start &&
		len(token.Text) == token.End-token.Start
}

// Only emphasis and whitespace may disappear between mapped source segments.
// Link destinations must remain a reason to reject the whole assertion.
func (w *claimIdentityWalk) advance(end int, text string) (bool, error) {
	if err := w.budget.spend(end - w.position); err != nil {
		return false, err
	}
	start := w.position
	for ; w.position < end; w.position++ {
		span := w.mapped[w.position]
		if !span.Valid(len(w.source)) || span.Start < w.previousStart {
			return false, nil
		}
		if span.Start > w.previous && !emphasisGap(string(w.source[w.previous:span.Start])) {
			return false, nil
		}
		w.previous, w.previousStart = max(w.previous, span.End), span.Start
		if text != "" {
			w.key.WriteByte(text[w.position-start])
		}
	}
	return true, nil
}
