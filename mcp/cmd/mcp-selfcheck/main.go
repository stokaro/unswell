// Command mcp-selfcheck verifies a server subprocess against retained CLI evidence.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stokaro/unswell/mcp/internal/selfcheck"
)

func main() {
	os.Exit(run())
}

func run() int {
	flags := flag.NewFlagSet("mcp-selfcheck", flag.ContinueOnError)
	expected := flags.String("expected", "", "complete passing CLI JSON report with included sources")
	output := flags.String("output", "", "destination for verified structured MCP evidence")
	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	count, err := selfcheck.Run(ctx, *expected, *output, flags.Args(), os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if _, err := fmt.Fprintf(os.Stdout,
		"MCP self-check matched CLI evidence for %d documents; failure, rewrite, and malformed-input probes passed.\n", count); err != nil {
		return 1
	}
	return 0
}
