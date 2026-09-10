package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation/evaluation"
	"github.com/stokaro/unswell/research/annotation/figures"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

// maxFigureBytes bounds one rendered chart. A figure describes a bounded record,
// so a large one means the renderer is wrong rather than the data rich.
const maxFigureBytes = 1 << 20

func figureFlags(args []string) (string, error) {
	var plot string
	flags := flag.NewFlagSet("corpus figures", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&plot, "plot", "", "Chart name: "+strings.Join(figures.Names(), " or "))
	if err := flags.Parse(args[1:]); err != nil {
		return "", err
	}
	if flags.NArg() != 0 || !slices.Contains(figures.Names(), plot) {
		return "", fmt.Errorf("figures requires --plot %s and no positional arguments",
			strings.Join(figures.Names(), " or "))
	}
	return plot, nil
}

// runFigures renders one chart from a saved evaluation record. It computes
// nothing: a figure that disagreed with the record beside it would be worse than
// no figure at all.
func runFigures(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	plot, err := figureFlags(args)
	if err != nil {
		return err
	}
	data, err := readInput(ctx, "figures", input)
	if err != nil {
		return err
	}
	result, err := loadEvaluation(data)
	if err != nil {
		return err
	}
	svg, err := figures.Render(plot, result.Summary)
	if err != nil {
		return err
	}
	if len(svg) > maxFigureBytes {
		return fmt.Errorf("rendered figure exceeds its byte limit")
	}
	count, err := commandio.Await(ctx, func() (int, error) { return output.Write(svg) })
	if err == nil && count != len(svg) {
		return io.ErrShortWrite
	}
	return err
}

func loadEvaluation(data []byte) (evaluation.Result, error) {
	var result evaluation.Result
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return evaluation.Result{}, fmt.Errorf("evaluation record: %w", err)
	}
	if result.Version != evaluation.Version {
		return evaluation.Result{}, fmt.Errorf("unsupported evaluation record %q", result.Version)
	}
	return result, nil
}
