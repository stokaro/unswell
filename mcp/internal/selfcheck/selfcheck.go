// Package selfcheck verifies a real MCP subprocess against saved CLI evidence.
package selfcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/mcp/internal/server"
	"github.com/stokaro/unswell/report"
)

// Run starts an explicitly supplied command and verifies its MCP tool results.
// Source contents are sent as data and never become process arguments or shell source.
func Run(ctx context.Context, expectedPath, outputPath string, argv []string, stderr io.Writer) (count int, returnedErr error) {
	if len(argv) == 0 {
		return 0, fmt.Errorf("an MCP server command after -- is required")
	}
	expected, err := readExpected(expectedPath)
	if err != nil {
		return 0, err
	}
	// #nosec G204 -- The operator selects this CI command; checked source is sent only through MCP data.
	command := exec.CommandContext(ctx, argv[0], argv[1:]...)
	command.Stderr = stderr
	client := mcp.NewClient(&mcp.Implementation{Name: "unswell-selfcheck", Version: unswell.Version}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: command}, nil)
	if err != nil {
		return 0, err
	}
	defer func() { returnedErr = errors.Join(returnedErr, session.Close()) }()
	evidence, err := verifySession(ctx, session, expected)
	if err != nil {
		return 0, err
	}
	if err := session.Close(); err != nil {
		return 0, err
	}
	if err := writeEvidence(outputPath, evidence); err != nil {
		return 0, err
	}
	return len(expected.Documents), nil
}

func verifySession(ctx context.Context, session *mcp.ClientSession, expected unswell.RunResult) (evidence, error) {
	if err := verifyDiscovery(ctx, session, expected); err != nil {
		return evidence{}, err
	}
	input := server.CheckInput{Sources: make([]server.Source, 0, len(expected.Documents))}
	for _, doc := range expected.Documents {
		input.Sources = append(input.Sources, server.Source{Name: doc.Name, Format: doc.Format, Text: doc.Source})
	}
	checked, err := check(ctx, session, input, "pass")
	if err != nil {
		return evidence{}, err
	}
	if !reflect.DeepEqual(normalize(expected), checked.Result) {
		return evidence{}, fmt.Errorf("MCP result differs from normalized CLI evidence")
	}
	probes, err := verifyProbes(ctx, session)
	if err != nil {
		return evidence{}, err
	}
	return evidence{Repository: checked, Probes: probes}, nil
}

func readExpected(path string) (unswell.RunResult, error) {
	// #nosec G304 -- Read only the explicitly selected, bounded CLI evidence file.
	file, err := os.Open(path)
	if err != nil {
		return unswell.RunResult{}, err
	}
	result, readErr := report.Read(file)
	if err := errors.Join(readErr, file.Close()); err != nil {
		return unswell.RunResult{}, err
	}
	if !result.Manifest.IncludeSource || result.Status != "complete" || !result.Gate.Passed || len(result.Documents) == 0 {
		return unswell.RunResult{}, fmt.Errorf("expected evidence must be a complete passing CLI scan with included sources")
	}
	return result, nil
}

func verifyDiscovery(ctx context.Context, session *mcp.ClientSession, expected unswell.RunResult) error {
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		return err
	}
	if len(tools.Tools) != 2 {
		return fmt.Errorf("expected both Unswell MCP tools")
	}
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "unswell_describe", Arguments: server.DescribeInput{}})
	if err != nil {
		return err
	}
	if result.IsError {
		return fmt.Errorf("MCP policy discovery failed")
	}
	var description server.Description
	if err := decodeResult(result, &description); err != nil {
		return err
	}
	if description.Policy.Hash != expected.Manifest.ConfigHash || description.Commit != expected.Manifest.ToolCommit {
		return fmt.Errorf("MCP policy or build revision differs from the checked CLI")
	}
	return nil
}

func check(ctx context.Context, session *mcp.ClientSession, input server.CheckInput, outcome string) (server.CheckOutput, error) {
	result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "unswell_check", Arguments: input})
	if err != nil {
		return server.CheckOutput{}, err
	}
	var checked server.CheckOutput
	if err := decodeResult(result, &checked); err != nil {
		return checked, err
	}
	if checked.Outcome != outcome || result.IsError != (outcome == "error") {
		return checked, fmt.Errorf("MCP probe returned %q (isError=%t); expected %q", checked.Outcome, result.IsError, outcome)
	}
	return checked, nil
}

func decodeResult(result *mcp.CallToolResult, value any) error {
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		return err
	}
	if result.StructuredContent == nil {
		return fmt.Errorf("MCP result has no structured evidence")
	}
	return json.Unmarshal(data, value)
}

func normalize(result unswell.RunResult) unswell.RunResult {
	result.Manifest.SelectionMode = "explicit"
	result.Manifest.IncludeSource = false
	for i := range result.Documents {
		result.Documents[i].Source = ""
	}
	for i := range result.Findings {
		result.Findings[i].Primary.Snippet = ""
		for j := range result.Findings[i].Related {
			result.Findings[i].Related[j].Snippet = ""
		}
	}
	return result
}

type evidence struct {
	Repository server.CheckOutput   `json:"repository"`
	Probes     []server.CheckOutput `json:"probes"`
}

func writeEvidence(path string, result evidence) error {
	// #nosec G304 -- The operator selects the CI evidence destination, outside the checked source tree.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	writeErr := json.NewEncoder(file).Encode(result)
	return errors.Join(writeErr, file.Close())
}
