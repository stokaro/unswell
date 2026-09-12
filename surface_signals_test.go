package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
)

const nominalProse = "We perform an evaluation of the implementation before the release."
const nounProse = "The service request response status code is recorded."
const passiveProse = "The request is carefully validated by the server before execution. " +
	"The response is securely recorded by the client after completion."

func surfaceMetric(t *testing.T, result unswell.RunResult, name string) float64 {
	t.Helper()
	c := qt.New(t)
	c.Assert(result.Findings, qt.HasLen, 1)
	for _, metric := range result.Findings[0].Evidence.Metrics {
		if metric.Name == name {
			return metric.Value
		}
	}
	t.Fatalf("missing metric %q", name)
	return 0
}

func TestSurfaceSyntaxMappingAndLocality(t *testing.T) {
	for _, row := range []struct{ id, text, snippet string }{
		{"syntax.nominalization-chain", strings.Replace(nominalProse, "evaluation", "**evaluation**", 1),
			"perform an **evaluation** of the implementation"},
		{"syntax.noun-stack", nounProse, "service request response status code"},
		{"syntax.passive-candidate-density", passiveProse, "is carefully validated"},
	} {
		t.Run(row.id, func(t *testing.T) {
			c := qt.New(t)
			text := "\ufeff\r\n" + row.text
			result := singleRuleResult(t, row.id, text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Primary.Snippet, qt.Equals, row.snippet)
			c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(text, row.snippet))
			clean := singleRuleResult(t, row.id, text+"\n\nThe connection closes.", "", "")
			c.Assert(clean.Findings, qt.DeepEquals, result.Findings)
			c.Assert(clean.Assessments[:len(result.Assessments)], qt.DeepEquals, result.Assessments)
		})
	}
}

func TestSurfaceSyntaxBoundaries(t *testing.T) {
	for _, row := range []struct{ name, id, text, parameters, extra string }{
		{"no verb", "syntax.nominalization-chain", "The evaluation of the implementation found a defect.", "", ""},
		{"no complement", "syntax.nominalization-chain", "We perform an evaluation before the release.", "", ""},
		{"suffix alone", "syntax.nominalization-chain", "We perform an operation of the client before the release.", "", ""},
		{"code boundary", "syntax.nominalization-chain", strings.Replace(nominalProse, "evaluation", "`evaluation`", 1), "", ""},
		{"approved noun", "syntax.nominalization-chain", nominalProse, "", "vocabulary:\n  terms: [evaluation]\n" +
			"  term_exemptions: [syntax.nominalization-chain]\n"},
		{"short noun sequence", "syntax.noun-stack", "The request has a status code.", "", ""},
		{"noun chunk boundary", "syntax.noun-stack", "The service request and response status code are recorded.", "", ""},
		{"proper name", "syntax.noun-stack", "The client uses TransportCacheEntry.", "", ""},
		{"format placeholder", "syntax.noun-stack", "configuration resource name %q", "", ""},
		{"approved term", "syntax.noun-stack", nounProse, "", "vocabulary:\n  terms: [response status code]\n" +
			"  term_exemptions: [syntax.noun-stack]\n"},
		{"predicate before preposition", "syntax.noun-stack",
			"The analysis completion state applies to every configured block in the document.", "", ""},
		{"predicate before object", "syntax.noun-stack",
			"Reader validation cannot reconstruct feature input hashes without the original inputs.", "", ""},
		{"predicate inside chunk", "syntax.noun-stack",
			"The prototype vet driver check comments remain disabled in every builtin profile.", "", ""},
		{"configured verb form", "syntax.noun-stack", nounProse, "{verbs: [request]}", ""},
		{"one passive candidate", "syntax.passive-candidate-density", strings.Split(passiveProse, ". ")[0] + ".", "", ""},
		{"stative adjective", "syntax.passive-candidate-density", strings.Repeat("The client is ready for the next request. ", 3), "", ""},
		{"passive window", "syntax.passive-candidate-density", passiveProse, "{window_sentences: 1}", ""},
		{"separate regions", "syntax.passive-candidate-density", strings.Replace(passiveProse, ". ", ".\n\n```go\nx()\n```\n\n", 1), "", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(singleRuleResult(t, row.id, row.text, row.parameters, row.extra).Findings, qt.HasLen, 0)
		})
	}
}

func TestSurfaceDictionariesAndNegation(t *testing.T) {
	c := qt.New(t)
	result := singleRuleResult(t, "syntax.nominalization-chain", "We execute a review of the implementation before release.",
		"{verbs: [execute], nouns: [review]}", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	text := strings.Replace(passiveProse, "is carefully", "is not carefully", 1)
	result = singleRuleResult(t, "syntax.passive-candidate-density", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "is not carefully validated")
	c.Assert(result.Findings[0].Related, qt.HasLen, 1)
}

func TestNounStackKeepsSubjectHeads(t *testing.T) {
	for _, row := range []struct{ name, text string }{
		{"finite verb after the stack", "The service request response status code changes after a retry."},
		{"configured form as chunk-final head", "The service request response stores are recorded."},
		{"plural head", "The access control policy requirements apply."},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "syntax.noun-stack", row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(surfaceMetric(t, result, "consecutive-common-nouns") >= 4, qt.IsTrue)
		})
	}
}

func TestSurfaceRulesRequireOptIn(t *testing.T) {
	c := qt.New(t)
	ids := []string{"syntax.nominalization-chain", "syntax.noun-stack", "syntax.passive-candidate-density", "syntax.parenthetical-load",
		"readability.long-paragraph", "readability.grade-metric", "format.em-dash-density", "format.list-fragmentation"}
	for _, profile := range []string{"technical", "strict", "minimal", "business", "reference", "custom"} {
		engine, err := unswell.New(unswell.Options{Config: []byte("version: 1\nextends: [builtin:" + profile + "]\n")})
		c.Assert(err, qt.IsNil)
		policy, err := engine.PolicyForFile("guide.md")
		c.Assert(err, qt.IsNil)
		for _, id := range ids {
			c.Assert(policy.Rules[id].Enabled, qt.IsFalse, qt.Commentf("%s/%s", profile, id))
			c.Assert(policy.Rules[id].Gate, qt.Equals, "none")
		}
	}
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if d.ID == "readability.grade-metric" || strings.HasPrefix(d.ID, "format.") {
			c.Assert(d.Status, qt.Equals, "experimental")
			c.Assert(d.Defaults.Score.Weight, qt.Equals, 0)
		}
	}
}

func TestSurfaceEscapedGoStringsAndSeparateComments(t *testing.T) {
	c := qt.New(t)
	text := "package sample\n// " + strings.Replace(passiveProse, ". ", ".\nfunc First() {}\n// ", 1) + "\nfunc Second() {}\n"
	engine := singleRuleEngine(t, "syntax.passive-candidate-density", "", "")
	result, err := engine.Analyze(t.Context(), document.Source{Name: "sample.go", Format: document.Go, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 0)
	text = "package sample\nconst message = \"The client waits \\u2014 then retries \\u2014 if permitted.\"\n"
	engine = singleRuleEngine(t, "format.em-dash-density", "{min_words: 1}", "")
	result, err = engine.Analyze(t.Context(), document.Source{Name: "sample.go", Format: document.Go, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "\\u2014")
	c.Assert(result.Findings[0].Related, qt.HasLen, 1)
}
