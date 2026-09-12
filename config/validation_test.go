package config_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/rule"
)

func builtinCatalog() []rule.Descriptor {
	catalog := []rule.Descriptor{}
	for _, implementation := range builtin.Rules() {
		catalog = append(catalog, implementation.Descriptor())
	}
	return catalog
}

const strictBase = "version: 1\nextends: [builtin:strict]\n"

// A configuration error must stop the run before any document is read. Each
// case names the setting it breaks so a later refactor cannot silently drop one
// of these checks and still pass.
func TestInvalidPolicySettingsFailBeforeAnalysis(t *testing.T) {
	catalog := builtinCatalog()
	for _, row := range []struct{ name, text, want string }{
		{"escaping include pattern", "files:\n  include: ['../outside/**']\n",
			`glob must be project relative: "\.\./outside/\*\*"`},
		{"unsupported glob syntax", "files:\n  include: ['[bad']\n",
			`unsupported glob "\[bad"; use \*, \*\*, and \?`},
		{"incomplete analysis", "analysis:\n  require_complete: false\n", "alpha requires complete analysis"},
		{"unavailable NLP backend", "analysis:\n  nlp: spacy\n", `unknown NLP backend "spacy"`},
		{"zero analysis limit", "analysis:\n  max_file_bytes: 0\n",
			"analysis limits must be positive and at most 1 GiB/count"},
		{"oversized analysis limit", "analysis:\n  max_tokens: 2000000000\n",
			"analysis limits must be positive and at most 1 GiB/count"},
		{"unknown gate mode", "gate:\n  mode: some\n", "gate.mode must be all or new"},
		{"incomplete runs allowed", "gate:\n  fail_on_incomplete: false\n", "alpha requires complete analysis"},
		{"score threshold below one", "gate:\n  sentence_score:\n    fail_at: 0\n", "invalid score threshold"},
		{"negative threshold words", "gate:\n  paragraph_score:\n    min_words: -1\n", "invalid score threshold"},
		{"probability gate without a pack", "gate:\n  probability:\n    fail_at: 0.5\n",
			"gate.probability requires calibration.model: pack"},
		{"probability threshold above one", "gate:\n  probability:\n    fail_at: 2\n",
			"gate.probability.fail_at must be greater than 0 and at most 1"},
		{"experimental pack gating a build",
			"calibration:\n  model: pack\n  accept_experimental: true\ngate:\n  probability:\n    fail_at: 0.5\n",
			"gate.probability requires an accepted pack; calibration.accept_experimental cannot gate a build"},
		{"unknown rule identifier", "rules:\n  no.such.rule:\n    enabled: true\n",
			`.unswell.yaml: unknown rule ID "no.such.rule"`},
		{"unknown top-level field", "unknown: true\n", "(?s).*field unknown not found.*"},
		{"duplicate key", "version: 2\n", `.unswell.yaml: duplicate configuration key "version"`},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := config.Load([]byte(strictBase+row.text), catalog)
			c.Assert(err, qt.ErrorMatches, row.want)
		})
	}
}

// Every accepted rule parameter has a declared range. A rule that lists a
// parameter without a range check would accept nonsense and fail later, inside
// analysis, where the message cannot name the setting.
func TestOutOfRangeRuleParametersFailValidation(t *testing.T) {
	catalog := builtinCatalog()
	for _, row := range []struct{ name, text, want string }{
		{"severity", "hype.vague-praise:\n    severity: fatal\n", `hype.vague-praise: invalid severity "fatal"`},
		{"gate", "hype.vague-praise:\n    gate: always\n", `hype.vague-praise: invalid gate "always"`},
		{"score weight", "hype.vague-praise:\n    score:\n      weight: 200\n",
			`hype.vague-praise: score weight and cap must be in \[0,100\]`},
		{"score cap", "hype.vague-praise:\n    score:\n      cap: -5\n",
			`hype.vague-praise: score weight and cap must be in \[0,100\]`},
		{"phrase position", "filler.wordy-phrase:\n    parameters:\n      positions: [middle]\n",
			`filler.wordy-phrase: invalid phrase position "middle"`},
		{"min_words", "density.connective-overuse:\n    parameters:\n      min_words: 20000\n",
			"density.connective-overuse: invalid min_words parameter"},
		{"onset", "syntax.long-sentence:\n    parameters:\n      onset: -1\n",
			"syntax.long-sentence: invalid onset parameter"},
		{"saturation", "syntax.long-sentence:\n    parameters:\n      onset: 5\n      saturation: 0\n",
			"syntax.long-sentence: invalid saturation parameter"},
		{"allowed_occurrences", "filler.empty-transition:\n    parameters:\n      allowed_occurrences: -1\n",
			"filler.empty-transition: invalid allowed_occurrences parameter"},
		{"saturation_occurrences",
			"filler.empty-transition:\n    parameters:\n      allowed_occurrences: 3\n      saturation_occurrences: 0\n",
			"filler.empty-transition: invalid saturation_occurrences parameter"},
		{"window_sentences", "filler.empty-transition:\n    parameters:\n      window_sentences: 0\n",
			"filler.empty-transition: invalid window_sentences parameter"},
		{"opener_words", "repetition.sentence-openers:\n    parameters:\n      opener_words: 9\n",
			"repetition.sentence-openers: invalid opener_words parameter"},
		{"max_answer_words", "syntax.rhetorical-question-density:\n    parameters:\n      max_answer_words: 0\n",
			"syntax.rhetorical-question-density: invalid max_answer_words parameter"},
		{"min_ngram_words", "repetition.ngram-density:\n    parameters:\n      min_ngram_words: 9\n",
			"repetition.ngram-density: invalid min_ngram_words parameter"},
		{"window_blocks", "format.list-fragmentation:\n    parameters:\n      window_blocks: 0\n",
			"format.list-fragmentation: invalid window_blocks parameter"},
		{"min_sentences", "readability.grade-metric:\n    parameters:\n      min_sentences: 0\n",
			"readability.grade-metric: invalid min_sentences parameter"},
		{"sentence_words", "readability.long-paragraph:\n    parameters:\n      sentence_words: 0\n",
			"readability.long-paragraph: invalid sentence_words parameter"},
		{"min_long_sentences", "readability.long-paragraph:\n    parameters:\n      min_long_sentences: 0\n",
			"readability.long-paragraph: invalid min_long_sentences parameter"},
		{"allowed_depth", "syntax.parenthetical-load:\n    parameters:\n      allowed_depth: 20\n",
			"syntax.parenthetical-load: invalid allowed_depth parameter"},
		{"saturation_depth", "syntax.parenthetical-load:\n    parameters:\n      saturation_depth: 40\n",
			"syntax.parenthetical-load: invalid saturation_depth parameter"},
		{"min_insertion_words", "syntax.parenthetical-load:\n    parameters:\n      min_insertion_words: 101\n",
			"syntax.parenthetical-load: invalid min_insertion_words parameter"},
		{"max_item_words", "format.list-fragmentation:\n    parameters:\n      max_item_words: 0\n",
			"format.list-fragmentation: invalid max_item_words parameter"},
		{"max_list_items", "format.list-fragmentation:\n    parameters:\n      max_list_items: 0\n",
			"format.list-fragmentation: invalid max_list_items parameter"},
		{"similarity", "repetition.near-sentence:\n    parameters:\n      similarity: 0\n",
			"repetition.near-sentence: invalid similarity parameter"},
		{"window", "repetition.exact-sentence:\n    parameters:\n      window: block\n",
			"repetition.exact-sentence: invalid window parameter"},
		{"dictionary word", "syntax.nominalization-chain:\n    parameters:\n      verbs: ['two words']\n",
			"syntax.nominalization-chain: verb and noun dictionaries require single words containing only letters"},
		{"empty dictionary entry", "syntax.nominalization-chain:\n    parameters:\n      nouns: ['']\n",
			"syntax.nominalization-chain: phrases must contain 1 to 1000 bytes"},
		{"empty phrase", "filler.wordy-phrase:\n    parameters:\n      phrases: ['']\n",
			"filler.wordy-phrase: phrases must contain 1 to 1000 bytes"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := config.Load([]byte(strictBase+"rules:\n  "+row.text), catalog)
			c.Assert(err, qt.ErrorMatches, row.want)
		})
	}
}

// The document itself is validated before any setting: size, version, language,
// aliases, and the number of YAML documents all decide whether the policy can be
// read at all.
func TestMalformedConfigurationDocumentsFail(t *testing.T) {
	catalog := builtinCatalog()
	for _, row := range []struct{ name, text, want string }{
		{"oversized document", strings.Repeat("# comment\n", 120000),
			"configuration resources exceed 1 MiB each or 8 MiB total"},
		{"unsupported version", "version: 2\n", ".unswell.yaml: unsupported config version 2"},
		{"unsupported language", strictBase + "language: fr\n", `.unswell.yaml: unsupported language "fr"`},
		{"two YAML documents", strictBase + "---\nversion: 1\n",
			".unswell.yaml: configuration must contain one YAML document"},
		{"YAML alias", "version: 1\nfiles: &a {include: ['**/*.md']}\noverrides: []\nvocabulary: *a\n",
			".unswell.yaml: YAML aliases and null overrides are not supported"},
		{"null override", "version: 1\nfiles: null\n",
			".unswell.yaml: YAML aliases and null overrides are not supported"},
		{"non-builtin profile", "version: 1\nextends: [policies/base.yaml]\n",
			`missing configuration resource "policies/base.yaml"`},
		{"unknown profile", "version: 1\nextends: [builtin:aggressive]\n", `.unswell.yaml: unknown profile "aggressive"`},
		{"unsupported calibration model", strictBase + "calibration:\n  model: forest\n",
			`.unswell.yaml: unsupported calibration model "forest"`},
		{"invalid calibration policy", strictBase + "calibration:\n  model: pack\n  on_incompatible: ignore\n",
			`.unswell.yaml: invalid calibration incompatibility policy "ignore"`},
		{"calibration without a model", strictBase + "calibration:\n  on_incompatible: fail\n",
			".unswell.yaml: calibration options require an explicit model"},
		{"unsupported origin model", strictBase + "origin:\n  model: forest\n",
			`.unswell.yaml: unsupported origin model "forest"`},
		{"invalid origin policy", strictBase + "origin:\n  model: pack\n  on_incompatible: ignore\n",
			`.unswell.yaml: invalid origin incompatibility policy "ignore"`},
		{"origin without a model", strictBase + "origin:\n  accept_experimental: true\n",
			".unswell.yaml: origin options require an explicit model"},
		{"too many rule sets",
			strictBase + "rule_sets:\n" + strings.Repeat("  - {version: 1, namespace: x, rules: []}\n", 11),
			".unswell.yaml: configuration exceeds 10 rule_sets"},
		{"invalid rule set", strictBase + "rule_sets:\n  - {version: 2}\n",
			".unswell.yaml: ruleset requires version 1 and a namespace such as company"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := config.Load([]byte(row.text), catalog)
			c.Assert(err, qt.ErrorMatches, row.want)
		})
	}
}

// An absent configuration is a valid one: the default profile loads without a file.
func TestEmptyConfigurationLoadsTheDefaultProfile(t *testing.T) {
	c := qt.New(t)
	policy, err := config.Load(nil, builtinCatalog())
	c.Assert(err, qt.IsNil)
	c.Assert(policy.Hash, qt.HasLen, 64)
	c.Assert(policy.Analysis.NLP, qt.Equals, "builtin-en")
	c.Assert(policy.Calibration, qt.IsNil)
	c.Assert(policy.Origin, qt.IsNil)
}
