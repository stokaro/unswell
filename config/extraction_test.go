package config_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/config"
)

func TestExtractionPolicyValidation(t *testing.T) {
	cases := []struct{ name, exception string }{
		{"missing reason", "{id: fixture, paths: ['*.go'], kinds: [string]}"},
		{"missing id", "{paths: ['*.go'], kinds: [string], reason: Fixture}"},
		{"unknown kind", "{id: fixture, paths: ['*.go'], kinds: [strings], reason: Fixture}"},
		{"unknown format", "{id: fixture, paths: ['*.go'], kinds: [string], formats: [golang], reason: Fixture}"},
		{"bad path", "{id: fixture, paths: ['../*.go'], kinds: [string], reason: Fixture}"},
		{"bad glob", "{id: fixture, paths: ['[abc].go'], kinds: [string], reason: Fixture}"},
		{"missing path", "{id: fixture, kinds: [string], reason: Fixture}"},
		{"unknown field", "{id: fixture, paths: ['*.go'], kinds: [string], reason: Fixture, enabled: true}"},
		{"comment symbol", "{id: fixture, paths: ['*.go'], kinds: [comment], symbols: [fixture], reason: Fixture}"},
		{"empty symbol", "{id: fixture, paths: ['*.go'], kinds: [string], symbols: [''], reason: Fixture}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := config.Load([]byte("version: 1\nextraction:\n  exceptions:\n    - "+tc.exception+"\n"), nil)
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestExtractionPolicyHashAndDuplicateIDs(t *testing.T) {
	c := qt.New(t)
	base := "version: 1\nextraction:\n  exceptions:\n"
	item := "    - {id: fixture, paths: ['*.go'], kinds: [string], reason: Fixture}\n"
	policy, err := config.Load([]byte(base+item), nil)
	c.Assert(err, qt.IsNil)
	defaultPolicy, err := config.Load([]byte("version: 1\n"), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(policy.Hash, qt.Not(qt.Equals), defaultPolicy.Hash)
	c.Assert(policy.Extraction.Exceptions, qt.HasLen, 1)
	_, err = config.Load([]byte(base+item+item), nil)
	c.Assert(err, qt.ErrorMatches, `duplicate extraction exception "fixture"`)
}

func TestExtractionContextValidation(t *testing.T) {
	cases := []string{
		"contexts: [comments]",
		"contexts: [comment, comment]",
		"contexts: comment",
		"contexts: null",
		"languages: {csharp: {contexts: [heading]}}",
		"languages: {yaml: {contexts: [string, string]}}",
		"languages: {golang: {contexts: [comment]}}",
		"languages: {go: {}}",
		"languages: {go: {contexts: [comment], enabled: true}}",
	}
	for _, value := range cases {
		t.Run(value, func(t *testing.T) {
			c := qt.New(t)
			_, err := config.Load([]byte("version: 1\nextraction:\n  "+value+"\n"), nil)
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestExtractionContextSetHash(t *testing.T) {
	c := qt.New(t)
	first, err := config.Load(
		[]byte("version: 1\nextraction:\n  contexts: [comment, string]\n  languages: {yaml: {contexts: [string, comment]}}\n"),
		nil,
	)
	c.Assert(err, qt.IsNil)
	second, err := config.Load(
		[]byte("version: 1\nextraction:\n  contexts: [string, comment]\n  languages: {yaml: {contexts: [comment, string]}}\n"),
		nil,
	)
	c.Assert(err, qt.IsNil)
	c.Assert(first.Hash, qt.Equals, second.Hash)
	disabled, err := config.Load([]byte("version: 1\nextraction:\n  contexts: []\n  languages: {yaml: {contexts: []}}\n"), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(disabled.Extraction.Contexts, qt.IsNotNil)
	c.Assert(disabled.Hash, qt.Not(qt.Equals), first.Hash)
}
