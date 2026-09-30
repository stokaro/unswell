package tetra

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

//go:embed input.schema.json
var inputSchema []byte

// Run imports one explicit JSON inventory without reading files or using a network.
// It writes a complete quarantine record before returning ErrIncomplete.
func Run(ctx context.Context, reader io.Reader, writer io.Writer) error {
	input, err := readInput(ctx, reader)
	if err != nil {
		return err
	}
	result, importError := Import(ctx, input)
	if importError != nil && !errors.Is(importError, ErrIncomplete) {
		return importError
	}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if len(data) > maximumOutputBytes {
		return fmt.Errorf("TETRA output byte limit exceeded")
	}
	data = append(data, '\n')
	_, err = commandio.Await(ctx, func() (int, error) {
		n, err := writer.Write(data)
		if err == nil && n != len(data) {
			err = io.ErrShortWrite
		}
		return n, err
	})
	if err != nil {
		return err
	}
	return importError
}

func readInput(ctx context.Context, reader io.Reader) (Input, error) {
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(reader, maximumInputBytes+1))
	})
	if err != nil {
		return Input{}, err
	}
	var input Input
	if err := jsoninput.Decode(ctx, data, maximumInputBytes, &input, jsoninput.Limits{Array: maximumFiles, Object: 16}); err != nil {
		return Input{}, err
	}
	if err := jsoninput.Schema(data, inputSchema, "urn:unswell:tetra-import:v1"); err != nil {
		return Input{}, err
	}
	return input, nil
}
