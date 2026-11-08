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
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation/internal/reviewbaseline"
)

type inventoryTestOutput struct {
	Version   int    `json:"version"`
	Basis     string `json:"basis"`
	InputHash string `json:"input_sha256"`
	Contract  string `json:"unit_contract"`
	Pages     []struct {
		ID       string               `json:"id"`
		Hash     string               `json:"sha256"`
		Excluded []document.Exclusion `json:"excluded"`
		Units    []struct {
			Ordinal  int             `json:"ordinal"`
			Binding  nlp.UnitBinding `json:"binding"`
			Text     string          `json:"text"`
			RawSpan  document.Span   `json:"raw_span"`
			RawSlice string          `json:"raw_slice"`
		} `json:"units"`
	} `json:"pages"`
}

func inventoryTestInput(t *testing.T, source string) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"version": 1, "basis": "source-function-inventory", "pages": []any{
		map[string]any{"id": "fixture.md", "format": "markdown", "sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(source))), "text": source},
	}})
	qt.New(t).Assert(err, qt.IsNil)
	return raw
}

func TestSourceInventoryRetainsOriginalSpansAndProtectedBoundaries(t *testing.T) {
	c := qt.New(t)
	source := "Café &amp; tea helps. Café &amp; tea helps.\n\nKeep `SECRET()` private.\n\n~~~go\nSECRET_FENCE()\n~~~\n"
	input := inventoryTestInput(t, source)
	var output bytes.Buffer
	c.Assert(reviewbaseline.ExportSourceInventory(t.Context(), bytes.NewReader(input), &output), qt.IsNil)
	var exported inventoryTestOutput
	c.Assert(json.Unmarshal(output.Bytes(), &exported), qt.IsNil)
	c.Assert(exported.Version, qt.Equals, 1)
	c.Assert(exported.Basis, qt.Equals, "source-function-inventory")
	c.Assert(exported.InputHash, qt.Equals, fmt.Sprintf("%x", sha256.Sum256(input)))
	c.Assert(exported.Contract, qt.Equals, nlp.UnitContract)
	c.Assert(exported.Pages, qt.HasLen, 1)
	page := exported.Pages[0]
	c.Assert(page.Hash, qt.Equals, fmt.Sprintf("%x", sha256.Sum256([]byte(source))))
	c.Assert(len(page.Units) >= 2, qt.IsTrue)
	c.Assert(len(page.Excluded) > 0, qt.IsTrue)
	var repeated []document.Span
	for i, unit := range page.Units {
		c.Assert(unit.Ordinal, qt.Equals, i)
		c.Assert(unit.Binding.Contract, qt.Equals, nlp.UnitContract)
		c.Assert(len(unit.Binding.Segments) > 0, qt.IsTrue)
		c.Assert(unit.RawSpan.Valid(len(source)), qt.IsTrue)
		c.Assert(unit.RawSlice, qt.Equals, source[unit.RawSpan.Start:unit.RawSpan.End])
		c.Assert(unit.Text, qt.Not(qt.Contains), "SECRET")
		c.Assert(unit.RawSlice, qt.Not(qt.Contains), "SECRET")
		for _, segment := range unit.Binding.Segments {
			c.Assert(segment.Valid(len(source)), qt.IsTrue)
			c.Assert(segment.Start >= unit.RawSpan.Start && segment.End <= unit.RawSpan.End, qt.IsTrue)
		}
		if unit.Text == "Café & tea helps." {
			c.Assert(unit.Binding.Kind, qt.Equals, "sentence")
			c.Assert(unit.RawSlice, qt.Contains, "&amp;")
			repeated = append(repeated, unit.RawSpan)
		}
	}
	c.Assert(repeated, qt.HasLen, 2)
	c.Assert(repeated[0], qt.Not(qt.Equals), repeated[1])
}

func TestSourceInventoryRejectsInvalidInputWithoutPartialOutput(t *testing.T) {
	valid := inventoryTestInput(t, "A meaningful sentence.")
	for name, input := range map[string][]byte{
		"hash":          bytes.Replace(valid, []byte(`"sha256":"`), []byte(`"sha256":"wrong`), 1),
		"unknown field": bytes.Replace(valid, []byte(`"version":1`), []byte(`"extra":true,"version":1`), 1),
		"trailing JSON": append(bytes.Clone(valid), []byte("{}")...),
		"format":        bytes.Replace(valid, []byte(`"markdown"`), []byte(`"unknown"`), 1),
		"NUL":           inventoryTestInput(t, "A meaningful sentence.\x00"),
	} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			c.Assert(reviewbaseline.ExportSourceInventory(t.Context(), bytes.NewReader(input), &output), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
	var output bytes.Buffer
	emptyPages := `{"version":1,"basis":"source-function-inventory","pages":[]}`
	qt.New(t).Assert(reviewbaseline.ExportSourceInventory(t.Context(), strings.NewReader(emptyPages), &output), qt.IsNotNil)
	qt.New(t).Assert(output.Len(), qt.Equals, 0)
}

func TestSourceInventoryCancellationAndWriterFailure(t *testing.T) {
	c := qt.New(t)
	input := inventoryTestInput(t, "A meaningful sentence.")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(reviewbaseline.ExportSourceInventory(ctx, bytes.NewReader(input), &bytes.Buffer{}), qt.ErrorIs, context.Canceled)
	c.Assert(reviewbaseline.ExportSourceInventory(t.Context(), bytes.NewReader(input), failingWriter{}), qt.ErrorIs, io.ErrClosedPipe)
}
