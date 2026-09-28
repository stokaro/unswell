package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	ts "github.com/odvcencio/gotreesitter"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestCPPRawStringsPreserveContentAndParserIsolation(t *testing.T) {
	// A host application may select production parsing for its own parsers.
	// Unswell must preserve valid raw strings without changing that setting.
	previous := ts.AdmissionCandidateRouteDefault()
	ts.SetAdmissionCandidateRouteDefault(false)
	t.Cleanup(func() { ts.SetAdmissionCandidateRouteDefault(previous) })
	for _, literal := range []string{`R"()"`, `R"tag()tag"`, `u8R"tag()tag"`, `LR"()"`} {
		t.Run(literal, func(t *testing.T) {
			c := qt.New(t)
			source := "// Visible comment.\r\nconst auto empty = " + literal + ";\r\n" +
				"const auto text = R\"note(Visible raw \\d text.)note\";\r\n"
			for range 2 {
				doc, err := extract.Parse(t.Context(), document.Source{
					Name: "raw.cpp", Format: document.CPP, Bytes: []byte(source),
				}, extract.Options{})
				c.Assert(err, qt.IsNil)
				c.Assert(doc.Blocks, qt.HasLen, 2)
				c.Assert(strings.TrimSpace(doc.Blocks[0].Text), qt.Equals, "Visible comment.")
				c.Assert(doc.Blocks[1].Text, qt.Equals, `Visible raw \d text.`)
				assertSourceMap(c, doc, source)
				c.Assert(ts.AdmissionCandidateRouteDefault(), qt.IsFalse)
			}
		})
	}
}

func TestCPPInvalidRawStringsRemainErrors(t *testing.T) {
	for _, source := range []string{
		`const auto text = R"tag(unclosed";`,
		`const auto text = R"tag()other";`,
		`const auto text = R"tag()tag;`,
		`const auto text = R"tag()tag"`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := extract.Parse(t.Context(), document.Source{
				Name: "invalid.cpp", Format: document.CPP, Bytes: []byte(source),
			}, extract.Options{})
			qt.New(t).Assert(err, qt.IsNotNil)
		})
	}
}
