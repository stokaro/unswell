package llmdet

import (
	"context"
	_ "embed"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

//go:embed proxy.schema.json
var proxySchema []byte

//go:embed ensemble.schema.json
var ensembleSchema []byte

const (
	// MaxProxyBytes bounds a JSON probability pack before typed allocation.
	MaxProxyBytes = 32 << 20
	// MaxEnsembleBytes bounds a JSON ensemble before typed allocation.
	MaxEnsembleBytes = 8 << 20
)

// LoadProxy rejects malformed or ambiguous JSON before constructing a proxy.
func LoadProxy(ctx context.Context, data []byte) (*Proxy, error) {
	var spec ProxySpec
	limits := jsoninput.Limits{Object: 3, Array: maxTopK, Arrays: map[string]int{"rows": maxRows, "context": 3}}
	if err := jsoninput.Decode(ctx, data, MaxProxyBytes, &spec, limits); err != nil {
		return nil, err
	}
	if err := jsoninput.Schema(data, proxySchema, "urn:unswell:llmdet:proxy:v1"); err != nil {
		return nil, err
	}
	return NewProxy(ctx, spec)
}

// LoadEnsemble rejects malformed or ambiguous JSON before constructing trees.
func LoadEnsemble(ctx context.Context, data []byte) (*Ensemble, error) {
	var spec EnsembleSpec
	limits := jsoninput.Limits{Object: 4, Array: maxLeaves, Arrays: map[string]int{"trees": maxTrees, "classes": maxClasses}}
	if err := jsoninput.Decode(ctx, data, MaxEnsembleBytes, &spec, limits); err != nil {
		return nil, err
	}
	if err := jsoninput.Schema(data, ensembleSchema, "urn:unswell:llmdet:ensemble:v1"); err != nil {
		return nil, err
	}
	return NewEnsemble(ctx, spec)
}
