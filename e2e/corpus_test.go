package e2e_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestCorpusSourcePreparation(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	manifest, err := os.ReadFile("../research/annotation/corpus/testdata/ptah-manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	root := "../research/annotation/corpus/testdata/ptah"
	output := researchCommand(t, binary, []string{"extract", "--root", root}, plan, 0)
	var artifact struct {
		Status string `json:"status"`
		Units  []struct {
			SourceID  string `json:"source_id"`
			Partition string `json:"partition"`
			Unit      struct {
				ID, Kind, Role, Text, Context string
				Source                        struct {
					Segments []document.Span `json:"segments"`
				}
			} `json:"unit"`
		} `json:"units"`
	}
	c.Assert(json.Unmarshal(output, &artifact), qt.IsNil)
	c.Assert(artifact.Status, qt.Equals, "unlabeled_candidates")
	projection := make([]corpusExpectedUnit, 0, len(artifact.Units))
	for _, candidate := range artifact.Units {
		c.Assert(candidate.Partition, qt.Equals, "development")
		unit := candidate.Unit
		projection = append(projection, corpusExpectedUnit{unit.ID, candidate.SourceID, unit.Kind, unit.Role,
			unit.Text, unit.Context, unit.Source.Segments})
	}
	actual, err := json.MarshalIndent(projection, "", "  ")
	c.Assert(err, qt.IsNil)
	expected, err := os.ReadFile("corpusdata/ptah-units.golden.json")
	c.Assert(err, qt.IsNil)
	c.Assert(append(actual, '\n'), qt.DeepEquals, expected)
	verified := researchCommand(t, binary, []string{"verify", "--root", root}, output, 0)
	var verification struct {
		Status      string `json:"status"`
		HumanCorpus string `json:"human_corpus"`
		Units       int    `json:"units"`
		SourceCount int    `json:"source_count"`
	}
	c.Assert(json.Unmarshal(verified, &verification), qt.IsNil)
	c.Assert(verification.Status, qt.Equals, "source_and_candidates_reproduced")
	c.Assert(verification.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(verification.Units, qt.Equals, 378)
	c.Assert(verification.SourceCount, qt.Equals, 8)
	policy := filepath.Join(t.TempDir(), "policy.yaml")
	c.Assert(os.WriteFile(policy, []byte("version: 1\nextends: [builtin:custom]\nrules:\n"+
		"  policy.banned-phrases: {enabled: true, parameters: {phrases: [schema]}}\n"), 0o600), qt.IsNil)
	measured := researchCommand(t, binary, []string{"measure", "--root", root, "--policy", policy}, output, 0)
	var findings struct {
		Status      string `json:"status"`
		HumanCorpus string `json:"human_corpus"`
		Documents   []struct {
			SourceID string         `json:"source_id"`
			Findings int            `json:"findings"`
			Unbound  int            `json:"unbound"`
			ByRule   map[string]int `json:"by_rule"`
		} `json:"documents"`
		Units []struct {
			UnitID   string `json:"unit_id"`
			Findings []struct {
				RuleID string `json:"rule_id"`
			} `json:"findings"`
		} `json:"units"`
	}
	c.Assert(json.Unmarshal(measured, &findings), qt.IsNil)
	c.Assert(findings.Status, qt.Equals, "verified_targets_with_policy_findings")
	c.Assert(findings.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(findings.Documents, qt.HasLen, 8)
	c.Assert(findings.Units, qt.HasLen, 378)
	total := 0
	for _, doc := range findings.Documents {
		c.Assert(doc.Unbound <= doc.Findings, qt.IsTrue)
		for rule, count := range doc.ByRule {
			c.Assert(rule, qt.Equals, "policy.banned-phrases")
			total += count
		}
	}
	c.Assert(total > 0, qt.IsTrue)
	bound := 0
	for _, unit := range findings.Units {
		bound += len(unit.Findings)
	}
	c.Assert(bound > 0, qt.IsTrue)
	conflict := bytes.Replace(manifest, []byte(`"partition": "development"`), []byte(`"partition": "final_test"`), 1)
	researchCommand(t, binary, []string{"plan"}, conflict, 2)
	corrupt := bytes.Replace(output, []byte(`"unlabeled_candidates"`), []byte(`"human_labeled"`), 1)
	researchCommand(t, binary, []string{"verify", "--root", root}, corrupt, 2)
}

type corpusExpectedUnit struct {
	ID       string          `json:"id"`
	Source   string          `json:"source"`
	Kind     string          `json:"kind"`
	Role     string          `json:"role"`
	Text     string          `json:"text"`
	Context  string          `json:"context"`
	Segments []document.Span `json:"segments"`
}
