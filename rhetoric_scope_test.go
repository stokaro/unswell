package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestInformationEvaluationScope(t *testing.T) {
	for _, row := range []struct{ text, snippet string }{
		{"That distinction is the whole value of the verb: the probe says the credential failed.",
			"That distinction is the whole value of the verb"},
		{"This is the whole purpose: a survey measured the reported outcome.", "This is the whole purpose"},
		{"That is the whole guarantee, and it is worth stating precisely: both records commit together.",
			"That is the whole guarantee, and it is worth stating precisely"},
		{"That separation matters for embedders: each renderer accepts typed nodes.", "That separation matters for embedders"},
		{"This distinction matters.", "This distinction matters"},
		{"This distinction counts here.", "This distinction counts here"},
		{"This is the part worth reading twice.", "This is the part worth reading twice"},
		{"\"Every usable live feeder\" is the part worth reading twice.",
			"\"Every usable live feeder\" is the part worth reading twice"},
		{"“The final result” is the detail worth remembering.", "“The final result” is the detail worth remembering"},
		{"The two operations agree, which is exactly the question.", "which is exactly the question"},
		{"The check passes, which is the case worth having.", "which is the case worth having"},
		{"That distinction is the whole value of the verb, and the probe writes a status.",
			"That distinction is the whole value of the verb"},
		{"This is worth remembering: the client returns \"retry\" when the attempt fails.", "This is worth remembering"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.evaluative-closure", row.text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.Primary.Snippet, qt.Equals, row.snippet)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(row.text, row.snippet))
		})
	}
}

func TestInformationEvaluationControls(t *testing.T) {
	for _, text := range []string{
		"The author says: that distinction is the whole value of the verb.",
		"The guide argues — this is worth remembering.",
		"The report quoted: this is the whole guarantee.",
		"\"That distinction is the whole value of the verb\".",
		"The guide says \"this is the part worth reading twice\".",
		"This is the part worth reading twice, according to the author.",
		"This is the measured benefit.",
		"That distinction matters when resolving the destination.",
		"This is the whole value of the verb if the target is empty.",
		"That distinction is not the whole value of the verb.",
		"That distinction is the whole measured value.",
		"This is worth reading the buffer again.",
		"The counter counts incoming requests.",
		"The distinction counts for three points.",
		"The field matters for serialization.",
		"The retry is worth the cost of opening a connection.",
		"The cost is real: the transaction writes a second record.",
		"That is the whole guarantee offered by the database.",
		"The lock is what keeps the writes atomic.",
		"The whole value is read from the register.",
		"The program writes `This is worth remembering`.",
		"Is that distinction the whole value of the verb?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestInstructionProjectionSubjectBoundary(t *testing.T) {
	for _, text := range []string{
		"A form that allows a user to upload a file could be written like this in HTML.",
		"The option which allows users to upload files is disabled.",
		"[#2848](https://example.test/2848) Allow user to get/set the rounding method used when calculating relative time.",
		"Allow users to configure the rounding method.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
	for _, text := range []string{
		"This allows users to configure the method.",
		"The `Field` function can be used to configure the label.",
		"The upload form allows users to select the file.",
	} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 1)
	}
}

func TestInformationEvaluationMappingAndPolicy(t *testing.T) {
	id := "filler.evaluative-closure"
	for _, source := range []document.Source{
		{Name: "page.md", Format: document.Markdown,
			Bytes: []byte("\ufeffThat distinction is the whole **value** of the verb: the probe says café.\r\n")},
		{Name: "page.go", Format: document.Go,
			Bytes: []byte("package example\n// That distinction is the whole value of the verb: the probe says café.\nfunc Example() {}\n")},
		{Name: "page.py", Format: document.Python,
			Bytes: []byte("message = 'That distinction is the whole value of the verb: the probe says café.'\n")},
	} {
		c := qt.New(t)
		r, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(r.Findings, qt.HasLen, 1)
		f := r.Findings[0]
		c.Assert(string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]), qt.Equals, f.Primary.Snippet)
		c.Assert(f.Primary.Snippet, qt.Not(qt.Contains), "the probe")
	}
	text := "That distinction matters for embedders."
	c := qt.New(t)
	c.Assert(singleRuleResult(t, id, text, "", windowTerm(id, "That distinction matters for embedders")).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, strings.TrimSuffix(text, ".")+strings.Repeat(" extra", 100), "", "").Findings, qt.HasLen, 0)
}
