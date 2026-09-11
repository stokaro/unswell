package llmdet_test

import (
	"bytes"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/llmdet"
)

type tableRow = llmdet.TableRow

// encodeTable writes a table in the binary form the converter produces.
func encodeTable(c *qt.C, rows []tableRow) []byte {
	c.Helper()
	data, err := llmdet.EncodeTable("toy", 16, llmdet.TableSource{File: "toy.npz", SHA256: strings.Repeat("a", 64)}, rows)
	c.Assert(err, qt.IsNil)
	return data
}

// A table measures the same sequences as the row proxy of the parity
// controls: the longest context wins, residual mass covers unlisted
// continuations, and the denominator counts every matched position.
func TestTableMatchesProxyArithmetic(t *testing.T) {
	c := qt.New(t)
	rows := []tableRow{
		{Context: []int{3}, Continuations: []int{4}, Probabilities: []float64{0.125}},
		{Context: []int{2, 3}, Continuations: []int{4}, Probabilities: []float64{0.25}},
		{Context: []int{1, 2, 3}, Continuations: []int{5, 6}, Probabilities: []float64{0.25, 0.25}},
		{Context: []int{4, 5, 6}, Continuations: []int{7}, Probabilities: []float64{0.5}},
	}
	table, err := llmdet.ReadTable(c.Context(), bytes.NewReader(encodeTable(c, rows)))
	c.Assert(err, qt.IsNil)
	c.Assert(table.Header().Model, qt.Equals, "toy")
	c.Assert(table.Header().Orders[2].Contexts, qt.Equals, 2)
	spec := llmdet.ProxySpec{Version: llmdet.ProxyVersion, VocabSize: 16}
	for _, row := range rows {
		spec.Rows = append(spec.Rows, llmdet.ProbabilityRow(row))
	}
	proxy, err := llmdet.NewProxy(c.Context(), spec)
	c.Assert(err, qt.IsNil)
	sequences := [][]int{{1, 2, 3, 4}, {1, 2, 3, 5}, {0, 2, 3, 4}, {0, 0, 3, 4}, {1, 2, 3, 4, 5, 6, 7}, {9, 9, 9, 9}, {1, 2, 3}, {}}
	for _, tokens := range sequences {
		want, err := proxy.Measure(c.Context(), tokens)
		c.Assert(err, qt.IsNil)
		got, err := table.Measure(c.Context(), tokens)
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.DeepEquals, want, qt.Commentf("%v", tokens))
	}
	outside, err := table.Measure(c.Context(), []int{1, 2, 3, 40})
	c.Assert(err, qt.IsNil)
	c.Assert(outside.Matched, qt.Equals, 1)
	c.Assert(outside.Residual, qt.Equals, 1)
}

// A table with unsorted contexts, a value outside the vocabulary, or a
// truncated payload is refused.
func TestTableRefusesMalformedFiles(t *testing.T) {
	c := qt.New(t)
	duplicate := encodeTable(c, []tableRow{{Context: []int{3}, Continuations: []int{4}, Probabilities: []float64{0.125}},
		{Context: []int{3}, Continuations: []int{5}, Probabilities: []float64{0.125}}})
	_, err := llmdet.ReadTable(c.Context(), bytes.NewReader(duplicate))
	c.Assert(err, qt.ErrorMatches, "table order 1: contexts are not sorted and distinct")
	outside := encodeTable(c, []tableRow{{Context: []int{3}, Continuations: []int{4}, Probabilities: []float64{1.5}}})
	_, err = llmdet.ReadTable(c.Context(), bytes.NewReader(outside))
	c.Assert(err, qt.ErrorMatches, "table order 1: probability outside the unit interval")
	valid := encodeTable(c, []tableRow{{Context: []int{3}, Continuations: []int{4}, Probabilities: []float64{0.125}}})
	_, err = llmdet.ReadTable(c.Context(), bytes.NewReader(valid[:len(valid)-2]))
	c.Assert(err, qt.IsNotNil)
	_, err = llmdet.ReadTable(c.Context(), bytes.NewReader(append(append([]byte{}, valid...), 0)))
	c.Assert(err, qt.ErrorMatches, "table has trailing bytes")
	_, err = llmdet.ReadTable(c.Context(), bytes.NewReader([]byte("nope")))
	c.Assert(err, qt.ErrorMatches, "table lacks its magic prefix")
}
