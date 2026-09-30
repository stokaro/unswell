package claimreview_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation/internal/claimreview"
)

func inputFixture(t *testing.T) map[string]any {
	t.Helper()
	doc := parsed(t, "The café retries on failure.", document.Plain)
	spec := specification(reference(t, doc, string(doc.Source), 1), "wording")
	return map[string]any{"version": claimreview.Version,
		"source": map[string]any{"path": doc.Name, "format": doc.Format, "text": string(doc.Source), "sha256": doc.Hash},
		"claims": []claimreview.Specification{spec}, "stages": []claimreview.Stage{}, "approvals": []claimreview.Approval{}}
}

func commandInput(t *testing.T, input map[string]any) string {
	c := qt.New(t)
	t.Helper()
	data, err := json.Marshal(input)
	c.Assert(err, qt.IsNil)
	return string(data)
}

func TestCommandPreparationAndReplayRetainIDs(t *testing.T) {
	c := qt.New(t)
	input := inputFixture(t)
	var prepared bytes.Buffer
	c.Assert(claimreview.Run(context.Background(), strings.NewReader(commandInput(t, input)), &prepared), qt.IsNil)
	var output struct {
		Claims             []claimreview.Claim `json:"claims"`
		Complete           bool                `json:"complete"`
		EditorialQualified bool                `json:"editorial_qualified"`
	}
	c.Assert(json.Unmarshal(prepared.Bytes(), &output), qt.IsNil)
	c.Assert(output.Complete, qt.IsFalse)
	c.Assert(output.Claims[0].Targets[0].Quote, qt.Equals, "")
	c.Assert(output.Claims[0].Diagnostic, qt.Equals, "")
	input["stages"] = []claimreview.Stage{{ID: "selection", Decisions: []claimreview.Decision{
		{ClaimID: output.Claims[0].ID, Status: "retained", Reason: "Accounting replay, not a new quality judgment."},
	}, Edits: []claimreview.Edit{}, Duplicates: []claimreview.Duplicate{}}}
	var replayed bytes.Buffer
	c.Assert(claimreview.Run(context.Background(), strings.NewReader(commandInput(t, input)), &replayed), qt.IsNil)
	c.Assert(json.Unmarshal(replayed.Bytes(), &output), qt.IsNil)
	c.Assert(output.Complete, qt.IsTrue)
	c.Assert(output.EditorialQualified, qt.IsFalse)
	c.Assert(replayed.String(), qt.Not(qt.Contains), "café")
	input["include_raw"] = true
	c.Assert(claimreview.Run(context.Background(), strings.NewReader(commandInput(t, input)), &prepared), qt.IsNil)
	c.Assert(prepared.String(), qt.Contains, "café")
}

func TestCommandRejectsAmbiguousOrIncompleteInputs(t *testing.T) {
	c := qt.New(t)
	valid := commandInput(t, inputFixture(t))
	inputs := []string{
		strings.Replace(valid, `"version":`, `"unknown":true,"version":`, 1),
		strings.Replace(valid, `"version":`, `"version":"old","version":`, 1),
		strings.Replace(valid, `"version":`, `"Version":"old","version":`, 1),
		strings.Replace(valid, claimreview.Version, "unsupported", 1),
		strings.Replace(valid, `"stages":[]`, `"stages":null`, 1),
		valid + "{}",
		strings.Replace(valid, "The café retries", "The client retries", 1),
	}
	for _, input := range inputs {
		var output bytes.Buffer
		c.Assert(claimreview.Run(context.Background(), strings.NewReader(input), &output), qt.IsNotNil)
		c.Assert(output.Len(), qt.Equals, 0)
	}
}

func TestCommandUncertainReplayEmitsIncompleteAndFails(t *testing.T) {
	c := qt.New(t)
	input := inputFixture(t)
	var prepared bytes.Buffer
	c.Assert(claimreview.Run(context.Background(), strings.NewReader(commandInput(t, input)), &prepared), qt.IsNil)
	var bound struct {
		Claims []claimreview.Claim `json:"claims"`
	}
	c.Assert(json.Unmarshal(prepared.Bytes(), &bound), qt.IsNil)
	input["stages"] = []claimreview.Stage{{ID: "audit", Decisions: []claimreview.Decision{
		{ClaimID: bound.Claims[0].ID, Status: "uncertain", Reason: "No separate semantic assessment exists."},
	}, Edits: []claimreview.Edit{}, Duplicates: []claimreview.Duplicate{}}}
	var output bytes.Buffer
	err := claimreview.Run(context.Background(), strings.NewReader(commandInput(t, input)), &output)
	c.Assert(err, qt.ErrorMatches, "required claim accounting remains uncertain")
	c.Assert(output.String(), qt.Contains, `"complete":false`)
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("writer failed") }

func TestCommandWriterErrorAndCancellation(t *testing.T) {
	c := qt.New(t)
	input := commandInput(t, inputFixture(t))
	err := claimreview.Run(context.Background(), strings.NewReader(input), failingWriter{})
	c.Assert(err, qt.ErrorMatches, "writer failed")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.Assert(claimreview.Run(ctx, strings.NewReader(input), failingWriter{}), qt.Equals, context.Canceled)
}
