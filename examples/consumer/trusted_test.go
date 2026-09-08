package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestConsumerCanCompareTrustedPermissions(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	before := []document.Source{{Name: "guide.md", Format: document.Markdown, Bytes: []byte("Certainly! The client retries.")}}
	after := []document.Source{{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("<!-- unswell-disable-next-block scaffold.chat-preamble -- External wording. -->\n\nCertainly! The client retries.")}}
	result, err := engine.AnalyzeChangedWithOptions(t.Context(), before, after, unswell.ChangeOptions{TrustedPolicy: true})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.PolicyComparison.Version, qt.Equals, "unswell-policy-comparison-v1")
	c.Assert(result.PolicyComparison.Complete, qt.IsTrue)
	c.Assert(result.Suppressions[0].TrustState, qt.Equals, "untrusted")
}
