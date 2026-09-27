package extract_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestPlainCodePreservesSurroundingProseAndRanges(t *testing.T) {
	for _, newline := range []string{"\n", "\r\n"} {
		t.Run(newline, func(t *testing.T) {
			c := qt.New(t)
			prefix := "\ufeffRésumé: the request must not retry before 30 seconds.\n\n  Use either call:\n"
			code := "     event_del_noblock(ev)\n     event_del_block(ev)\n"
			suffix := "  The second call must not return before completion.\n"
			prefix = strings.ReplaceAll(prefix, "\n", newline)
			code = strings.ReplaceAll(code, "\n", newline)
			suffix = strings.ReplaceAll(suffix, "\n", newline)
			text := prefix + code + suffix
			source := []byte(text)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "notes.txt", Format: document.Plain, Bytes: source}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(string(source), qt.Equals, text)
			c.Assert(doc.Excluded, qt.DeepEquals, []document.Exclusion{{
				Span: document.Span{Start: len(prefix), End: len(prefix) + len(code)}, Reason: "plain-code",
			}})
			c.Assert(doc.Blocks, qt.HasLen, 3)
			c.Assert(doc.Blocks[0].Text, qt.Contains, "must not retry before 30 seconds")
			c.Assert(doc.Blocks[1].Text, qt.Contains, "Use either call:")
			c.Assert(doc.Blocks[2].Text, qt.Contains, "must not return before completion")
			for _, block := range doc.Blocks {
				for _, span := range block.Map {
					c.Assert(span.Valid(len(source)), qt.IsTrue)
					c.Assert(span.End <= len(prefix) || span.Start >= len(prefix)+len(code), qt.IsTrue)
				}
			}
		})
	}
}

func TestPlainCodeRecognizesBoundedCExamples(t *testing.T) {
	for name, code := range map[string]string{
		"macro and body":         "    #define OP1 1\n    void func(int op) {\n      switch (op) {\n        ...\n      }\n    }\n",
		"separate functions":     "    void func_one(void) { ... }\n    void func_two(void) { ... }\n",
		"statement fragment":     "    struct event *ev;\n    ev = event_new(base, fd, cb);\n    event_add(ev, &tv);\n",
		"single terminated call": "    event_add(ev, &tv);\n",
		"include and omitted setup": "    #include <event.h>\n    ...\n    struct event *ev = malloc(sizeof(struct event));\n" +
			"    event_set(ev, fd, cb);\n",
	} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			prefix := "  The example uses this API:\n\n"
			text := prefix + code + "\n  Do not retry after a failed check.\n"
			doc, err := extract.Parse(t.Context(), document.Source{
				Name: "notes.txt", Format: document.Plain, Bytes: []byte(text),
			}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Excluded, qt.DeepEquals, []document.Exclusion{{
				Span: document.Span{Start: len(prefix), End: len(prefix) + len(code)}, Reason: "plain-code",
			}})
			c.Assert(doc.Blocks, qt.HasLen, 2)
			c.Assert(doc.Blocks[1].Text, qt.Contains, "Do not retry")
		})
	}
}

func TestPlainCodeRetainsUnprovenRegions(t *testing.T) {
	for name, body := range map[string]string{
		"indented prose": "    The client must not retry when (timeout) is zero.\n" +
			"    It returns an error; it does not discard the request.\n",
		"annotated identifiers":   "    event_add(ev) -- add the event\n    event_del(ev) -- remove the event\n",
		"single bare signature":   "    event_add(ev)\n",
		"quoted signature":        "    \"event_add(ev);\"\n",
		"ordinary semicolon list": "    long delays;\n    short retries;\n",
		"malformed program":       "    void function(void) {\n      return;\n",
		"code mixed with prose":   "    event_add(ev);\n    The request must not be lost.\n",
		"oversized region":        "    void function(void) {\n" + strings.Repeat("      call();\n", 129) + "    }\n",
		"oversized bytes":         "    void function(void) { /*" + strings.Repeat("x", 16<<10) + "*/ }\n",
	} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			text := "  Consider these details:\n" + body + "  Retain the condition.\n"
			doc, err := extract.Parse(t.Context(), document.Source{
				Name: "notes.txt", Format: document.Plain, Bytes: []byte(text),
			}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Excluded, qt.HasLen, 0)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Equals, text)
		})
	}
}

func TestPlainCodeNeedsIntroducerAndGreaterIndent(t *testing.T) {
	for _, text := range []string{
		"These calls are available.\n    event_add(ev);\n    event_del(ev);\n",
		"Use these calls:\nevent_add(ev);\nevent_del(ev);\n",
	} {
		c := qt.New(t)
		doc, err := extract.Parse(t.Context(), document.Source{
			Name: "notes.txt", Format: document.Plain, Bytes: []byte(text),
		}, extract.Options{})
		c.Assert(err, qt.IsNil)
		c.Assert(doc.Excluded, qt.HasLen, 0)
	}
}

func TestPlainCodeCandidateBudget(t *testing.T) {
	c := qt.New(t)
	text := strings.Repeat("Example:\n    event_add(ev);\n\n", 101)
	_, err := extract.Parse(t.Context(), document.Source{
		Name: "notes.txt", Format: document.Plain, Bytes: []byte(text),
	}, extract.Options{})
	c.Assert(err, qt.ErrorMatches, "plain-text code recognition exceeds 100 candidates")
}

func TestPlainCodeCancellation(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := extract.Parse(ctx, document.Source{Name: "notes.txt", Format: document.Plain,
		Bytes: []byte("Example:\n    event_add(ev);\n")}, extract.Options{})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
