package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

const activeReframing = "PostgreSQL does not keep the declaration. It stores the parsed form. " +
	"The reader does not copy the text. It prints the stored form."

func TestActionReframing(t *testing.T) {
	for _, text := range []string{
		activeReframing,
		strings.ReplaceAll(activeReframing, ". It", "; it"),
		strings.ReplaceAll(activeReframing, "does not", "doesn't"),
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "syntax.repeated-reframing", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.Primary.Snippet, qt.Contains, "keep the declaration")
			c.Assert(f.Related, qt.HasLen, 3)
			c.Assert(f.Related[0].Snippet, qt.Contains, "stores the parsed form")
			c.Assert(f.Evidence.Metrics[0].Value, qt.Equals, float64(2))
		})
	}
}

func TestActionReframingControls(t *testing.T) {
	for _, text := range []string{
		"PostgreSQL does not keep the declaration. It stores the parsed form.",
		"The client does not write the file. The server stores the request. " +
			"The reader does not copy the text. The writer prints the form.",
		"The server doesn't keep the declaration. It doesn't store the parsed form. " +
			"The reader doesn't copy the text. It doesn't print the form.",
		"The server keeps the declaration; the client stores the parsed form; the reader copies the text; the writer prints the form.",
		"The server does not `keep` the declaration. It stores the parsed form. " +
			"The reader does not `copy` the text. It prints the stored form.",
		"The author says: the server does not keep the declaration. It stores the parsed form. " +
			"The author says: the reader does not copy the text. It prints the stored form.",
		"\"PostgreSQL does not keep the declaration.\" It stores the parsed form. " +
			"\"The reader does not copy the text.\" It prints the stored form.",
		"Does PostgreSQL not keep the declaration? It stores the parsed form. " +
			"Does the reader not copy the text? It prints the stored form.",
		"PostgreSQL does not keep the declaration.\n\n## Storage\n\nIt stores the parsed form. " +
			"The reader does not copy the text. It prints the stored form.",
		"PostgreSQL does not keep the declaration. It stores the parsed form.\n\n```sql\nSELECT 1;\n```\n\n" +
			"The reader does not copy the text. It prints the stored form.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "syntax.repeated-reframing", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestDocumentMap(t *testing.T) {
	for _, text := range []string{
		"PostgreSQL is the primary target, and this page is the map of what Ptah manages beyond portable table DDL.",
		"This guide is a roadmap to how upgrades work.",
		"The section is a map of which settings control the runtime.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.document-metadiscourse", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Message, qt.Contains, "metaphorical map")
			c.Assert(r.Findings[0].Evidence.Suggestion, qt.Contains, "reference links")
		})
	}
}

func TestDocumentMapControls(t *testing.T) {
	for _, text := range []string{
		"This page is a map of Europe.",
		"This page is a geographic map of where the servers are located.",
		"The memory map shows which addresses are allocated.",
		"This page is not a map of what Ptah manages.",
		"The author says this page is a map of what Ptah manages.",
		"\"This page is a map of what Ptah manages.\"",
		"This page is the `map` of what Ptah manages.",
		"Is this page a map of what Ptah manages?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.document-metadiscourse", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestAbstractInformation(t *testing.T) {
	for _, text := range []string{
		"Offline, the declaration is the evidence.",
		"The specification is the proof.",
		"That description is the whole truth.",
		"That is a boundary rather than a missing feature.",
		"This is a design choice instead of a defect.",
		"That is a constraint, not a preference.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.evaluative-closure", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Message, qt.Contains, "abstract")
		})
	}
}

func TestAbstractInformationControls(t *testing.T) {
	for _, text := range []string{
		"The declaration is the evidence for the model selected in this run.",
		"The signed declaration is the evidence.",
		"The latency is the evidence.",
		"The declaration is not the evidence.",
		"The declaration is `the evidence`.",
		"If the model is offline, the declaration is the evidence.",
		"The author says: the declaration is the evidence.",
		"\"The declaration is the evidence.\"",
		"That is a boundary between databases rather than schemas.",
		"The checksum is a constraint rather than a preference.",
		"That is a boundary rather than a missing feature if the server supports it.",
		"That is a boundary rather than a missing feature because the server owns the operation.",
		"When refresh fails, that is a boundary rather than a missing feature.",
		"If refresh is unsupported, that is a boundary rather than a missing feature.",
		"Because the server owns refresh, that is a boundary rather than a missing feature.",
		"The author says that is a boundary rather than a missing feature.",
		"That is a `boundary` rather than a missing feature.",
		"\"That is a boundary rather than a missing feature.\"",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestRhetoricalRhythmSourceMapping(t *testing.T) {
	text := "\ufeffRésumé 🙂.\r\n\r\n**Offline, the declaration is the evidence.**\r\n\r\n" +
		"PostgreSQL is the target, and this page is the map of what the `schema` contains."
	c := qt.New(t)
	for _, row := range []struct{ id, snippet string }{
		{"filler.evaluative-closure", "Offline, the declaration is the evidence"},
		{"filler.document-metadiscourse", "this page is the map of what the `schema` contains"},
	} {
		r := singleRuleResult(t, row.id, text, "", "")
		c.Assert(r.Findings, qt.HasLen, 1)
		f := r.Findings[0]
		c.Assert(f.Primary.Snippet, qt.Equals, row.snippet)
		c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(text, row.snippet))
		c.Assert(text[f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, row.snippet)
	}
}
