package reviewbaseline_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/research/annotation/internal/reviewbaseline"
)

func fixture(t *testing.T, equalLength bool) map[string]any {
	t.Helper()
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	var pages []map[string]any
	markers := []string{"markerzero", "markerone", "markertwo", "markerthree", "markerfour"}
	for i := range 15 {
		positive := fmt.Sprintf("%s item%d has the ability to provide access to the selected configuration feature.", markers[i%5], i)
		negative := fmt.Sprintf("Client%d retries three times before returning an error.", i)
		if equalLength {
			positive, negative = fmt.Sprintf("Positive%d value is present.", i), fmt.Sprintf("Negative%d value is present.", i)
		}
		text := positive + "\n\n" + negative
		doc, err := extract.Parse(t.Context(), document.Source{Name: "test.md", Format: document.Markdown, Bytes: []byte(text)},
			extract.Options{IncludeStructure: true})
		c.Assert(err, qt.IsNil)
		var units []map[string]any
		for j, block := range doc.Blocks {
			prepared, err := nlp.PrepareUnits(t.Context(), block, provider, nlp.UnitOptions{
				Kinds: []string{"paragraph", "fragment"}, Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences},
				Limits: nlp.UnitLimits{MaxBytes: 65536, MaxContextBytes: 65536, MaxUnits: 100, MaxTokens: 1000, MaxSegments: 4096}})
			c.Assert(err, qt.IsNil)
			c.Assert(prepared, qt.HasLen, 1)
			units = append(units, map[string]any{"binding": prepared[0].Binding(), "label": 1 - j})
		}
		id := fmt.Sprintf("page-%02d", i)
		pages = append(pages, map[string]any{"id": id, "repository": "fixture", "source_group": "fixture/" + id,
			"cohort": "fixture", "format": "markdown", "sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(text))),
			"text": text, "group": id, "fold": i % 5, "units": units, "excluded_blocks": []int{}})
	}
	return map[string]any{"version": 1, "basis": "exposed-assistant-unit-selection", "pages": pages}
}

func execute(t *testing.T, input map[string]any) map[string]any {
	t.Helper()
	c := qt.New(t)
	data, err := json.Marshal(input)
	c.Assert(err, qt.IsNil)
	var output bytes.Buffer
	c.Assert(reviewbaseline.Run(t.Context(), bytes.NewReader(data), &output), qt.IsNil)
	var result map[string]any
	c.Assert(json.Unmarshal(output.Bytes(), &result), qt.IsNil)
	return result
}

func firstFoldModels(t *testing.T, result map[string]any) []map[string]any {
	t.Helper()
	c := qt.New(t)
	var models []map[string]any
	for _, item := range result["models"].([]any) {
		model := item.(map[string]any)
		if model["evaluation_fold"] == float64(0) {
			models = append(models, model)
		}
	}
	c.Assert(models, qt.HasLen, 4)
	return models
}

func TestTrainingDoesNotReadHeldOutLabelsOrVocabulary(t *testing.T) {
	c := qt.New(t)
	input := fixture(t, false)
	first := firstFoldModels(t, execute(t, input))
	for _, model := range first {
		for _, feature := range model["features"].([]any) {
			c.Assert(feature.(string), qt.Not(qt.Contains), "markerzero")
			c.Assert(feature.(string), qt.Not(qt.Contains), "markerone")
		}
	}
	for _, page := range input["pages"].([]map[string]any) {
		if page["fold"] != 0 {
			continue
		}
		for _, unit := range page["units"].([]map[string]any) {
			unit["label"] = 1 - unit["label"].(int)
		}
	}
	second := firstFoldModels(t, execute(t, input))
	for i := range first {
		for _, key := range []string{"features", "parameters", "training_sha256", "training_groups", "operating_point"} {
			c.Assert(second[i][key], qt.DeepEquals, first[i][key], qt.Commentf("held-out labels affected %s", key))
		}
	}
}

func TestTiedScoresCannotInventAnOperatingPoint(t *testing.T) {
	c := qt.New(t)
	result := execute(t, fixture(t, true))
	for _, item := range result["models"].([]any) {
		model := item.(map[string]any)
		if model["kind"] != "L" {
			continue
		}
		point := model["operating_point"].(map[string]any)
		c.Assert(point["available"], qt.Equals, false)
		c.Assert(point["threshold"], qt.IsNil)
		c.Assert(point["reason"], qt.Equals, "no_operating_point")
		for _, item := range model["predictions"].([]any) {
			c.Assert(item.(map[string]any)["selected"], qt.Equals, false)
		}
	}
}

func TestPageOrderDoesNotChangeFittedModels(t *testing.T) {
	c := qt.New(t)
	input := fixture(t, false)
	before := execute(t, input)
	slices.Reverse(input["pages"].([]map[string]any))
	after := execute(t, input)
	c.Assert(after["models"], qt.DeepEquals, before["models"])
}

func TestInvalidInputsReturnNoScores(t *testing.T) {
	c := qt.New(t)
	for _, failure := range []string{"source", "group", "label", "binding"} {
		t.Run(failure, func(t *testing.T) {
			c := qt.New(t)
			input := fixture(t, false)
			pages := input["pages"].([]map[string]any)
			switch failure {
			case "source":
				pages[0]["text"] = "Source changed."
			case "group":
				pages[1]["group"] = pages[0]["group"]
			case "label":
				pages[0]["units"].([]map[string]any)[0]["label"] = 2
			case "binding":
				pages[0]["units"].([]map[string]any)[0]["binding"] = nlp.UnitBinding{}
			}
			data, err := json.Marshal(input)
			c.Assert(err, qt.IsNil)
			var output bytes.Buffer
			c.Assert(reviewbaseline.Run(t.Context(), bytes.NewReader(data), &output), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var output bytes.Buffer
	c.Assert(reviewbaseline.Run(ctx, strings.NewReader("{}"), &output), qt.ErrorIs, context.Canceled)
}
