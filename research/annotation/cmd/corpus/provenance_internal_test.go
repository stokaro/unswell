package main

// White-box tests: the label source is chosen by unexported option parsing,
// and the property under test is that a round and provenance labels are
// alternatives, never combined or omitted together.

import (
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestProvenanceLabelsReplaceARound(t *testing.T) {
	c := qt.New(t)
	_, err := commandOptions([]string{"train", "--root", ".", "--labels", "provenance", "--feature", "prose-words", "--kind", "paragraph"})
	c.Assert(err, qt.IsNil)
	_, err = commandOptions([]string{"join", "--root", ".", "--labels", "provenance", "--feature", "prose-words"})
	c.Assert(err, qt.IsNil)
	_, err = commandOptions([]string{"join", "--root", ".", "--labels", "provenance", "--round", "round.json", "--feature", "prose-words"})
	c.Assert(err, qt.ErrorMatches, "join and train require --round or --labels, and at least one --feature")
	_, err = commandOptions([]string{"join", "--root", ".", "--feature", "prose-words"})
	c.Assert(err, qt.ErrorMatches, "join and train require --round or --labels, and at least one --feature")
	_, err = commandOptions([]string{"join", "--root", ".", "--labels", "guess", "--feature", "prose-words"})
	c.Assert(err, qt.ErrorMatches, "--labels accepts provenance or cohort")
	_, err = commandOptions([]string{"train", "--root", ".", "--labels", "provenance", "--lexical", "--kind", "paragraph"})
	c.Assert(err, qt.IsNil)
	_, err = commandOptions([]string{"train", "--root", ".", "--labels", "provenance", "--rule-config", "rules.yaml",
		"--feature", "policy.banned-phrases", "--kind", "paragraph"})
	c.Assert(err, qt.IsNil)
	_, err = commandOptions([]string{"measure", "--root", ".", "--policy", "p.yaml", "--labels", "provenance"})
	c.Assert(err, qt.ErrorMatches, "only join and train accept --round, --labels, and --feature")

	common := []string{"compare", "--plan", "p.json", "--protocol", "m.md", "--comparator", "b.json", "--corpus", "c.json"}
	_, err = comparisonFlags(append(slices.Clone(common), "--labels", "provenance"))
	c.Assert(err, qt.IsNil)
	_, err = comparisonFlags(append(slices.Clone(common), "--labels", "provenance", "--round", "r.json"))
	c.Assert(err, qt.ErrorMatches, "compare requires .*either --round or --labels")
	_, err = comparisonFlags(append(slices.Clone(common), "--labels", "guess"))
	c.Assert(err, qt.ErrorMatches, "--labels accepts provenance or cohort")

	_, err = evaluationFlags([]string{"evaluate", "--corpus", "c.json", "--labels", "provenance"})
	c.Assert(err, qt.IsNil)
	_, err = evaluationFlags([]string{"evaluate", "--corpus", "c.json", "--labels", "provenance", "--round", "r.json"})
	c.Assert(err, qt.ErrorMatches, "evaluate requires --corpus and either --round or --labels")
	_, err = evaluationFlags([]string{"evaluate", "--corpus", "c.json", "--labels", "guess"})
	c.Assert(err, qt.ErrorMatches, "--labels accepts provenance or cohort")
}
