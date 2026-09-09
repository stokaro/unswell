// Command annotate prepares review packets, measures agreement, and exports decisions.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

func main() {
	os.Exit(mainCode())
}

func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout); err != nil {
		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return 130
		}
		_, _ = commandio.Await(ctx, func() (int, error) { return fmt.Fprintln(os.Stderr, err) })
		if ctx.Err() != nil {
			return 130
		}
		return 2
	}
	return 0
}

func run(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: annotate {validate|packet|agreement|decisions} < round.json")
	}
	if !slices.Contains([]string{"validate", "packet", "agreement", "decisions"}, args[0]) {
		return fmt.Errorf("unknown annotation command %q", args[0])
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := commandio.Await(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(input, annotation.MaxBytes+1))
	})
	if err != nil {
		return err
	}
	round, err := annotation.Load(ctx, data)
	if err != nil {
		return err
	}
	result, err := annotationOutput(ctx, round, args[0])
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	_, err = commandio.Await(ctx, func() (struct{}, error) {
		return struct{}{}, encoder.Encode(result)
	})
	return err
}

func annotationOutput(ctx context.Context, round *annotation.Round, command string) (any, error) {
	switch command {
	case "validate":
		return map[string]string{"status": "structurally_valid", "version": annotation.Version}, nil
	case "packet":
		return round.Packet(ctx)
	case "agreement":
		return round.Agreement(ctx)
	case "decisions":
		return round.Decisions(ctx)
	default:
		return nil, fmt.Errorf("unknown annotation command %q", command)
	}
}
