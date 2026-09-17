package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestDiscourseRoleJudgments(t *testing.T) {
	for _, text := range []string{
		"The source column is the part worth reading.",
		"The third row is the one that matters most and the two agree on it.",
		"This is the part people miss.",
		"This step is very important!",
		"The distinction is important to understand.",
		"The detail is worth stating in full.",
		"This is what the signature is for.",
		"Matching the community binary is what the surface is for.",
		"I want to stress that this is the first release of the library.",
		"We would like to point out that the command resets the cache.",
		"I hope this guide helped you configure the client.",
		"We hope that this guide has helped readers configure the client.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestDiscourseRoleControls(t *testing.T) {
	for _, text := range []string{
		"The report contains the verdict, the row counts, and the findings.",
		"The row counts differ after the import.",
		"The row counts the number of requests.",
		"The row is important for cache invalidation.",
		"The column is the one that contains the checksum.",
		"The detail is not worth reading.",
		"The step is important when the connection fails.",
		"The reader says: this step is very important.",
		"This is what the buffer is for.",
		"This is what the wrapper is for.",
		"The allocator is what the plugin is for.",
		"We want to implement the next release.",
		"We believe that the worker failed.",
		"I hope the worker has finished the request.",
		"I hope this guide did not mislead you.",
		"Does this guide help you configure the client?",
		"I `want to stress that` the client retries.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.evaluative-closure", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestCopularQualityJudgments(t *testing.T) {
	for _, text := range []string{
		"The process is straightforward.",
		"Opening a backport request is fairly straightforward.",
		"The API is very easy to use.",
		"The implementation is faster than event notification.",
		"The implementation should be much faster than event notification.",
		"This is supposed to be faster.",
		"The interface is simple and flexible.",
		"The process is easy to configure.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unscoped-assurance", text, "", "").Findings, qt.HasLen, 1)
		})
	}
}

func TestCopularQualityControls(t *testing.T) {
	for _, text := range []string{
		"The response is a complex number.",
		"The request is a simple type.",
		"The simple API returns a response.",
		"The path is fast when the cache contains the result.",
		"The query is faster because the index covers the key.",
		"The query is faster by avoiding a full scan.",
		"The query is faster: the index covers the key.",
		"The query is 30 percent faster than a full scan.",
		"The query is faster in the benchmark.",
		"The query is not faster.",
		"Is the query faster?",
		"The user says the query is faster.",
		"The query is `faster`.",
		"The query is easy to lose after failure.",
		"The query is expensive for large inputs.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unscoped-assurance", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestRelativeOperandIsNotActor(t *testing.T) {
	for _, text := range []string{
		"It contains a digest that is combined with the cache keys of the inputs to determine " +
			"the stable checksum that can be used to cache the operation result.",
		"The response includes a token that can be used to fetch records.",
		"The request includes a filename that can be used to open the database.",
	} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
	}
}

func TestDiscourseRoleSourceMapping(t *testing.T) {
	text := "Résumé.\r\n\r\nThe source **column** is the part worth reading."
	r := singleRuleResult(t, "filler.evaluative-closure", text, "", "")
	c := qt.New(t)
	c.Assert(r.Findings, qt.HasLen, 1)
	f := r.Findings[0]
	c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(text, "The source"))
	c.Assert(f.Primary.Snippet, qt.Equals, "The source **column** is the part worth reading")
}
