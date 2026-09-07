// Package command owns MCP startup flags, explicit configuration files, and stderr diagnostics.
package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/mcp/internal/server"
)

// Run starts the server on a supplied transport. Only protocol messages use that transport.
func Run(ctx context.Context, args []string, stderr io.Writer, transport mcp.Transport) error {
	flags := flag.NewFlagSet("unswell-mcp", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "explicit Unswell policy file; omitted uses builtin defaults")
	timeout := flags.Duration("timeout", 30*time.Second, "maximum duration of one check (at most 5m)")
	version := flags.Bool("version", false, "print version to stderr and exit")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if *version {
		_, err := fmt.Fprintf(stderr, "unswell-mcp %s (%s)\n", unswell.Version, unswell.BuildCommit)
		return err
	}
	data, err := readPolicy(*configPath)
	if err != nil {
		return err
	}
	instance, err := server.New(server.Options{Config: data, Timeout: *timeout})
	if err != nil {
		return err
	}
	return instance.Run(ctx, transport)
}

func readPolicy(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	// #nosec G304 -- Only the operator's explicit startup policy path is read, under a byte limit.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err := errors.Join(readErr, file.Close()); err != nil {
		return nil, err
	}
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("configuration exceeds 1 MiB")
	}
	return data, nil
}
