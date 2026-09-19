package representation_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/iotest"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/research/annotation/internal/representation"
)

func input(text string) []byte {
	data, err := json.Marshal(map[string]any{"version": 1, "pages": []map[string]string{
		{"id": "example", "format": "markdown", "text": text, "sha256": fmt.Sprintf("%x", sha256.Sum256([]byte(text)))},
	}})
	if err != nil {
		panic(err)
	}
	return data
}

func TestExportPreservesProtectedPieceBoundaries(t *testing.T) {
	c := qt.New(t)
	text := "The client `--retry` option controls the retry count.\n"
	protectedStart := strings.Index(text, "`--retry`")
	protectedEnd := protectedStart + len("`--retry`")
	var output bytes.Buffer
	c.Assert(representation.Run(t.Context(), bytes.NewReader(input(text)), &output), qt.IsNil)
	var report unswell.RunResult
	c.Assert(json.Unmarshal(output.Bytes(), &report), qt.IsNil)
	c.Assert(report.Status, qt.Equals, "complete")
	c.Assert(report.Documents, qt.HasLen, 1)
	c.Assert(report.PreparedFeatures.Sources, qt.HasLen, 1)
	fragments := 0
	for _, unit := range report.PreparedFeatures.Sources[0].Units {
		c.Assert(unit.Binding.Kind, qt.Equals, "fragment")
		if unit.Binding.Kind == "fragment" {
			fragments++
		}
		for _, span := range unit.Segments {
			c.Assert(span.Start < protectedEnd && span.End > protectedStart, qt.IsFalse,
				qt.Commentf("protected flag leaked into prose: %v", span))
		}
	}
	c.Assert(fragments, qt.Equals, 2)
}

func TestRejectsSourceDriftAndDuplicateIDs(t *testing.T) {
	for _, change := range []string{"hash", "duplicate"} {
		t.Run(change, func(t *testing.T) {
			c := qt.New(t)
			var value map[string]json.RawMessage
			c.Assert(json.Unmarshal(input("A plain sentence."), &value), qt.IsNil)
			var pages []map[string]string
			c.Assert(json.Unmarshal(value["pages"], &pages), qt.IsNil)
			if change == "hash" {
				pages[0]["text"] = "A changed sentence."
			} else {
				pages = append(pages, pages[0])
			}
			data, err := json.Marshal(map[string]any{"version": 1, "pages": pages})
			c.Assert(err, qt.IsNil)
			var output bytes.Buffer
			c.Assert(representation.Run(t.Context(), bytes.NewReader(data), &output), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
}

func TestCanceledAndUnreadableInput(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var output bytes.Buffer
	c.Assert(representation.Run(ctx, bytes.NewReader(input("A sentence.")), &output), qt.ErrorIs, context.Canceled)
	c.Assert(output.Len(), qt.Equals, 0)
	c.Assert(representation.Run(t.Context(), iotest.ErrReader(context.DeadlineExceeded), &output),
		qt.ErrorIs, context.DeadlineExceeded)
	c.Assert(output.Len(), qt.Equals, 0)
}

func TestOriginalTokensRespectEngineLanguageExclusion(t *testing.T) {
	c := qt.New(t)
	var output bytes.Buffer
	c.Assert(representation.Run(t.Context(), bytes.NewReader(input(
		"这一段文字不是英文因此不应进入英语编辑规则的测量结果。\n\nThe client reads the retry count.\n")), &output), qt.IsNil)
	var report struct {
		Original []struct {
			Blocks []struct {
				Excluded bool  `json:"excluded"`
				Segments []any `json:"segments"`
			} `json:"blocks"`
		} `json:"original_tokens"`
	}
	c.Assert(json.Unmarshal(output.Bytes(), &report), qt.IsNil)
	c.Assert(report.Original, qt.HasLen, 1)
	c.Assert(report.Original[0].Blocks, qt.HasLen, 2)
	c.Assert(report.Original[0].Blocks[0].Excluded, qt.IsTrue)
	c.Assert(report.Original[0].Blocks[0].Segments, qt.HasLen, 0)
	c.Assert(report.Original[0].Blocks[1].Excluded, qt.IsFalse)
}
