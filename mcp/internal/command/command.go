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
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/internal/appconfig"
	"github.com/stokaro/unswell/mcp/internal/server"
)

// Run starts the server on a supplied transport. Only protocol messages use that transport.
func Run(ctx context.Context, args []string, stderr io.Writer, transport mcp.Transport) error {
	flags := flag.NewFlagSet("unswell-mcp", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var features, preparedFeatures, preparedKinds []string
	flags.Func("prepared-feature", "collect a prepared target feature; repeat for a set", appendFlag(&preparedFeatures))
	flags.Func("prepared-kind", "select sentence, paragraph, or fragment; repeat for a set", appendFlag(&preparedKinds))
	flags.Func("feature", "collect a shared block feature by ID; repeat to select a set", func(id string) error {
		features = append(features, id)
		return nil
	})
	configPath := flags.String("config", "", "explicit Unswell policy file; omitted uses builtin defaults")
	baselinePath := flags.String("baseline", "", "explicit local baseline loaded once at startup; tools cannot update it")
	gateMode := flags.String("gate-mode", "", "override gate mode: all or new (new requires a baseline)")
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
	accepted, err := readBaseline(*baselinePath)
	if err != nil {
		return err
	}
	instance, err := server.New(server.Options{Features: features,
		PreparedFeatures: preparedFeatures, PreparedKinds: preparedKinds, ConfigBundle: &loaded.Bundle,
		Timeout: *timeout, Baseline: accepted, GateMode: *gateMode})
	if err != nil {
		return err
	}
	return instance.Run(ctx, transport)
}

func readBaseline(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	// #nosec G304 -- The operator explicitly selects a local artifact; reads have a byte limit.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, baseline.MaxBytes+1))
	if err := errors.Join(readErr, file.Close()); err != nil {
		return nil, err
	}
	if len(data) > baseline.MaxBytes {
		return nil, fmt.Errorf("baseline exceeds %d bytes", baseline.MaxBytes)
	}
	return data, nil
}

func appendFlag(values *[]string) func(string) error {
	return func(value string) error {
		*values = append(*values, value)
		return nil
	}
}
