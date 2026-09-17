package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestContextualWordingConstructions(t *testing.T) {
	for _, row := range []struct{ id, text, snippet string }{
		{"filler.document-justification", "This page is what each option means.", "This page is what each option means"},
		{"filler.document-justification", "The guide owns the explanation.", "The guide owns the explanation"},
		{"filler.document-justification", "This guide covers the deployment process. It owns the sequence, not the detail.",
			"It owns the sequence, not the detail"},
		{"filler.evaluative-closure", "That distinction is the whole value of the verb:",
			"That distinction is the whole value of the verb"},
		{"filler.evaluative-closure", "This separation is the entire benefit of the approach.",
			"This separation is the entire benefit of the approach"},
		{"filler.evaluative-closure", "Several engine-specific points are worth knowing before adopting them:",
			"Several engine-specific points are worth knowing before adopting them"},
		{"filler.evaluative-closure", "The reason for each is worth knowing in advance.",
			"The reason for each is worth knowing in advance"},
		{"filler.unscoped-assurance", "Caddy exposes an unprecedented level of control compared to any web server in existence.",
			"Caddy exposes an unprecedented level of control compared to any web server in existence"},
		{"filler.unscoped-assurance", "The service delivers unparalleled performance.", "The service delivers unparalleled performance"},
		{"filler.unscoped-assurance", "The engine provides an unmatched degree of flexibility.",
			"The engine provides an unmatched degree of flexibility"},
		{"filler.unscoped-assurance", "Caddy is also ridiculously extensible, with a powerful plugin system.",
			"Caddy is also ridiculously extensible, with a powerful plugin system"},
		{"filler.unscoped-assurance", "The parser is unbelievably reliable.", "The parser is unbelievably reliable"},
		{"filler.unscoped-assurance", "The signature attests to a number nobody could have checked.", "nobody could have checked"},
		{"filler.unscoped-assurance", "A paused run is one nobody can act on.", "nobody can act on"},
		{"filler.unscoped-assurance", "Everybody wants the fallback.", "Everybody wants the fallback"},
		{"filler.unscoped-assurance", "No one could have verified the result.", "No one could have verified the result"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, row.id, row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			f := result.Findings[0]
			c.Assert(f.Primary.Snippet, qt.Equals, row.snippet)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(row.text, row.snippet))
			c.Assert(f.Evidence.Suggestion, qt.Not(qt.Equals), "")
		})
	}
}

func TestContextualWordingPreservesScope(t *testing.T) {
	for id, texts := range map[string][]string{
		"filler.document-justification": {
			"This page is what the browser cached.",
			"The file owns the instructions.",
			"This guide covers deployment.\n\nIt owns the sequence, not the detail.",
			"This guide covers deployment. The file owns the code. It owns the sequence, not the detail.",
			"This guide covers deployment, but the file contains code. It owns the sequence, not the detail.",
			"This guide covers deployment; it owns the sequence, not the detail.",
			"This guide owns the sequence when an import fails.",
			"This guide is not what the fields mean.",
			"The example says \"The guide owns the explanation\".",
			"The guide `owns` the explanation.",
		},
		"filler.evaluative-closure": {
			"The field is the whole value of the number.",
			"That distinction is the whole point of the isolation test.",
			"The distinction is the whole value of the feature when isolation is disabled.",
			"That distinction is not the whole value of the verb.",
			"Is that distinction the whole value of the verb?",
			"The result is worth caching before the next request.",
			"The reason is worth knowing because it changes the retry limit.",
			"The reason is worth `knowing` in advance.",
			"The example says \"The reason is worth knowing\".",
		},
		"filler.unscoped-assurance": {
			"The service does not deliver unparalleled performance.",
			"The service delivers unparalleled performance in the measured benchmark.",
			"The engine provides an unmatched degree of flexibility: it accepts both file formats.",
			"The engine provides unparalleled performance at 500 requests per second.",
			"The engine provides unmatched tokens to the error handler.",
			"The engine is incredibly fast when the cache is warm.",
			"The engine is incredibly fast because it skips repeated work.",
			"The engine is `incredibly fast`.",
			"The example says \"The engine is incredibly fast\".",
			"The benchmark returns \"ok\", and the engine is incredibly fast.",
			"Nobody can check the signature without the public key.",
			"No one can verify the signature without the public key.",
			"Nobody can act on the record until the lock expires.",
			"Everybody can verify the result if the public key is installed.",
			"Everybody wants the fallback only when the primary fails.",
			"The log contains `nobody could have checked`.",
			"The report disputes the claim nobody could have checked the result.",
			"The author says the service delivers unparalleled performance.",
			"The engine is not ridiculously extensible.",
		},
	} {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) {
				qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
			})
		}
	}
}

func TestContextualWordingMappingAndPolicy(t *testing.T) {
	c := qt.New(t)
	id := "filler.unscoped-assurance"
	text := "\ufeffThe café service delivers **unparalleled** performance.\r\n"
	result := singleRuleResult(t, id, text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "The café service delivers **unparalleled** performance")
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, len("\ufeff"))
	phrase := "The service delivers unparalleled performance"
	c.Assert(singleRuleResult(t, id, phrase+".", "", windowTerm(id, phrase)).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, phrase+".", "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
}
