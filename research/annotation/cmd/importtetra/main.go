// Command importtetra imports explicit local TETRA XML into grouped research pairs.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/internal/tetra"
)

func main() { os.Exit(mainCode()) }

func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := tetra.Run(ctx, os.Stdin, os.Stdout); err != nil {
		if ctx.Err() != nil {
			return 130
		}
		_, _ = commandio.Await(ctx, func() (int, error) { return fmt.Fprintln(os.Stderr, err) })
		return 2
	}
	return 0
}
