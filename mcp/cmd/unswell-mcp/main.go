// Command unswell-mcp serves offline prose checks over MCP stdio.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell/mcp/internal/command"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Reserve the original stdout for protocol frames. Dependency diagnostics
	// that use the Go stdout variable go to stderr before any source is parsed.
	protocolOutput := os.Stdout
	os.Stdout = os.Stderr
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	transport := &mcp.IOTransport{Reader: os.Stdin, Writer: protocolOutput}
	if err := command.Run(ctx, os.Args[1:], os.Stderr, transport); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, context.Canceled) {
			return 130
		}
		return 2
	}
	return 0
}
