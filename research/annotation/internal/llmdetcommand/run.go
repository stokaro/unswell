// Package llmdetcommand runs local numerical probes without a text-scanning policy.
package llmdetcommand

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
	"github.com/stokaro/unswell/research/annotation/llmdet"
)

const maxInputBytes = 4 << 20

type result struct {
	Version     string `json:"version"`
	Kind        string `json:"kind"`
	ModelSHA256 string `json:"model_sha256"`
	Results     any    `json:"results"`
}

// Run reads an explicit local numerical pack and a bounded batch from input.
// It does not download resources, tokenize prose, or apply a CI gate.
func Run(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	kind, path, err := parseOptions(args)
	if err != nil {
		return err
	}
	maximum := llmdet.MaxProxyBytes
	if kind == "ensemble" {
		maximum = llmdet.MaxEnsembleBytes
	}
	pack, err := readPack(ctx, path, maximum)
	if err != nil {
		return err
	}
	data, err := readBounded(ctx, input, maxInputBytes)
	if err != nil {
		return err
	}
	values, err := calculate(ctx, kind, pack, data)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(pack)
	report := result{Version: "unswell-llmdet-probe-v1", Kind: kind, ModelSHA256: hex.EncodeToString(hash[:]), Results: values}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	count, err := commandio.Await(ctx, func() (int, error) { return output.Write(encoded) })
	if err == nil && count != len(encoded) {
		return io.ErrShortWrite
	}
	return err
}

func parseOptions(args []string) (string, string, error) {
	flags := flag.NewFlagSet("llmdetprobe", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	kind := flags.String("kind", "", "proxy or ensemble")
	path := flags.String("model", "", "local JSON numerical pack")
	if err := flags.Parse(args); err != nil {
		return "", "", err
	}
	if (*kind != "proxy" && *kind != "ensemble") || *path == "" || flags.NArg() != 0 {
		return "", "", fmt.Errorf("usage: llmdetprobe --kind {proxy|ensemble} --model local.json < vectors.json")
	}
	return *kind, *path, nil
}

func readPack(ctx context.Context, path string, maximum int) ([]byte, error) {
	return commandio.Await(ctx, func() ([]byte, error) {
		file, err := os.Open(path) // #nosec G304 -- The developer explicitly selects the local numerical pack.
		if err != nil {
			return nil, err
		}
		data, readErr := readBounded(ctx, file, maximum)
		return data, errors.Join(readErr, file.Close())
	})
}

func readBounded(ctx context.Context, input io.Reader, maximum int) ([]byte, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(input, int64(maximum)+1))
	})
	if err != nil {
		return nil, err
	}
	if len(data) > maximum {
		return nil, fmt.Errorf("probe input exceeds %d bytes", maximum)
	}
	return data, nil
}

func calculate(ctx context.Context, kind string, pack, data []byte) (any, error) {
	if kind == "proxy" {
		proxy, err := llmdet.LoadProxy(ctx, pack)
		if err != nil {
			return nil, err
		}
		return batch(ctx, data, llmdet.MaxTokens, "integer", proxy.Measure)
	}
	ensemble, err := llmdet.LoadEnsemble(ctx, pack)
	if err != nil {
		return nil, err
	}
	return batch(ctx, data, 128, "number", ensemble.Predict)
}

func batch[T int | float64, R any](ctx context.Context, data []byte, width int, numberType string,
	run func(context.Context, []T) (R, error),
) ([]R, error) {
	var vectors [][]T
	if err := jsoninput.Decode(ctx, data, maxInputBytes, &vectors, jsoninput.Limits{Array: max(width, 1000), Object: 1}); err != nil {
		return nil, err
	}
	schema := fmt.Sprintf(`{"type":"array","minItems":1,"maxItems":1000,"items":{"type":"array","maxItems":%d,"items":{"type":%q}}}`,
		width, numberType)
	if err := jsoninput.Schema(data, []byte(schema), "urn:unswell:llmdet:vectors:v1"); err != nil {
		return nil, err
	}
	if len(vectors) == 0 || len(vectors) > 1000 {
		return nil, fmt.Errorf("probe requires 1 to 1000 vectors")
	}
	results := make([]R, 0, len(vectors))
	for _, vector := range vectors {
		if vector == nil || len(vector) > width {
			return nil, fmt.Errorf("invalid or excessive vector")
		}
		value, err := run(ctx, vector)
		if err != nil {
			return nil, err
		}
		results = append(results, value)
	}
	return results, nil
}
