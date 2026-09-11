package unswell_test

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
)

// rewriteFixtures is the E5 regression set of the pattern protocol: for every
// builtin rule, a text that carries the construction and a rewrite that keeps
// the facts without it. The rule fires on the first and stays silent on the
// second; the rewrite keeps every identifier and number of the original and at
// least half of its words, so a fixture cannot pass by deleting the content.
type rewriteFixtures struct {
	Version  string           `json:"version"`
	Fixtures []rewriteFixture `json:"fixtures"`
}

type rewriteFixture struct {
	Rule   string          `json:"rule"`
	Config string          `json:"config,omitempty"`
	Format document.Format `json:"format,omitempty"`
	Before string          `json:"before"`
	After  string          `json:"after"`
	// Dropped lists identifiers that belong to the construction itself, such
	// as the AI of a self-reference, which a rewrite removes on purpose.
	Dropped []string `json:"dropped,omitempty"`
}

var (
	// An identifier carries an uppercase letter after its first character, a
	// digit, or an underscore; a sentence-initial capital alone does not count.
	fixtureIdentifier = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
	fixtureNumber     = regexp.MustCompile(`\d+`)
)

func loadRewriteFixtures(t *testing.T) rewriteFixtures {
	t.Helper()
	c := qt.New(t)
	data, err := os.ReadFile("builtin/testdata/rewrites-v1.json")
	c.Assert(err, qt.IsNil)
	var fixtures rewriteFixtures
	c.Assert(json.Unmarshal(data, &fixtures), qt.IsNil)
	c.Assert(fixtures.Version, qt.Equals, "unswell-rewrite-fixtures-v1")
	return fixtures
}

func retainedTokens(text string) []string {
	var tokens []string
	for _, token := range fixtureIdentifier.FindAllString(text, -1) {
		if strings.ContainsAny(token, "_0123456789") || token[1:] != strings.ToLower(token[1:]) {
			tokens = append(tokens, token)
		}
	}
	return append(tokens, fixtureNumber.FindAllString(text, -1)...)
}

func rewriteFindings(t *testing.T, fixture rewriteFixture, text string) int {
	t.Helper()
	c := qt.New(t)
	config := "version: 1\nextends: [builtin:custom]\nrules:\n  " + fixture.Rule + ":\n    enabled: true\n" + fixture.Config
	engine, err := unswell.New(unswell.Options{Config: []byte(config)})
	c.Assert(err, qt.IsNil)
	format := fixture.Format
	if format == "" {
		format = document.Plain
	}
	result, err := engine.Analyze(t.Context(), document.Source{Name: "fixture.md", Format: format, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
	count := 0
	for _, finding := range result.Findings {
		c.Assert(finding.RuleID, qt.Equals, fixture.Rule)
		count++
	}
	return count
}

func TestRewriteFixturesCoverEveryBuiltinRule(t *testing.T) {
	c := qt.New(t)
	fixtures := loadRewriteFixtures(t)
	covered := map[string]int{}
	for _, fixture := range fixtures.Fixtures {
		covered[fixture.Rule]++
	}
	for _, implementation := range builtin.Rules() {
		id := implementation.Descriptor().ID
		c.Check(covered[id] > 0, qt.IsTrue, qt.Commentf("rule %s has no rewrite fixture", id))
		delete(covered, id)
	}
	c.Assert(covered, qt.HasLen, 0, qt.Commentf("fixtures name unknown rules"))
}

func TestRewriteFixturesStopFiringAfterTheRewrite(t *testing.T) {
	fixtures := loadRewriteFixtures(t)
	for _, fixture := range fixtures.Fixtures {
		t.Run(fixture.Rule, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(rewriteFindings(t, fixture, fixture.Before) > 0, qt.IsTrue, qt.Commentf("before: %q", fixture.Before))
			c.Assert(rewriteFindings(t, fixture, fixture.After), qt.Equals, 0, qt.Commentf("after: %q", fixture.After))
			for _, token := range retainedTokens(fixture.Before) {
				if slices.Contains(fixture.Dropped, token) {
					continue
				}
				c.Check(fixture.After, qt.Contains, token, qt.Commentf("the rewrite drops %q", token))
			}
			before, after := len(strings.Fields(fixture.Before)), len(strings.Fields(fixture.After))
			c.Check(after*2 >= before, qt.IsTrue, qt.Commentf("the rewrite keeps %d of %d words", after, before))
		})
	}
}
