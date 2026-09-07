// Package command owns MCP startup flags, explicit configuration files, and stderr diagnostics.
package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/internal/appconfig"
	"github.com/stokaro/unswell/mcp/internal/server"
)

// Run starts the server on a supplied transport. Only protocol messages use that transport.
func Run(ctx context.Context, args []string, stderr io.Writer, transport mcp.Transport) error {
	flags := flag.NewFlagSet("unswell-mcp", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "explicit Unswell policy file; omitted uses builtin defaults")
	projectRoot := flags.String("project-root", "", "root for local policy dependencies and logical source names")
	allowOutside := flags.Bool("allow-config-outside-root", false, "explicitly permit local configuration outside the project root")
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
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	loaded, err := appconfig.Load(ctx, appconfig.Options{Dir: dir, Root: *projectRoot, Path: *configPath,
		AllowOutsideRoot: *allowOutside})
	if err != nil {
		return err
	}
	instance, err := server.New(server.Options{ConfigBundle: &loaded.Bundle, Timeout: *timeout})
	if err != nil {
		return err
	}
	return instance.Run(ctx, transport)
}
