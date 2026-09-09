// Command llmdetprobe evaluates local numerical controls for LLMDet research.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/internal/llmdetcommand"
)

func main() { os.Exit(mainCode()) }

func mainCode() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := llmdetcommand.Run(ctx, os.Args[1:], os.Stdin, os.Stdout); err != nil {
		if ctx.Err() != nil {
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
