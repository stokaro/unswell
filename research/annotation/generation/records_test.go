package generation_test

import (
	"maps"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

const neutralPrompt = "# Condition `neutral`, version 1\n\nShared body:\n\n```text\n" +
	"Write about {words} words. Use the material below.\n```\n"

func taskSet() generation.Tasks {
	text := "Handler serves the request for the item and returns the stored value to the caller without delay at all."
	sheet, ok := generation.ExtractFactSheet([]byte("// x\nfunc Handler(w Writer, r *Request) error {\n"), 4, text)
	if !ok {
		panic("fixture fact sheet")
	}
	return generation.Tasks{Version: generation.TasksVersion, Protocol: "unswell-llm-patterns-v1", Seed: "s", Cohort: "historical",
		Tasks: []generation.Task{{ID: "t1", SourceID: "src1", UnitID: "u1", GroupID: "g", Partition: "training", Cohort: "historical",
			Repository: "org/lib", Ecosystem: "go", Path: "h.go", Role: "comment", Words: len(strings.Fields(text)), Text: text,
			TextSHA256: "x", FactSheet: sheet}}}
}

func TestRequestsPairTasksWithOperationsAndPrompts(t *testing.T) {
	c := qt.New(t)
	prompt, err := generation.ParsePrompt("neutral", "neutral-v1.md", []byte(neutralPrompt))
	c.Assert(err, qt.IsNil)
	c.Assert(prompt.Body, qt.Equals, "Write about {words} words. Use the material below.")
	_, err = generation.ParsePrompt("bad", "bad.md", []byte("no fence"))
	c.Assert(err, qt.IsNotNil)
	tasks := taskSet()
	requests, err := generation.BuildRequests("run-1", tasks, "tasks-sha", []generation.PromptCondition{prompt}, generation.Operations())
	c.Assert(err, qt.IsNil)
	c.Assert(requests.Requests, qt.HasLen, 2)
	generate, polish := requests.Requests[0], requests.Requests[1]
	c.Assert(generate.Operation, qt.Equals, "generate")
	c.Assert(generate.Text, qt.Contains, "Write the documentation described below.\n\nWrite about 19 words.")
	c.Assert(generate.Material, qt.Equals, tasks.Tasks[0].FactSheet.Text())
	c.Assert(generate.Text, qt.Contains, "Signature:\nfunc Handler(w Writer, r *Request) error")
	c.Assert(generate.Text, qt.Not(qt.Contains), "serves the request")
	c.Assert(polish.Material, qt.Equals, tasks.Tasks[0].Text)
	c.Assert(polish.Text, qt.Contains, "Edit the text below for clarity and correctness.")
	c.Assert(generate.ID, qt.Not(qt.Equals), polish.ID)
	c.Assert(generate.InputSHA256, qt.HasLen, 64)
	_, err = generation.BuildRequests("run-1", tasks, "x", []generation.PromptCondition{prompt}, []string{"invent"})
	c.Assert(err, qt.IsNotNil)
	_, err = generation.BuildRequests("", tasks, "x", []generation.PromptCondition{prompt}, generation.Operations())
	c.Assert(err, qt.IsNotNil)
}

func responses(requests generation.Requests) generation.Responses {
	generate, polish := requests.Requests[0], requests.Requests[1]
	return generation.Responses{Version: generation.ResponsesVersion, Run: "run-1", Family: "anthropic-claude",
		Model: "claude-x", ModelBasis: "harness environment", Harness: "agent", AgentType: "Explore", GeneratedOn: "2026-09-11",
		Parameters: "unavailable", Responses: []generation.Response{
			{RequestID: generate.ID, Text: "  Handler handles a Request through a Writer and reports an error when the item is missing.  ",
				Status: "complete", Tokens: 100},
			{RequestID: polish.ID, Text: "", Status: "refused", Note: "declined"},
		}}
}

func TestRecordsMeasureLengthOverlapAndCoverage(t *testing.T) {
	c := qt.New(t)
	prompt, _ := generation.ParsePrompt("neutral", "neutral-v1.md", []byte(neutralPrompt))
	tasks := taskSet()
	requests, err := generation.BuildRequests("run-1", tasks, "tasks-sha", []generation.PromptCondition{prompt}, generation.Operations())
	c.Assert(err, qt.IsNil)
	run := responses(requests)
	records, err := generation.BuildRecords(requests, "requests-sha", tasks, run)
	c.Assert(err, qt.IsNil)
	c.Assert(records.Version, qt.Equals, generation.GenerationVersion)
	c.Assert(records.Coverage, qt.DeepEquals, generation.Coverage{Requests: 2, Complete: 1, Refused: 1})
	byOp := map[string]generation.Record{}
	for _, record := range records.Records {
		byOp[record.Operation] = record
	}
	generate := byOp["generate"]
	c.Assert(generate.Text, qt.Equals, "Handler handles a Request through a Writer and reports an error when the item is missing.")
	c.Assert(generate.RealizedWords, qt.Equals, 16)
	c.Assert(generate.RequestedWords, qt.Equals, 19)
	c.Assert(generate.OffLength, qt.IsFalse)
	c.Assert(generate.Overlap, qt.Equals, 0.0)
	c.Assert(generate.Comparability, qt.DeepEquals, generation.Comparability{IdentifiersCovered: 4, IdentifiersTotal: 4})
	c.Assert(generate.Input[0].Text, qt.Equals, requests.Requests[0].Text)
	c.Assert(generate.Parameters, qt.Equals, "unavailable")
	c.Assert(generate.Cost, qt.Equals, "unavailable")
	c.Assert(generate.Contamination, qt.Equals, "unknown")
	polish := byOp["polish"]
	c.Assert(polish.Status, qt.Equals, "refused")
	c.Assert(polish.RealizedWords, qt.Equals, 0)
	// A verbatim copy of the original scores full overlap and is flagged.
	copied := run
	copied.Responses = []generation.Response{{RequestID: requests.Requests[1].ID, Text: tasks.Tasks[0].Text, Status: "complete"}}
	records, err = generation.BuildRecords(requests, "requests-sha", tasks, copied)
	c.Assert(err, qt.IsNil)
	c.Assert(records.Records[0].Overlap, qt.Equals, 1.0)
	c.Assert(records.Records[0].OverlapHigh, qt.IsTrue)
	c.Assert(records.Coverage.Missing, qt.Equals, 1)
	c.Assert(records.Coverage.OverlapHigh, qt.Equals, 1)
	// Unknown requests, repeated requests, unknown statuses, and missing
	// identity fields are refused.
	for _, edit := range []func(*generation.Responses){
		func(r *generation.Responses) { r.Responses[0].RequestID = "nope" },
		func(r *generation.Responses) { r.Responses = append(r.Responses, r.Responses[0]) },
		func(r *generation.Responses) { r.Responses[0].Status = "maybe" },
		func(r *generation.Responses) { r.Model = "" },
		func(r *generation.Responses) { r.Run = "other" },
	} {
		edited := responses(requests)
		edited.Responses = append([]generation.Response(nil), edited.Responses...)
		edit(&edited)
		_, err := generation.BuildRecords(requests, "x", tasks, edited)
		c.Assert(err, qt.IsNotNil)
	}
}

func TestManifestsImportCompleteResponsesOnly(t *testing.T) {
	c := qt.New(t)
	prompt, _ := generation.ParsePrompt("neutral", "neutral-v1.md", []byte(neutralPrompt))
	tasks := taskSet()
	requests, _ := generation.BuildRequests("run-1", tasks, "tasks-sha", []generation.PromptCondition{prompt}, generation.Operations())
	records, err := generation.BuildRecords(requests, "requests-sha", tasks, responses(requests))
	c.Assert(err, qt.IsNil)
	historical := corpus.Acquisition{Version: corpus.AcquisitionVersion,
		Manifest: corpus.AcquisitionHeader{ID: "historical-org__lib", Seed: "unswell-llm-patterns-v1",
			Weights: corpus.Weights{Training: 5000, Development: 1500, Calibration: 1500, FinalTest: 2000},
			Policy:  extract.Policy{}, UnitKinds: []string{"sentence", "paragraph", "fragment"}},
		Repository: corpus.RepositoryRecord{Name: "org/lib", Commit: "abc", Topic: "libraries", Purpose: "Explain the library",
			Ecosystem: "go", Origin: annotation.Origin{Label: "unknown", Scope: "repository", Evidence: "Dated snapshot."},
			Rights:  annotation.Rights{License: "MIT", Evidence: "LICENSE", AllowedUses: []string{"annotation", "evaluation", "training"}},
			Notices: []string{"LICENSE"}}}
	options := generation.ImportOptions{RecordsPath: "runs/run-1/records.json",
		Historical: func(string) (corpus.Acquisition, error) { return historical, nil },
		Notice:     func(string, string) ([]byte, error) { return []byte("MIT License\n"), nil }}
	imported, err := generation.BuildManifests(records, tasks, options)
	c.Assert(err, qt.IsNil)
	c.Assert(imported, qt.HasLen, 1)
	item := imported[0]
	c.Assert(item.Slug, qt.Equals, "org__lib")
	c.Assert(item.Manifest.ID, qt.Equals, "controlled-org__lib")
	c.Assert(item.Manifest.Sources, qt.HasLen, 1)
	source := item.Manifest.Sources[0]
	c.Assert(source.Path, qt.Equals, "generated/"+requests.Requests[0].ID+".md")
	var complete generation.Record
	for _, record := range records.Records {
		if record.Status == "complete" {
			complete = record
		}
	}
	c.Assert(string(item.Files[source.Path]), qt.Equals, complete.Text+"\n")
	c.Assert(string(item.Files["LICENSE"]), qt.Equals, "MIT License\n")
	c.Assert(source.Repository, qt.Equals, "org/lib")
	c.Assert(source.Role, qt.Equals, "comment")
	c.Assert(source.Origin.Label, qt.Equals, "generated")
	c.Assert(source.Origin.GenerationRecord, qt.Equals, "runs/run-1/records.json#"+requests.Requests[0].ID)
	c.Assert(source.GenerationTasks, qt.DeepEquals, []string{"t1"})
	c.Assert(source.Snapshot.Cohort, qt.Equals, "controlled")
	c.Assert(source.Notices[0].Path, qt.Equals, "LICENSE")
	// The manifest passes the corpus contract and plans with its files.
	files := map[string][]byte{}
	maps.Copy(files, item.Files)
	plan, err := corpus.MakePlan(t.Context(), item.Manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, files)
	c.Assert(err, qt.IsNil)
	c.Assert(artifact.Units[0].Cohort, qt.Equals, "controlled")
	// The source carries the origin; a unit never does, by the corpus contract.
	c.Assert(plan.Manifest.Sources[0].Origin.Label, qt.Equals, "generated")
	// A later run keeps its shards beside the earlier run's through the suffix.
	options.Suffix = "-run2"
	later, err := generation.BuildManifests(records, tasks, options)
	c.Assert(err, qt.IsNil)
	c.Assert(later[0].Manifest.ID, qt.Equals, "controlled-org__lib-run2")
	c.Assert(later[0].Slug, qt.Equals, "org__lib")
	c.Assert(artifact.Units[0].Unit.Origin.Label, qt.Equals, "unknown")
}
