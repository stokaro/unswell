package ruleset_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/ruleset"
)

func definition(match, extra string) []byte {
	return []byte(`version: 1
namespace: company
release: "1"
license: MIT
provenance: Maintained by the documentation team.
rules:
  - id: company.wording
    summary: Remove the repeated introduction.
    message: Start with the subject.
    scope: sentence
    severity: error
    gate: forbid
    examples:
      fail: ["Let's dive into the configuration."]
      pass: ["The client opens a connection."]
` + extra + "    match:\n" + indent(match, 6))
}

func indent(value string, spaces int) string {
	return strings.Repeat(" ", spaces) + strings.ReplaceAll(strings.TrimSpace(value), "\n", "\n"+strings.Repeat(" ", spaces)) + "\n"
}

func check(t *testing.T, definitions []byte, text string, format document.Format) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	options := unswell.Options{RuleSets: [][]byte{definitions},
		Config: []byte("version: 1\nextends: [builtin:custom-v1]\nrules:\n  company.wording:\n    enabled: true\n")}
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(context.Background(), document.Source{Name: "sample", Format: format, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	return result
}

func TestLexicalMatchers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, match, text string
		want              int
	}{
		{"contraction", "type: phrase\nat: start\nvalues: [\"Let's dive into\"]", "Let's dive into the configuration.", 1},
		{"curly contraction", "type: phrase\nvalues: [\"Let's dive into\"]", "Let’s dive into the configuration.", 1},
		{"case sensitive", "type: phrase\ncase_sensitive: true\nvalues: [Dive]", "We dive into the topic.", 0},
		{"whole token", "type: token-set\nvalues: [in]", "Look inside the file.", 0},
		{"end", "type: phrase\nat: end\nvalues: [further assistance]", "Ask for further assistance!", 1},
		{"end rejects following words", "type: phrase\nat: end\nvalues: [further assistance]", "Seek further assistance tomorrow.", 0},
		{"start", "type: phrase\nat: start\nvalues: [Moreover]", "The client, moreover, retries.", 0},
		{"regex", "type: regex\npattern: '\\bseamless(?:ly)?\\b'", "The client works seamlessly.", 1},
		{"duplicate values", "type: phrase\nvalues: [seamless, seamless]", "The seamless client retries.", 1},
		{"sentence boundary", "type: phrase\nvalues: [client retries]", "The client. Retries may fail.", 0},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := check(t, definition(test.match, ""), test.text, document.Plain)
			c := qt.New(t)
			c.Assert(result.Findings, qt.HasLen, test.want)
			c.Assert(result.Gate.Passed, qt.Equals, test.want == 0)
		})
	}
}

func TestRegexOriginalRanges(t *testing.T) {
	t.Parallel()
	definition := definition("type: regex\npattern: 'important to note'", "")
	for _, test := range []struct {
		name, text string
		format     document.Format
	}{
		{"markdown", "😀 It is **important** to note that retries may fail.\r\n", document.Markdown},
		{"Go escape", "package sample\nconst message = \"It is important\\u0020to note that retries may fail.\"\n", document.Go},
		{"YAML escape", "message: \"It is important\\u0020to note that retries may fail.\"\n", document.YAML},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			result := check(t, definition, test.text, test.format)
			c.Assert(result.Findings, qt.HasLen, 1)
			location := result.Findings[0].Primary
			c.Assert(location.Span.Start, qt.Equals, strings.Index(test.text, "important"))
			c.Assert(location.Span.End, qt.Equals, strings.Index(test.text, "note")+len("note"))
			for _, span := range location.Segments {
				c.Assert(span.Valid(len(test.text)), qt.IsTrue)
			}
			if test.format == document.Markdown {
				c.Assert(len(location.Segments) > 1, qt.IsTrue)
			}
		})
	}
}

func TestProtectedBoundaries(t *testing.T) {
	t.Parallel()
	for _, match := range []string{
		"type: phrase\nvalues: [important to note]",
		"type: regex\npattern: 'important.*note'",
		"type: sequence\ntokens:\n  - value: important\n  - gap: {min: 0, max: 10}\n  - value: note",
	} {
		result := check(t, definition(match, ""), "It is important `code` to note that retries may fail.", document.Markdown)
		qt.New(t).Assert(result.Findings, qt.HasLen, 0)
	}
}

func TestSequenceTargetsAndNegation(t *testing.T) {
	t.Parallel()
	match := `type: sequence
at: start
tokens:
  - value: This
  - gap: {min: 0, max: 2}
  - pos: JJ
    target: true
  - value: system
`
	c := qt.New(t)
	text := "This is a powerful system."
	result := check(t, definition(match, ""), text, document.Plain)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Evidence.Kind, qt.Equals, "heuristic")
	c.Assert(result.Findings[0].Primary.Span, qt.Equals,
		document.Span{Start: strings.Index(text, "powerful"), End: strings.Index(text, " system")})
	// The pinned tagger labels seamless as NN here; the matcher must not invent JJ.
	result = check(t, definition(match, ""), "This is a seamless system.", document.Plain)
	c.Assert(result.Findings, qt.HasLen, 0)
	negated := "type: sequence\ntokens:\n  - value: a\n  - value: reliable\n    not: true\n    target: true\n  - value: client"
	for _, test := range []struct {
		text string
		want int
	}{{"Use a reliable client.", 0}, {"Use a seamless client.", 1}} {
		result := check(t, definition(negated, ""), test.text, document.Plain)
		c.Assert(result.Findings, qt.HasLen, test.want)
	}
}

func TestConditionsCountersAndExceptions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, match, extra, text string
		want                     int
	}{
		{"count", "type: count\nmin: 2\nmatch: {type: token-set, values: [seamless]}", "",
			"This seamless and seamless client works.", 1},
		{"count rejects one", "type: count\nmin: 2\nmatch: {type: token-set, values: [seamless]}", "",
			"This seamless client works.", 0},
		{"density", "type: density\nper: words\nmin: 20\nmatch: {type: token-set, values: [seamless]}", "",
			"This seamless client works.", 1},
		{"feature guard", "type: all\nitems:\n  - {type: feature, name: prose.words, min: 4}\n  - {type: phrase, values: [seamless]}", "",
			"This seamless client works.", 1},
		{"guard fails", "type: all\nitems:\n  - {type: feature, name: prose.words, min: 10}\n  - {type: phrase, values: [seamless]}", "",
			"This seamless client works.", 0},
		{"any", "type: any\nitems:\n  - {type: token-set, values: [seamless]}\n  - {type: token-set, values: [effortless]}", "",
			"This effortless client works.", 1},
		{"not", "type: not\nmatch: {type: token-set, values: [retry]}", "", "The connection may fail.", 1},
		{"exception", "type: token-set\nvalues: [seamless]",
			"    except:\n      - {type: token-set, values: [quoted]}\n", "The quoted seamless client works.", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := check(t, definition(test.match, test.extra), test.text, document.Plain)
			qt.New(t).Assert(result.Findings, qt.HasLen, test.want)
		})
	}
}

func TestInvalidDefinitionsFailBeforeAnalysis(t *testing.T) {
	t.Parallel()
	for index, match := range []string{
		"type: shell\ncommand: echo",
		"type: phrase\nvalues: [word]\npattern: unsupported",
		"type: regex\npattern: '(?<=word)thing'",
		"type: regex\npattern: '[unterminated'",
		"type: regex\npattern: 'a*'",
		"type: sequence\ntokens: [{gap: {min: 0, max: 2}}]",
		"type: sequence\ntokens: [{value: a}, {gap: {min: 0, max: 33}}, {value: client}]",
		"type: sequence\ntokens: [{pos: NOUN}]",
		"type: feature\nname: dependencies.depth\nmin: 2",
		"type: count\nmin: .nan\nmatch: {type: phrase, values: [word]}",
		"type: count\nmin: 2\nmax: 1\nmatch: {type: phrase, values: [word]}",
		"type: phrase\nvalues: [word]\nvalues: [other]",
		"type: phrase\nvalues: null",
		`type: regex
pattern: '\b'`,
		`type: regex
pattern: '\b|word'`,
		"type: count\nmin: 1\nmatch: {type: feature, name: prose.words, min: 1}",
		"type: count\nmin: 1\nmatch: {type: not, match: {type: phrase, values: [word]}}",
		"type: density\nper: words\nmin: 1\nmatch: {type: count, min: 1, match: {type: phrase, values: [word]}}",
	} {
		t.Run(fmt.Sprintf("case%d", index), func(t *testing.T) {
			_, err := ruleset.Load(definition(match, ""))
			qt.New(t).Assert(err, qt.IsNotNil)
		})
	}
}
