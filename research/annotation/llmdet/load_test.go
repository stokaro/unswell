package llmdet_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/llmdet"
)

const emptyProxyJSON = `{"version":"unswell-llmdet-proxy-v1","vocab_size":16,"rows":[]}`

func TestLoadRejectsAmbiguousAndIncompletePacks(t *testing.T) {
	c := qt.New(t)
	ensembleJSON, err := json.Marshal(smallEnsemble())
	c.Assert(err, qt.IsNil)
	for _, pack := range []struct {
		name string
		data string
		load func(context.Context, []byte) error
	}{
		{"proxy", emptyProxyJSON, loadProxy},
		{"ensemble", string(ensembleJSON), loadEnsemble},
	} {
		t.Run(pack.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(pack.load(t.Context(), []byte(pack.data)), qt.IsNil)
			for _, test := range []struct{ name, data string }{
				{"empty", ""}, {"null", "null"}, {"object", "{}"}, {"array", "[]"},
				{"trailing", pack.data + " {}"},
				{"unknown", strings.Replace(pack.data, "{", `{"unknown":0,`, 1)},
				{"duplicate", strings.Replace(pack.data, "{", `{"version":"other",`, 1)},
				{"case alias", strings.Replace(pack.data, `"version"`, `"Version"`, 1)},
				{"null version", strings.Replace(pack.data, `"version":"unswell-llmdet-`+pack.name+`-v1"`, `"version":null`, 1)},
				{"invalid UTF-8", strings.Replace(pack.data, "unswell", "\xff", 1)},
				{"surrogate", strings.Replace(pack.data, "unswell", `\ud800`, 1)},
			} {
				t.Run(test.name, func(t *testing.T) {
					c := qt.New(t)
					c.Assert(pack.load(t.Context(), []byte(test.data)), qt.IsNotNil)
				})
			}
		})
	}
}

func loadProxy(ctx context.Context, data []byte) error {
	_, err := llmdet.LoadProxy(ctx, data)
	return err
}

func loadEnsemble(ctx context.Context, data []byte) error {
	_, err := llmdet.LoadEnsemble(ctx, data)
	return err
}

func TestLoadRejectsNullNumericFields(t *testing.T) {
	c := qt.New(t)
	data, err := json.Marshal(smallEnsemble())
	c.Assert(err, qt.IsNil)
	for _, field := range []string{`"class":0`, `"feature":0`, `"threshold":1`, `"left":-1`, `"right":-2`} {
		name, _, _ := strings.Cut(field, ":")
		bad := strings.Replace(string(data), field, name+":null", 1)
		c.Assert(bad, qt.Not(qt.Equals), string(data))
		_, err := llmdet.LoadEnsemble(t.Context(), []byte(bad))
		c.Assert(err, qt.IsNotNil)
	}
	for _, entries := range []string{
		`{"context":[null],"continuations":[],"probabilities":[]}`,
		`{"context":[1],"continuations":[null],"probabilities":[0]}`,
		`{"context":[1],"continuations":[2],"probabilities":[null]}`,
		`{"context":[1],"continuations":[],"probabilities":null}`,
	} {
		bad := strings.Replace(emptyProxyJSON, `"rows":[]`, `"rows":[`+entries+`]`, 1)
		_, err := llmdet.LoadProxy(t.Context(), []byte(bad))
		c.Assert(err, qt.IsNotNil)
	}
}

func TestLoadBoundsAndCancellation(t *testing.T) {
	c := qt.New(t)
	_, err := llmdet.LoadProxy(t.Context(), bytes.Repeat([]byte{' '}, llmdet.MaxProxyBytes+1))
	c.Assert(err, qt.IsNotNil)
	_, err = llmdet.LoadEnsemble(t.Context(), bytes.Repeat([]byte{' '}, llmdet.MaxEnsembleBytes+1))
	c.Assert(err, qt.IsNotNil)
	bad := strings.Replace(emptyProxyJSON, "[]", "["+strings.Repeat("{},", 100_000)+"{}]", 1)
	_, err = llmdet.LoadProxy(t.Context(), []byte(bad))
	c.Assert(err, qt.ErrorMatches, `.*exceeds 100000 entries`)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = llmdet.LoadProxy(ctx, []byte(emptyProxyJSON))
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = llmdet.NewEnsemble(ctx, smallEnsemble())
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestZeroModelsReturnErrors(t *testing.T) {
	c := qt.New(t)
	for _, proxy := range []*llmdet.Proxy{nil, {}} {
		_, err := proxy.Measure(t.Context(), nil)
		c.Assert(err, qt.IsNotNil)
	}
	for _, model := range []*llmdet.Ensemble{nil, {}} {
		_, err := model.Predict(t.Context(), nil)
		c.Assert(err, qt.IsNotNil)
	}
}

func TestModelsSupportConcurrentIndependentCalls(t *testing.T) {
	c := qt.New(t)
	model, err := llmdet.NewEnsemble(t.Context(), smallEnsemble())
	c.Assert(err, qt.IsNil)
	proxy, err := llmdet.NewProxy(t.Context(), llmdet.ProxySpec{Version: llmdet.ProxyVersion, VocabSize: 16,
		Rows: []llmdet.ProbabilityRow{{Context: []int{3}, Continuations: []int{4}, Probabilities: []float64{0.5}}},
	})
	c.Assert(err, qt.IsNil)
	for range 16 {
		t.Run("independent", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			for range 10 {
				result, err := model.Predict(t.Context(), []float64{1})
				c.Assert(err, qt.IsNil)
				c.Assert(result.Raw, qt.DeepEquals, []float64{2, 0})
				result.Raw[0] = math.NaN()
				result.Classes[0] = "changed"
				value, err := proxy.Measure(t.Context(), []int{1, 2, 3, 4})
				c.Assert(err, qt.IsNil)
				c.Assert(value.Reason, qt.Equals, "")
				c.Assert(value.Value, qt.IsNotNil)
				c.Assert(*value.Value, qt.Equals, 0.5)
				*value.Value = 7
			}
		})
	}
}

func FuzzLoadNumericalPack(f *testing.F) {
	f.Add([]byte(emptyProxyJSON))
	f.Add([]byte(`{"version":"unswell-llmdet-ensemble-v1","feature_count":1,"classes":["a","b"],"trees":[]}`))
	f.Add([]byte(`{"version":null}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		c := qt.New(t)
		if len(data) > 65536 {
			return
		}
		proxy, err := llmdet.LoadProxy(t.Context(), data)
		if err == nil {
			_, err = proxy.Measure(t.Context(), nil)
			c.Assert(err, qt.IsNil)
		}
		_, _ = llmdet.LoadEnsemble(t.Context(), data)
	})
}
