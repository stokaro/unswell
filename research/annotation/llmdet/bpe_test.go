package llmdet_test

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/llmdet"
)

func readGzip(c *qt.C, path string) []byte {
	c.Helper()
	file, err := os.Open(path) // #nosec G304 -- The test reads its own fixture.
	c.Assert(err, qt.IsNil)
	reader, err := gzip.NewReader(file)
	c.Assert(err, qt.IsNil)
	data, err := io.ReadAll(reader)
	c.Assert(err, qt.IsNil)
	c.Assert(file.Close(), qt.IsNil)
	return data
}

// The tokenizer reproduces the published GPT-2 encodings of the reference
// cases, which tiktoken produced from the same vocabulary and merges.
func TestBPEMatchesReferenceEncodings(t *testing.T) {
	c := qt.New(t)
	tokenizer, err := llmdet.LoadBPE(c.Context(), readGzip(c, "testdata/gpt2/vocab.json.gz"), readGzip(c, "testdata/gpt2/merges.txt.gz"))
	c.Assert(err, qt.IsNil)
	data, err := os.ReadFile("testdata/gpt2/reference.json")
	c.Assert(err, qt.IsNil)
	var cases []struct {
		Text   string `json:"text"`
		Tokens []int  `json:"tokens"`
	}
	c.Assert(json.Unmarshal(data, &cases), qt.IsNil)
	c.Assert(len(cases) > 10, qt.IsTrue)
	for _, row := range cases {
		got, err := tokenizer.Encode(c.Context(), row.Text)
		c.Assert(err, qt.IsNil, qt.Commentf("%q", row.Text))
		if len(row.Tokens) == 0 {
			c.Assert(got, qt.HasLen, 0, qt.Commentf("%q", row.Text))
			continue
		}
		c.Assert(got, qt.DeepEquals, row.Tokens, qt.Commentf("%q", row.Text))
	}
	again, err := tokenizer.Encode(c.Context(), cases[0].Text)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, cases[0].Tokens)
}

// A vocabulary that cannot represent a symbol, a merge outside the
// vocabulary, and oversized inputs are refused.
func TestBPERefusesIncompleteFiles(t *testing.T) {
	c := qt.New(t)
	vocabulary := []byte(`{"a": 0, "b": 1, "ab": 2, "Ġ": 3}`)
	tokenizer, err := llmdet.LoadBPE(c.Context(), vocabulary, []byte("#version: 0.2\na b\n"))
	c.Assert(err, qt.IsNil)
	tokens, err := tokenizer.Encode(c.Context(), "ab a")
	c.Assert(err, qt.IsNil)
	c.Assert(tokens, qt.DeepEquals, []int{2, 3, 0})
	_, err = tokenizer.Encode(c.Context(), "abc")
	c.Assert(err, qt.ErrorMatches, `tokenizer vocabulary lacks "c"`)
	_, err = llmdet.LoadBPE(c.Context(), vocabulary, []byte("a c\n"))
	c.Assert(err, qt.ErrorMatches, `merge "a c" produces a token outside the vocabulary`)
	_, err = llmdet.LoadBPE(c.Context(), vocabulary, []byte("a b\na b\n"))
	c.Assert(err, qt.ErrorMatches, `merge "a b" is repeated`)
	_, err = llmdet.LoadBPE(c.Context(), []byte(`{"a": 0, "b": 0}`), []byte("a b\n"))
	c.Assert(err, qt.ErrorMatches, "tokenizer vocabulary has an invalid or repeated entry")
	_, err = llmdet.LoadBPE(c.Context(), vocabulary, []byte(""))
	c.Assert(err, qt.ErrorMatches, "tokenizer requires at least one merge")
	_, err = tokenizer.Encode(c.Context(), string([]byte{0xff}))
	c.Assert(err, qt.ErrorMatches, "tokenizer text requires valid UTF-8 within .*")
	var missing *llmdet.BPE
	_, err = missing.Encode(c.Context(), "a")
	c.Assert(err, qt.ErrorMatches, "missing tokenizer")
}
