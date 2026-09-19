package reviewbaseline_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/research/annotation/internal/reviewbaseline"
)

func encodedFixture(t *testing.T, input map[string]any) ([]byte, map[string]any) {
	t.Helper()
	c := qt.New(t)
	raw, err := json.Marshal(input)
	c.Assert(err, qt.IsNil)
	var text bytes.Buffer
	c.Assert(reviewbaseline.ExportText(t.Context(), bytes.NewReader(raw), &text), qt.IsNil)
	var export struct {
		Rows []struct {
			Page string `json:"page"`
			Unit int    `json:"unit"`
			Hash string `json:"text_sha256"`
			Text string `json:"text"`
		}
	}
	c.Assert(json.Unmarshal(text.Bytes(), &export), qt.IsNil)
	var rows []map[string]any
	for _, r := range export.Rows {
		c.Assert(fmt.Sprintf("%x", sha256.Sum256([]byte(r.Text))), qt.Equals, r.Hash)
		v := make([]float64, 128)
		v[r.Unit%2] = 1
		rows = append(rows, map[string]any{"page": r.Page, "unit": r.Unit, "text_sha256": r.Hash, "tokens": 10, "values": v})
	}
	return raw, map[string]any{"version": 1, "input_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)),
		"encoder_sha256": strings.Repeat("a", 64), "rows": rows}
}

func executeEncoded(t *testing.T, input []byte, vectors map[string]any) map[string]any {
	t.Helper()
	c := qt.New(t)
	raw, err := json.Marshal(vectors)
	c.Assert(err, qt.IsNil)
	var out bytes.Buffer
	c.Assert(reviewbaseline.RunEncoded(t.Context(), bytes.NewReader(input), bytes.NewReader(raw), &out), qt.IsNil)
	var result map[string]any
	c.Assert(json.Unmarshal(out.Bytes(), &result), qt.IsNil)
	return result
}

func TestEncodedModelsKeepUnavailableTargets(t *testing.T) {
	c := qt.New(t)
	raw, vectors := encodedFixture(t, fixture(t, false))
	rows := vectors["rows"].([]map[string]any)
	rows[0]["tokens"], rows[0]["values"], rows[0]["reason"] = 257, nil, "sequence_limit"
	result := executeEncoded(t, raw, vectors)
	c.Assert(result["rows"], qt.Equals, float64(30))
	models := result["models"].([]any)
	c.Assert(models, qt.HasLen, 10)
	for _, item := range models {
		m := item.(map[string]any)
		if m["evaluation_fold"] != float64(0) {
			continue
		}
		p := m["predictions"].([]any)[0].(map[string]any)
		c.Assert(p["raw_score"], qt.IsNil)
		c.Assert(p["selected"], qt.Equals, false)
		c.Assert(p["reason"], qt.Equals, "sequence_limit")
	}
}

func TestEncodedTrainingIgnoresEvaluationLabels(t *testing.T) {
	c := qt.New(t)
	input := fixture(t, false)
	raw, vectors := encodedFixture(t, input)
	first := executeEncoded(t, raw, vectors)["models"].([]any)
	for _, p := range input["pages"].([]map[string]any) {
		if p["fold"] != 0 {
			continue
		}
		for _, u := range p["units"].([]map[string]any) {
			u["label"] = 1 - u["label"].(int)
		}
	}
	raw, err := json.Marshal(input)
	c.Assert(err, qt.IsNil)
	vectors["input_sha256"] = fmt.Sprintf("%x", sha256.Sum256(raw))
	second := executeEncoded(t, raw, vectors)["models"].([]any)
	for i := range 2 {
		a, b := first[i].(map[string]any), second[i].(map[string]any)
		for _, field := range []string{"parameters", "features", "training_sha256", "operating_point"} {
			c.Assert(a[field], qt.DeepEquals, b[field], qt.Commentf("held-out labels affected %s", field))
		}
	}
}

func TestInvalidEncodingsWriteNoScores(t *testing.T) {
	for _, failure := range []string{"input", "encoder", "count", "binding", "dimension", "norm", "unavailable", "tokens", "reason"} {
		t.Run(failure, func(t *testing.T) {
			c := qt.New(t)
			raw, vectors := encodedFixture(t, fixture(t, false))
			rows := vectors["rows"].([]map[string]any)
			corruptEncoding(vectors, rows, failure)
			data, err := json.Marshal(vectors)
			c.Assert(err, qt.IsNil)
			var output bytes.Buffer
			c.Assert(reviewbaseline.RunEncoded(t.Context(), bytes.NewReader(raw), bytes.NewReader(data), &output), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
}

func corruptEncoding(v map[string]any, rows []map[string]any, failure string) {
	switch failure {
	case "input":
		v["input_sha256"] = "wrong"
	case "encoder":
		v["encoder_sha256"] = "wrong"
	case "count":
		v["rows"] = rows[:1]
	case "binding":
		rows[0]["text_sha256"] = "wrong"
	case "dimension":
		rows[0]["values"] = []float64{1}
	case "norm":
		rows[0]["values"].([]float64)[0] = 2
	case "unavailable":
		rows[0]["tokens"] = 257
	case "tokens":
		rows[0]["tokens"] = 0
	case "reason":
		rows[0]["reason"] = "invented"
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestEncodedCancellationAndWriterErrors(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(reviewbaseline.RunEncoded(ctx, strings.NewReader("{}"), strings.NewReader("{}"), io.Discard),
		qt.ErrorIs, context.Canceled)
	raw, v := encodedFixture(t, fixture(t, false))
	data, err := json.Marshal(v)
	c.Assert(err, qt.IsNil)
	c.Assert(reviewbaseline.RunEncoded(t.Context(), bytes.NewReader(raw), bytes.NewReader(data), failingWriter{}),
		qt.ErrorIs, io.ErrClosedPipe)
	c.Assert(reviewbaseline.ExportText(t.Context(), bytes.NewReader(raw), failingWriter{}), qt.ErrorIs, io.ErrClosedPipe)
}

func TestExportPreservesExtractedPiecesAndProtectedGaps(t *testing.T) {
	c := qt.New(t)
	source := "Keep the &amp; separator before `SECRET_CODE()` and never drop it.\n\n~~~go\nSECRET_FENCE()\n~~~\n"
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "fixture.md", Format: document.Markdown, Bytes: []byte(source)},
		extract.Options{IncludeStructure: true})
	c.Assert(err, qt.IsNil)
	var expected []map[string]any
	var texts []string
	for _, b := range doc.Blocks {
		units, err := nlp.PrepareUnits(t.Context(), b, provider, nlp.UnitOptions{Kinds: []string{"paragraph", "fragment"},
			Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences}, Limits: nlp.UnitLimits{MaxBytes: 65536,
				MaxContextBytes: 65536, MaxUnits: 100, MaxTokens: 1000, MaxSegments: 4096}})
		c.Assert(err, qt.IsNil)
		for _, u := range units {
			expected = append(expected, map[string]any{"binding": u.Binding(), "label": nil})
			texts = append(texts, u.Block().Text)
		}
	}
	p := map[string]any{"id": "fixture.md", "repository": "fixture", "source_group": "fixture/page", "cohort": "fixture",
		"format": "markdown", "sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(source))), "text": source,
		"group": "fixture", "fold": 0, "units": expected, "excluded_blocks": []int{}}
	data, err := json.Marshal(map[string]any{"version": 1, "basis": "exposed-assistant-unit-selection", "pages": []any{p}})
	c.Assert(err, qt.IsNil)
	var output bytes.Buffer
	c.Assert(reviewbaseline.ExportText(t.Context(), bytes.NewReader(data), &output), qt.IsNil)
	var exported struct {
		Rows []struct {
			Text string `json:"text"`
			Hash string `json:"text_sha256"`
		}
	}
	c.Assert(json.Unmarshal(output.Bytes(), &exported), qt.IsNil)
	var actual []string
	for _, r := range exported.Rows {
		actual = append(actual, r.Text)
		c.Assert(r.Hash, qt.Equals, fmt.Sprintf("%x", sha256.Sum256([]byte(r.Text))))
		c.Assert(r.Text, qt.Not(qt.Contains), "SECRET")
	}
	c.Assert(actual, qt.DeepEquals, texts)
	c.Assert(actual, qt.HasLen, 2)
	c.Assert(actual[0], qt.Contains, "& separator")
}
