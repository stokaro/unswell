// Command reviewclaims prepares and replays local source-bound editorial claims.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/stokaro/unswell/research/annotation/internal/claimreview"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
)

func main() { os.Exit(mainCode()) }

func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := claimreview.Run(ctx, os.Stdin, os.Stdout); err != nil {
		if ctx.Err() != nil {
			return 130
		}
		_, _ = commandio.Await(ctx, func() (int, error) { return fmt.Fprintln(os.Stderr, err) })
		return 2
	}
	return 0
}
