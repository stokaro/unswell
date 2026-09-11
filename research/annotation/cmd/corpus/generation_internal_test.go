package main

// White-box tests: the requests, generations, and paired commands read and write explicit local files
// and wire the generation package; package main exposes no importable API for that wiring.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

func writeJSON(c *qt.C, dir, name string, value any) string {
	c.Helper()
	data, err := json.Marshal(value)
	c.Assert(err, qt.IsNil)
	path := filepath.Join(dir, name)
	c.Assert(os.WriteFile(path, data, 0o600), qt.IsNil)
	return path
}

func TestGenerationCommandsRunEndToEnd(t *testing.T) {
	c := qt.New(t)
	dir := t.TempDir()
	text := "Handler serves the request for the item and returns the stored value to the caller without delay at all."
	sheet, ok := generation.ExtractFactSheet([]byte("// x\nfunc Handler(w Writer, r *Request) error {\n"), 4, text)
	c.Assert(ok, qt.IsTrue)
	tasks := generation.Tasks{Version: generation.TasksVersion, Protocol: "unswell-llm-patterns-v1", Seed: "s", Cohort: "historical",
		Partitions: []string{"training"}, Roles: []string{"comment"}, Requested: 1, Strata: []generation.Stratum{},
		Tasks: []generation.Task{{ID: "t1", SourceID: "src1", UnitID: "u1", GroupID: "g", Partition: "training", Cohort: "historical",
			Repository: "org/lib", Ecosystem: "go", Path: "h.go", Role: "comment", Words: len(strings.Fields(text)), Text: text,
			TextSHA256: "x", FactSheet: sheet}}}
	tasksPath := writeJSON(c, dir, "tasks.json", tasks)
	var requestsOut bytes.Buffer
	c.Assert(run(t.Context(), []string{"requests", "--tasks", tasksPath, "--prompts", "../../../methods/prompts", "--run", "run-1"},
		nil, &requestsOut), qt.IsNil)
	var requests generation.Requests
	c.Assert(json.Unmarshal(requestsOut.Bytes(), &requests), qt.IsNil)
	c.Assert(requests.Requests, qt.HasLen, 4)
	c.Assert(requests.Prompts, qt.HasLen, 2)
	requestsPath := filepath.Join(dir, "requests.json")
	c.Assert(os.WriteFile(requestsPath, requestsOut.Bytes(), 0o600), qt.IsNil)
	responses := generation.Responses{Version: generation.ResponsesVersion, Run: "run-1", Family: "anthropic-claude", Model: "m",
		ModelBasis: "transcript", Harness: "agents", AgentType: "Explore", GeneratedOn: "2026-09-11", Parameters: "unavailable"}
	for i, request := range requests.Requests {
		status := "complete"
		if i == 3 {
			status = "refused"
		}
		responses.Responses = append(responses.Responses, generation.Response{RequestID: request.ID, Status: status,
			Text: "Handler serves a Request through a Writer and returns the stored value or an error to the caller."})
	}
	responsesPath := writeJSON(c, dir, "responses.json", responses)
	work := filepath.Join(dir, "work")
	historical := filepath.Join(dir, "historical", "records")
	c.Assert(os.MkdirAll(filepath.Join(work, "historical", "org__lib"), 0o750), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(work, "historical", "org__lib", "LICENSE"), []byte("MIT License\n"), 0o600), qt.IsNil)
	c.Assert(os.MkdirAll(historical, 0o750), qt.IsNil)
	writeJSON(c, historical, "org__lib.json", corpus.Acquisition{Version: corpus.AcquisitionVersion,
		Manifest: corpus.AcquisitionHeader{ID: "historical-org__lib", Seed: "unswell-llm-patterns-v1",
			Weights: corpus.Weights{Training: 5000, Development: 1500, Calibration: 1500, FinalTest: 2000},
			Policy:  extract.Policy{}, UnitKinds: []string{"sentence", "paragraph", "fragment"}},
		Repository: corpus.RepositoryRecord{Name: "org/lib", Commit: "abc", Topic: "libraries", Purpose: "Explain the library",
			Ecosystem: "go", Origin: annotation.Origin{Label: "unknown", Scope: "repository", Evidence: "Dated snapshot."},
			Rights:  annotation.Rights{License: "MIT", Evidence: "LICENSE", AllowedUses: []string{"annotation", "evaluation", "training"}},
			Notices: []string{"LICENSE"}}})
	records := filepath.Join(dir, "records.json")
	shards := filepath.Join(dir, "shards")
	var coverageOut bytes.Buffer
	c.Assert(run(t.Context(), []string{"generations", "--tasks", tasksPath, "--requests", requestsPath, "--responses", responsesPath,
		"--records", records, "--work", work, "--output", shards, "--historical", historical}, nil, &coverageOut), qt.IsNil)
	var coverage generation.Coverage
	c.Assert(json.Unmarshal(coverageOut.Bytes(), &coverage), qt.IsNil)
	c.Assert(coverage, qt.DeepEquals, generation.Coverage{Requests: 4, Complete: 3, Refused: 1})
	// The controlled shard plans and extracts from the files the command wrote.
	// #nosec G304 -- The test reads the shard the command just wrote under its temporary directory.
	shardData, err := os.ReadFile(filepath.Join(shards, "controlled-org__lib.json"))
	c.Assert(err, qt.IsNil)
	var plan, artifact bytes.Buffer
	c.Assert(run(t.Context(), []string{"plan"}, bytes.NewReader(shardData), &plan), qt.IsNil)
	c.Assert(run(t.Context(), []string{"extract", "--root", filepath.Join(work, "controlled", "org__lib")}, &plan, &artifact), qt.IsNil)
	var candidates corpus.Artifact
	c.Assert(json.Unmarshal(artifact.Bytes(), &candidates), qt.IsNil)
	c.Assert(len(candidates.Units) > 0, qt.IsTrue)
	c.Assert(candidates.Units[0].Cohort, qt.Equals, "controlled")
	// The paired tables read the record with the findings of the responses;
	// without any historical finding the pairs stay unmeasured but the command runs.
	policy := filepath.Join(dir, "policy.yaml")
	c.Assert(os.WriteFile(policy, []byte("version: 1\nextends: [builtin:technical-v1]\n"), 0o600), qt.IsNil)
	var measured bytes.Buffer
	c.Assert(run(t.Context(), []string{"measure", "--root", filepath.Join(work, "controlled", "org__lib"), "--policy", policy},
		bytes.NewReader(artifact.Bytes()), &measured), qt.IsNil)
	findings := filepath.Join(dir, "findings.json")
	c.Assert(os.WriteFile(findings, measured.Bytes(), 0o600), qt.IsNil)
	var pairedOut bytes.Buffer
	c.Assert(run(t.Context(), []string{"paired", "--records", records, "--tasks", tasksPath,
		"--classes", "../../../methods/rule-classes-v1.json", "--findings", findings}, nil, &pairedOut), qt.IsNil)
	var paired patterns.Paired
	c.Assert(json.Unmarshal(pairedOut.Bytes(), &paired), qt.IsNil)
	c.Assert(paired.Version, qt.Equals, patterns.PairedVersion)
	c.Assert(paired.Arms, qt.HasLen, 4)
	c.Assert(paired.Arms[0].UnmeasuredOriginals, qt.Equals, 1)
	for _, args := range [][]string{{"requests"}, {"generations", "--tasks", tasksPath}, {"paired", "--records", records}} {
		c.Assert(run(t.Context(), args, nil, &pairedOut), qt.IsNotNil, qt.Commentf("%v", args))
	}
}
