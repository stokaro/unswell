package generation_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

func documentFixture(t *testing.T) (corpus.Artifact, []byte, generation.DocumentBrief, *generation.Sampler) {
	t.Helper()
	c := qt.New(t)
	data := []byte("# Client\n\nThe client retries `Connect()` twice.\n\nIt preserves the timeout.\n")
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	source := corpus.Source{ID: "source", Path: "client.md", Repository: "org/client", Format: document.Markdown,
		Role: "documentation", Bytes: len(data), SHA256: digest, Snapshot: &corpus.Snapshot{Cohort: "historical"}}
	artifact := corpus.Artifact{Version: corpus.Version, Status: "unlabeled_candidates",
		Plan: corpus.Plan{Manifest: corpus.Manifest{Sources: []corpus.Source{source}}},
		Units: []corpus.Candidate{
			{SourceID: "source", Words: 7, Unit: annotation.Unit{Kind: "paragraph"}},
			{SourceID: "source", Words: 7, Unit: annotation.Unit{Kind: "paragraph"}},
			{SourceID: "source", Words: 1, Unit: annotation.Unit{Kind: "fragment"}},
			{SourceID: "source", Words: 14, Unit: annotation.Unit{Kind: "sentence"}},
		}}
	brief := generation.DocumentBrief{Version: "unswell-document-brief-v1", SourceID: "source", SourceSHA256: digest,
		Reviewer: "agent", Purpose: "Describe the retry contract.", Facts: []generation.DocumentFact{
			{Text: "Connect(): two retries; timeout preserved.", Span: document.Span{Start: 10, End: len(data)}},
		}}
	sampler, err := generation.NewSampler(generation.Options{Protocol: "test", Seed: "fixed", Cohort: "historical",
		Partitions: []string{"training"}, Roles: []string{"documentation"}, Count: 1, MinGroups: 1, MinWords: 15, MaxWords: 15,
		Ecosystems: map[string]string{"org/client": "go"}}, plan(map[string]string{"source": "training"}))
	c.Assert(err, qt.IsNil)
	return artifact, data, brief, sampler
}

func TestDocumentTasksKeepWholeSourceAndUseSeparateFacts(t *testing.T) {
	c := qt.New(t)
	artifact, data, brief, sampler := documentFixture(t)
	c.Assert(sampler.AddDocuments(t.Context(), artifact, func(string) ([]byte, error) { return data, nil },
		map[string]generation.DocumentBrief{"source": brief}), qt.IsNil)
	tasks, err := sampler.Sample()
	c.Assert(err, qt.IsNil)
	c.Assert(tasks.Groups, qt.Equals, 1)
	c.Assert(tasks.Tasks[0].Words, qt.Equals, 15)
	c.Assert(tasks.Tasks[0].Text, qt.Equals, string(data))
	c.Assert(tasks.Tasks[0].Scope, qt.Equals, generation.DocumentScope)
	c.Assert(tasks.Tasks[0].Brief.SHA256, qt.HasLen, 64)
	requests, err := generation.BuildRequests("test", tasks, "digest",
		[]generation.PromptCondition{{ID: "neutral", Body: "Use {words} words."}}, generation.Operations())
	c.Assert(err, qt.IsNil)
	c.Assert(requests.Requests[0].Material, qt.Equals, brief.Text())
	c.Assert(requests.Requests[0].Material, qt.Not(qt.Contains), "# Client")
	c.Assert(requests.Requests[1].Material, qt.Equals, string(data))
	// The sampler owns its facts; later curator edits cannot alter the tasks.
	brief.Facts[0].Text = "Changed"
	c.Assert(tasks.Tasks[0].Brief.Facts[0].Text, qt.Not(qt.Equals), "Changed")
	tasks.Tasks[0].Brief.Facts[0].Text = "Tampered"
	_, err = generation.BuildRequests("test", tasks, "digest", requests.Prompts, generation.Operations())
	c.Assert(err, qt.ErrorMatches, ".*invalid brief digest")
}

func TestDocumentTasksRejectMismatchedEvidenceAndIncompleteInputs(t *testing.T) {
	for _, name := range []string{"source bytes", "source ID", "fact range", "brief digest", "purpose encoding", "incomplete", "canceled"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			artifact, data, brief, sampler := documentFixture(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			switch name {
			case "source bytes":
				data = append(data, '!')
			case "source ID":
				brief.SourceID = "another"
			case "fact range":
				brief.Facts[0].Span.End = len(data) + 1
			case "brief digest":
				brief.SHA256 = "bad"
			case "purpose encoding":
				brief.Purpose = string([]byte{255})
			case "incomplete":
				artifact.Status = "failed"
			case "canceled":
				cancel()
			}
			c.Assert(sampler.AddDocuments(ctx, artifact, func(string) ([]byte, error) { return data, nil },
				map[string]generation.DocumentBrief{"source": brief}), qt.IsNotNil)
		})
	}
}

func TestDocumentTasksRejectRepeatedSources(t *testing.T) {
	c := qt.New(t)
	artifact, data, brief, sampler := documentFixture(t)
	read := func(string) ([]byte, error) { return data, nil }
	briefs := map[string]generation.DocumentBrief{"source": brief}
	c.Assert(sampler.AddDocuments(t.Context(), artifact, read, briefs), qt.IsNil)
	c.Assert(sampler.AddDocuments(t.Context(), artifact, read, briefs), qt.ErrorMatches, ".*already added")
}
