// Command annotate validates rounds, prepares blinded packets, and measures agreement.
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
		_, _ = awaitIO(ctx, func() (int, error) { return fmt.Fprintln(os.Stderr, err) })
		if ctx.Err() != nil {
			return 130
		}
		return 2
	}
	return 0
}

func run(ctx context.Context, args []string, input io.Reader, output io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: annotate {validate|packet|agreement} < round.json")
	}
	if !slices.Contains([]string{"validate", "packet", "agreement"}, args[0]) {
		return fmt.Errorf("unknown annotation command %q", args[0])
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := awaitIO(ctx, func() ([]byte, error) {
		return io.ReadAll(io.LimitReader(input, annotation.MaxBytes+1))
	})
	if err != nil {
		return err
	}
	round, err := annotation.Load(ctx, data)
	if err != nil {
		return err
	}
	var result any
	switch args[0] {
	case "validate":
		result = map[string]string{"status": "structurally_valid", "version": annotation.Version}
	case "packet":
		result, err = round.Packet(ctx)
	case "agreement":
		result, err = round.Agreement(ctx)
	default:
		return fmt.Errorf("unknown annotation command %q", args[0])
	}
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	_, err = awaitIO(ctx, func() (struct{}, error) {
		return struct{}{}, encoder.Encode(result)
	})
	return err
}
