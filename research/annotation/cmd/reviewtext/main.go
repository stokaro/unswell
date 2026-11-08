// Command reviewtext exports verified eligible prose for isolated research.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/internal/reviewbaseline"
)

func main() { os.Exit(mainCode()) }

func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	var err error
	switch {
	case len(os.Args) == 1:
		err = reviewbaseline.ExportText(ctx, os.Stdin, os.Stdout)
	case len(os.Args) == 2 && os.Args[1] == "--source-inventory":
		err = reviewbaseline.ExportSourceInventory(ctx, os.Stdin, os.Stdout)
	default:
		err = fmt.Errorf("usage: reviewtext [--source-inventory]")
	}
	if err != nil {
		if ctx.Err() != nil {
			return 130
		}
		_, _ = commandio.Await(ctx, func() (int, error) { return fmt.Fprintln(os.Stderr, err) })
		return 2
	}
	return 0
}
