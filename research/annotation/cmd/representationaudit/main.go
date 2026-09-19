// Command representationaudit exports mapped blocks and prepared model units.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/stokaro/unswell/research/annotation/internal/representation"
)

func main() { os.Exit(mainCode()) }

func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := representation.Run(ctx, os.Stdin, os.Stdout); err != nil {
		if ctx.Err() != nil {
			return 130
		}
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 2
	}
	return 0
}
