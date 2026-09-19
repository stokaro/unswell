// Command encodedreview compares frozen encoder columns on exposed review labels.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/internal/reviewbaseline"
)

func main() { os.Exit(mainCode()) }
func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	err := run(ctx)
	if err == nil {
		return 0
	}
	if ctx.Err() != nil {
		return 130
	}
	_, _ = commandio.Await(ctx, func() (int, error) { return fmt.Fprintln(os.Stderr, err) })
	return 2
}
func run(ctx context.Context) error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: encodedreview EMBEDDINGS_JSON < VERIFIED_INPUT_JSON")
	}
	root, err := os.OpenRoot(filepath.Dir(os.Args[1]))
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	f, err := root.Open(filepath.Base(os.Args[1]))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return reviewbaseline.RunEncoded(ctx, os.Stdin, f, os.Stdout)
}
