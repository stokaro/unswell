package main

// White-box tests: unexported option parsing chooses the bank options. A
// compression bank stands apart from the lexical and rule baselines. A
// reservation joins any of them.

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCompressionBankOptionsStandApart(t *testing.T) {
	c := qt.New(t)
	_, err := commandOptions([]string{"train", "--root", ".", "--labels", "cohort", "--compression-bank", "bank.json", "--kind", "paragraph"})
	c.Assert(err, qt.IsNil)
	parsed, err := commandOptions([]string{"train", "--root", ".", "--labels", "cohort", "--compression-bank", "bank.json",
		"--feature", "prose-words", "--kind", "paragraph"})
	c.Assert(err, qt.IsNil)
	c.Assert(parsed.compressionBank, qt.Equals, "bank.json")
	_, err = commandOptions([]string{"train", "--root", ".", "--labels", "cohort", "--reserve-bank", "bank.json",
		"--feature", "prose-words", "--kind", "paragraph"})
	c.Assert(err, qt.IsNil)
	_, err = commandOptions([]string{"train", "--root", ".", "--labels", "cohort", "--reserve-bank", "bank.json",
		"--lexical", "--kind", "paragraph"})
	c.Assert(err, qt.IsNil)
	for _, extra := range [][]string{{"--lexical"}, {"--rule-config", "rules.yaml", "--feature", "activation/x"},
		{"--reserve-bank", "other.json"}} {
		args := append([]string{"train", "--root", ".", "--labels", "cohort", "--compression-bank", "bank.json",
			"--kind", "paragraph"}, extra...)
		_, err = commandOptions(args)
		c.Assert(err, qt.ErrorMatches, "--compression-bank excludes --lexical, --rule-config, and --reserve-bank")
	}
	_, err = commandOptions([]string{"join", "--root", ".", "--labels", "cohort", "--compression-bank", "bank.json"})
	c.Assert(err, qt.IsNotNil)

	options, err := evaluationFlags([]string{"predict", "--root", ".", "--model", "m.json", "--plan", "p.json",
		"--protocol", "m.md", "--compression-bank", "bank.json"})
	c.Assert(err, qt.IsNil)
	c.Assert(options.bank, qt.Equals, "bank.json")

	_, err = compressionBankFlags([]string{"reference-bank", "--root", ".", "--labels", "cohort", "--selection", "s.json"})
	c.Assert(err, qt.IsNil)
	_, err = compressionBankFlags([]string{"reference-bank", "--root", ".", "--round", "r.json", "--selection", "s.json"})
	c.Assert(err, qt.IsNil)
	_, err = compressionBankFlags([]string{"reference-bank", "--root", ".", "--round", "r.json", "--labels", "cohort",
		"--selection", "s.json"})
	c.Assert(err, qt.ErrorMatches, "reference-bank requires .*either --round or --labels.*")
	_, err = compressionBankFlags([]string{"reference-bank", "--root", ".", "--labels", "guess", "--selection", "s.json"})
	c.Assert(err, qt.ErrorMatches, "--labels accepts provenance or cohort")
}
