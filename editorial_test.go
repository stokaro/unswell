package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
)

func editorialEngine(t *testing.T, id, extra string) *unswell.Engine {
	t.Helper()
	c := qt.New(t)
	config := "version: 1\nextends: [builtin:custom]\nrules:\n  " + id + ":\n    enabled: true\n" + extra
	engine, err := unswell.New(unswell.Options{Config: []byte(config), IncludeSource: true})
	c.Assert(err, qt.IsNil)
	return engine
}

func editorialResult(t *testing.T, id, text, extra string) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	result, err := editorialEngine(t, id, extra).Analyze(t.Context(),
		document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	return result
}

func TestEditorialTechnicalCounterexamples(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"filler.section-announcement", "In this section, we will explain the protocol."},
		{"filler.empty-transition", "The flag records whether that being said was part of the transcript."},
		{"hype.vague-praise", "This is not a game-changing solution."},
		{"hype.metaphor-cluster", "The renderer draws a landscape. The travel API measures the journey. The textile is a tapestry."},
		{"hype.absolute-claim", "Does this guarantee complete safety? No. The service eliminates all risk only if every condition holds."},
		{"hype.absolute-claim", "The service cannot promise that it guarantees zero downtime."},
		{"filler.weak-intensifiers", "The very first request carries the very same identifier as the very last request in this retry sequence."},
		{"filler.stacked-hedging", "The request may fail, and the server might retry; the observer could disconnect."},
		{"filler.stacked-hedging", "The result may potentially change."},
		{"syntax.paired-contrast-density", "It is not about throughput. It is about latency."},
		{"syntax.triad-density", "The service requires authentication, authorization, and a checksum. It checks size, version, and type."},
		{"syntax.triad-density", "A powerful, innovative, transformative, unparalleled platform starts. " +
			"A powerful, innovative, transformative, unparalleled platform ends."},
		{"syntax.whether-preface-density", "Whether you are using HTTP or HTTPS, check the certificate. The client waits."},
		{"syntax.rhetorical-question-density", "Does the client retry? Yes. Does the server cache errors? No."},
		{"syntax.rhetorical-question-density", "The result? 30 milliseconds. The benefit? 20 fewer allocations."},
		{"syntax.not-only-density", "But it does not only read. But it does not only write."},
		{"syntax.not-only-density", "It not `only` reads but also writes. It not `only` checks but also validates."},
	} {
		t.Run(row.id+"/"+row.text, func(t *testing.T) {
			c := qt.New(t)
			result := editorialResult(t, row.id, row.text, "")
			c.Assert(result.Findings, qt.HasLen, 0)
		})
	}
}

func TestEditorialWindowBoundaries(t *testing.T) {
	c := qt.New(t)
	const first = "In this section, we will describe setup."
	const second = "In this section, we will describe deployment."
	for _, row := range []struct {
		name, middle string
		want         int
	}{
		{"adjacent paragraphs", "\n\n", 1},
		{"heading", "\n\n## Another topic\n\n", 1},
		{"setext heading", "\n\nAnother topic\n-------------\n\n", 1},
		{"code block", "\n\n```go\nconst Value = 1\n```\n\n", 0},
		{"code before heading", "\n\n```go\nconst Value = 1\n```\n\n## Another topic\n\n", 0},
		{"symbol code before heading", "\n\n    ===\n\n## Another topic\n\n", 0},
		{"protected sentence", " Check `Value` before continuing. ", 0},
		{"beyond window", " " + strings.Repeat("The client waits. ", 8), 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := editorialResult(t, "filler.section-announcement", first+row.middle+second, "")
			c.Assert(result.Findings, qt.HasLen, row.want)
		})
	}
	engine := editorialEngine(t, "filler.section-announcement", "")
	text := "package p\n// " + first + "\nfunc First() {}\n// " + second + "\nfunc Second() {}\n"
	result, err := engine.Analyze(t.Context(), document.Source{Name: "p.go", Format: document.Go, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 0)
	transition := "With that being said, the server starts.\n\n## Another topic\n\nIt goes without saying that the client waits."
	c.Assert(editorialResult(t, "filler.empty-transition", transition, "").Findings, qt.HasLen, 0)
}

func TestNewEditorialRulesRequireExplicitOptIn(t *testing.T) {
	c := qt.New(t)
	newRules := 0
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if d.Defaults.Enabled {
			continue
		}
		newRules++
		c.Assert(d.Status, qt.Equals, "experimental")
		c.Assert(d.Defaults.Gate, qt.Equals, "none")
		c.Assert(d.Examples, qt.HasLen, 2)
		for _, profile := range []string{"technical", "strict", "minimal", "business", "reference", "custom"} {
			engine, err := unswell.New(unswell.Options{Config: []byte("version: 1\nextends: [builtin:" + profile + "]\n")})
			c.Assert(err, qt.IsNil)
			policy, err := engine.PolicyForFile("guide.md")
			c.Assert(err, qt.IsNil)
			c.Assert(policy.Rules[d.ID].Enabled, qt.IsFalse)
		}
	}
	c.Assert(newRules, qt.Equals, 11)
}
