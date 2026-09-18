package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestQualitySubjectGrammar(t *testing.T) {
	for _, text := range []string{
		"Setting up a configuration file is simple.",
		"Bringing up the server is straightforward.",
		"The way the `--type` flag functions is simple.",
		"The way the loader parses records is straightforward.",
		"The way we configure the server is complicated.",
		"To deploy the service, opening a request is easy.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unscoped-assurance", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestNominalCapabilityChains(t *testing.T) {
	for _, text := range []string{
		"The adapter supports the ability to specify patterns.",
		"The library provides the capability to parse records.",
		"The renderer offers users the ability to format a date.",
		"The console gives the reader the capability to configure the queue.",
		"The adapter is designed to provide the ability to read records.",
		"The adapter provides users the ability to allow readers to specify a header.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
			c.Assert(r.Findings[0].Evidence.Suggestion, qt.Contains, "capability is not an obligation")
		})
	}
}

func TestAttentionGrammar(t *testing.T) {
	for _, text := range []string{
		"It is interesting to note here that the client retries.",
		"It is important to remember that the client retries when the connection closes.",
		"It is particularly useful to know that the client retries.",
		"The server caches records, and it is important to note that the client retries.",
		"That distinction is the whole point of reading the list.",
		"The distinction is the entire purpose of understanding the explanation.",
		"That is the part the reader has to read.",
		"This is the detail that users need to know.",
		"The result is the part of the output a reader has to look at.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestClauseGrammarControls(t *testing.T) {
	tests := map[string][]string{
		"filler.unscoped-assurance": {
			"Raise the timeout for connections that are slow to open.",
			"Setting up the server is simple when the provider is installed.",
			"The way the loader parses records is faster because it reuses a buffer.",
			"To deploy the service, select drivers which are easy to replace.",
			"The query avoids a scan and is faster.",
		},
		"filler.instruction-scaffolding": {
			"The adapter supports pattern matching.",
			"The service gives administrators the ability to rotate keys.",
			"The service provides users the ability to read records only after approval.",
			"The service does not support the ability to read records.",
			"The interface offers `the ability to parse` records.",
			"The adapter provides the ability to reconnect.",
			"The adapter supplies credentials that give users the ability to read records.",
		},
		"filler.evaluative-closure": {
			"It is important to verify that the signature matches.",
			"It is useful to check that the port is open.",
			"It is required to read the header before decoding the payload.",
			"This is the part the reader has to read before sending the request.",
			"This is the detail that users need to verify.",
			"The parser's purpose is reading the header.",
			"That distinction is the whole point of reading the list before deleting records.",
			"It is `important to remember that` the connection closes.",
			"It is important to note that the client retries.",
			"It is crucial to note that the client retries.",
		},
	}
	for id, texts := range tests {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) {
				qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
			})
		}
	}
}

func TestClauseGrammarSourceMapping(t *testing.T) {
	c := qt.New(t)
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown,
			Bytes: []byte("Résumé.\r\n\r\nTo configure the client, setting up the **service** is straightforward.")},
		{Name: "guide.go", Format: document.Go,
			Bytes: []byte("package sample\n// Setting up the café service is straightforward.\nfunc Example() {}\n")},
	} {
		r, err := singleRuleEngine(t, "filler.unscoped-assurance", "", "").Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(r.Findings, qt.HasLen, 1)
		f := r.Findings[0]
		c.Assert(string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]), qt.Equals, f.Primary.Snippet)
		c.Assert(strings.HasPrefix(strings.ToLower(f.Primary.Snippet), "setting up"), qt.IsTrue)
	}
}
