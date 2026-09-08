package feature_test

import (
	"context"
	"encoding/json"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func testIdentity() feature.Identity {
	return feature.Identity{NLP: nlp.Identity{Name: "fixture", Version: "1",
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS}},
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS},
		Source:       "fixture-v1", Policy: "technical-v1", Vocabulary: "none", Preprocessing: "fixture-normal-v1"}
}

func testLimits() feature.Limits {
	return feature.Limits{MaxTokens: 1000, MaxUniqueWords: 1000, MaxBytes: 10000, MaxBlocks: 100}
}

func testBlock(rows ...string) document.Block {
	block := document.Block{Kind: "paragraph"}
	var text strings.Builder
	for i, row := range rows {
		sentence := document.Sentence{ID: i}
		for word := range strings.FieldsSeq(row) {
			start := text.Len()
			text.WriteString(word)
			sentence.Tokens = append(sentence.Tokens, document.Token{
				Text: word, Normal: document.Normalize(word), Tag: "NN", Word: document.IsWord(word),
				Protected: strings.ContainsRune(word, 0), Start: start, End: text.Len(),
				Spans: []document.Span{{Start: start, End: text.Len()}},
			})
			text.WriteByte(' ')
		}
		block.Sentences = append(block.Sentences, sentence)
	}
	block.Text = text.String()
	block.Span = document.Span{End: len(block.Text)}
	for i := range len(block.Text) {
		block.Map = append(block.Map, document.Span{Start: i, End: i + 1})
	}
	return block
}

func numeric(t *testing.T, m feature.Measurements, id string) float64 {
	t.Helper()
	c := qt.New(t)
	value, err := m.Value(id)
	c.Assert(err, qt.IsNil)
	c.Assert(value.Reason, qt.Equals, "")
	c.Assert(value.Number, qt.IsNotNil)
	return *value.Number
}

func TestCountsUnicodeAndProtectedWords(t *testing.T) {
	c := qt.New(t)
	block := testBlock("café café 12 long_identifier .")
	block.Sentences[0].Tokens[2].Tag = "CD"
	block.Sentences[0].Tokens[3].Protected = true
	m, err := feature.Measure(t.Context(), block, testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(m.Counts(), qt.Equals, feature.Counts{Available: true, Words: 3, Characters: 10, Sentences: 1, TokenVisits: 5})
	c.Assert(numeric(t, m, "noun-token-ratio"), qt.Equals, float64(2)/3)
	c.Assert(numeric(t, m, "type-token-ratio"), qt.Equals, float64(2)/3)
	c.Assert(numeric(t, m, "hapax-token-ratio"), qt.Equals, float64(1)/3)
	c.Assert(math.Abs(numeric(t, m, "automated-readability-index")+4.23) < 1e-12, qt.IsTrue)
	c.Assert(m.Spans(), qt.DeepEquals, []document.Span{{Start: 0, End: 5}, {Start: 6, End: 11}, {Start: 12, End: 14}})
}

func TestSentenceStatistics(t *testing.T) {
	c := qt.New(t)
	m, err := feature.Measure(t.Context(), testBlock("a b", "c d e f", "g h i j k l"), testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(m.SentenceLengths(), qt.DeepEquals, []int{2, 4, 6})
	c.Assert(numeric(t, m, "mean-sentence-words"), qt.Equals, float64(4))
	c.Assert(numeric(t, m, "sentence-word-stddev"), qt.Equals, math.Sqrt(float64(8)/3))
	c.Assert(numeric(t, m, "shortest-sentence"), qt.Equals, float64(2))
	c.Assert(numeric(t, m, "longest-sentence"), qt.Equals, float64(6))
}

func TestPOSRatiosUseDeclaredPrefixes(t *testing.T) {
	c := qt.New(t)
	block := testBlock("reliable clients carefully retry now")
	for i, tag := range []string{"JJ", "NNS", "RB", "VBP", "RB"} {
		block.Sentences[0].Tokens[i].Tag = tag
	}
	m, err := feature.Measure(t.Context(), block, testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	for _, id := range []string{"noun-token-ratio", "verb-token-ratio", "adjective-token-ratio"} {
		c.Assert(numeric(t, m, id), qt.Equals, 0.2)
	}
	c.Assert(numeric(t, m, "adverb-token-ratio"), qt.Equals, 0.4)
}

func TestMissingIsNotZero(t *testing.T) {
	c := qt.New(t)
	empty, err := feature.Measure(t.Context(), testBlock(), testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(numeric(t, empty, "prose-words"), qt.Equals, float64(0))
	value, err := empty.Value("type-token-ratio")
	c.Assert(err, qt.IsNil)
	c.Assert(value.Number, qt.IsNil)
	c.Assert(value.Reason, qt.Equals, "no_prose_words")
	data, err := json.Marshal(value)
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Contains, `"number":null`)
	identity := testIdentity()
	identity.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences}
	m, err := feature.Measure(t.Context(), testBlock("a b"), identity, testLimits())
	c.Assert(err, qt.IsNil)
	value, err = m.Value("noun-token-ratio")
	c.Assert(err, qt.IsNil)
	c.Assert(value.Number, qt.IsNil)
	c.Assert(value.Reason, qt.Equals, "capability_missing")
	c.Assert(numeric(t, m, "prose-words"), qt.Equals, float64(2))
	_, err = m.Value("dependency.depth")
	c.Assert(err, qt.ErrorMatches, `unknown feature.*`)
}

func TestSetOwnershipAndConcurrentReads(t *testing.T) {
	c := qt.New(t)
	block := testBlock("one two", "three")
	set, err := feature.NewSet(t.Context(), []document.Block{block}, testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	m, err := set.Block(0)
	c.Assert(err, qt.IsNil)
	before := m.Hash()
	block.Sentences[0].Tokens[0].Spans[0].End = 100
	block.Map[0].End = 100
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			local, getErr := set.Block(0)
			c.Check(getErr, qt.IsNil)
			c.Check(local.Hash(), qt.Equals, before)
			c.Check(local.SentenceLengths(), qt.DeepEquals, []int{2, 1})
			c.Check(local.Spans()[0].End, qt.Equals, 3)
			local.SentenceLengths()[0] = 100
			local.Spans()[0].End = 100
			*local.Values()[0].Number = 100
		})
	}
	workers.Wait()
	c.Assert(numeric(t, m, "prose-words"), qt.Equals, float64(3))
}

func TestIdentityCoversEffectiveInputs(t *testing.T) {
	c := qt.New(t)
	block, identity := testBlock("a b"), testIdentity()
	first, err := feature.Measure(t.Context(), block, identity, testLimits())
	c.Assert(err, qt.IsNil)
	for _, change := range []func(*feature.Identity){
		func(i *feature.Identity) { i.Source = "changed-source" },
		func(i *feature.Identity) { i.Policy = "changed-policy" },
		func(i *feature.Identity) { i.Vocabulary = "changed-terms" },
		func(i *feature.Identity) { i.Preprocessing = "changed-normalizer" },
		func(i *feature.Identity) { i.NLP.ModelHash = "changed-model" },
	} {
		modified := testIdentity()
		change(&modified)
		other, measureErr := feature.Measure(t.Context(), block, modified, testLimits())
		c.Assert(measureErr, qt.IsNil)
		c.Assert(other.Hash(), qt.Not(qt.Equals), first.Hash())
	}
	slices.Reverse(identity.Capabilities)
	other, err := feature.Measure(t.Context(), block, identity, testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(other.Hash(), qt.Equals, first.Hash())
}

func TestCancellation(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := feature.Measure(ctx, testBlock("a"), testIdentity(), testLimits())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = feature.NewSet(ctx, nil, testIdentity(), testLimits())
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
