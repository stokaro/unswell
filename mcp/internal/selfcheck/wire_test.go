package selfcheck_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/selfcheck"
	"github.com/stokaro/unswell/mcp/internal/server"
	"github.com/stokaro/unswell/report"
)

func TestRunSplitsLargeRepliesThroughARealMCPProcess(t *testing.T) {
	c := qt.New(t)
	features := []string{"prose-words", "type-token-ratio", "activation/readability.long-paragraph",
		"activation/filler.announced-importance", "activation/syntax.noun-stack", "activation/filler.section-announcement",
		"activation/filler.stacked-hedging", "activation/syntax.not-only-density", "activation/syntax.paired-contrast-density",
		"activation/syntax.triad-density", "activation/syntax.whether-preface-density", "activation/syntax.rhetorical-question-density",
		"activation/syntax.passive-candidate-density", "activation/repetition.exact-sentence", "activation/repetition.sentence-openers",
		"activation/repetition.paragraph-openers"}
	expected := wireEvidence(t, features)
	var serialized bytes.Buffer
	c.Assert(report.Write(&serialized, "json", expected, report.Options{}), qt.IsNil)
	directory := t.TempDir()
	input, output := filepath.Join(directory, "expected.json"), filepath.Join(directory, "checked.json")
	c.Assert(os.WriteFile(input, serialized.Bytes(), 0o600), qt.IsNil)
	args := []string{wireServer(t), "--prepared-feature", "prose-words", "--prepared-kind", "sentence", "--prepared-kind", "paragraph"}
	for _, feature := range features {
		args = append(args, "--feature", feature)
	}
	var diagnostics bytes.Buffer
	count, err := selfcheck.Run(t.Context(), input, output, args, &diagnostics)
	c.Assert(err, qt.IsNil, qt.Commentf("%s", diagnostics.String()))
	c.Assert(count, qt.Equals, len(expected.Documents))
	checkWireEvidence(t, output, expected)
}

func wireEvidence(t *testing.T, features []string) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{IncludeSource: true, Features: features,
		PreparedFeatures: []string{"prose-words"}, PreparedKinds: []string{"sentence", "paragraph"}})
	c.Assert(err, qt.IsNil)
	const text = "The client opens connections.\n\nThe server accepts requests.\n\n" +
		"An idle connection closes after five minutes.\n\nTimeouts cancel pending operations.\n\n" +
		"Requests carry unique identifiers.\n\nErrors include the original operation.\n\n" +
		"The worker returns completed results.\n\nApplications can retry failed requests.\n\n" +
		"Retries use a new connection.\n\nEach request has a deadline.\n\n" +
		"Clients validate response identifiers.\n\nA canceled operation returns its last error.\n\n" +
		"New messages wait in a bounded queue.\n\nWorkers process one message at a time.\n\n" +
		"Queue overflow rejects the new message.\n\nThe process exits after every worker stops.\n\n" +
		"Each payload uses UTF-8.\n\nHeader values retain their original case.\n\n" +
		"An empty response contains no records.\n\nMetrics count completed operations.\n\n" +
		"Logs include the request identifier.\n\nShutdown closes the listening socket.\n\n" +
		"The configuration sets the idle timeout.\n\nMalformed messages receive an error."
	sources := make([]document.Source, 192)
	for i := range sources {
		sources[i] = document.Source{Name: fmt.Sprintf("draft-%03d.md", i), Format: document.Markdown, Bytes: []byte(text)}
	}
	expected, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	c.Assert(expected.Gate.Passed, qt.IsTrue)
	return expected
}

func wireServer(t *testing.T) string {
	t.Helper()
	c := qt.New(t)
	name := "unswell-mcp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	// #nosec G204 -- Compile the fixed local server package into the test's temporary directory.
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "./cmd/unswell-mcp")
	build.Dir = "../.."
	output, err := build.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("%s", output))
	return binary
}

func checkWireEvidence(t *testing.T, path string, expected unswell.RunResult) {
	t.Helper()
	c := qt.New(t)
	// #nosec G304 -- Read the self-check artifact from this test's temporary directory.
	data, err := os.ReadFile(path)
	c.Assert(err, qt.IsNil)
	var result struct {
		Batches []server.CheckOutput `json:"repository_batches"`
	}
	c.Assert(json.Unmarshal(data, &result), qt.IsNil)
	c.Assert(len(result.Batches) > 1, qt.IsTrue)
	var names, want []string
	for _, batch := range result.Batches {
		c.Assert(batch.Result.Features, qt.IsNotNil)
		c.Assert(batch.Result.PreparedFeatures, qt.IsNotNil)
		for _, doc := range batch.Result.Documents {
			names = append(names, doc.Name)
		}
	}
	for _, doc := range expected.Documents {
		want = append(want, doc.Name)
	}
	c.Assert(names, qt.DeepEquals, want)
}
