package unswell_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func policyResourceChange(kind string) unswell.PolicyChange {
	return unswell.PolicyChange{Path: "policy.yaml", Kind: kind,
		BeforeHash: strings.Repeat("a", 64), AfterHash: strings.Repeat("b", 64)}
}

func TestTrustedComparisonRejectsInvalidResourceAssertions(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{NoGate: true})
	c.Assert(err, qt.IsNil)
	change := policyResourceChange("config")
	for _, invalid := range []unswell.PolicyChange{
		{Path: "../policy.yaml", Kind: "config", BeforeHash: change.BeforeHash},
		{Path: "policy.yaml", Kind: "source-permissions", BeforeHash: change.BeforeHash},
		{Path: "policy.yaml", Kind: "config", BeforeHash: "not-a-hash"},
		{Path: "policy.yaml", Kind: "config", BeforeHash: change.BeforeHash, AfterHash: change.BeforeHash},
	} {
		result, err := engine.AnalyzeChangedWithOptions(t.Context(), nil, nil,
			unswell.ChangeOptions{TrustedPolicy: true, PolicyChanges: []unswell.PolicyChange{invalid}})
		c.Assert(err, qt.IsNotNil)
		c.Assert(result.Gate.Passed, qt.IsFalse)
		c.Assert(result.Manifest.Complete, qt.IsFalse)
	}
	for _, options := range []unswell.ChangeOptions{
		{PolicyChanges: []unswell.PolicyChange{change}},
		{TrustedPolicy: true, PolicyChanges: []unswell.PolicyChange{change, change}},
		{TrustedPolicy: true, PolicyChanges: make([]unswell.PolicyChange, 257)},
	} {
		_, err := engine.AnalyzeChangedWithOptions(t.Context(), nil, nil, options)
		c.Assert(err, qt.IsNotNil)
	}
	input := changedSource(debtProse)
	want, err := engine.AnalyzeChanged(t.Context(), input, input)
	c.Assert(err, qt.IsNil)
	got, err := engine.AnalyzeChangedWithOptions(t.Context(), input, input, unswell.ChangeOptions{})
	c.Assert(err, qt.IsNil)
	c.Assert(got, qt.DeepEquals, want)
}

func TestTrustedPermissionMustKeepItsTargetAndReason(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy)})
	c.Assert(err, qt.IsNil)
	permission := "<!-- unswell-disable-next-block filler.announced-importance -- Contract wording. -->\n\n"
	before := changedSource(permission + debtProse)
	for _, after := range []string{
		permission + strings.ReplaceAll(debtProse, "may retry", "must not retry"),
		strings.ReplaceAll(permission, "Contract wording", "New rationale") + debtProse,
		"# New context\n\n" + permission + debtProse,
	} {
		result, err := engine.AnalyzeChangedWithOptions(t.Context(), before, changedSource(after), unswell.ChangeOptions{TrustedPolicy: true})
		c.Assert(err, qt.IsNil)
		c.Assert(result.Gate.Passed, qt.IsFalse)
		c.Assert(result.Suppressions[0].TrustState, qt.Equals, "untrusted")
	}
}

func TestTrustedPolicyChangesRescanOldDebtUnderTheExistingPolicy(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy)})
	c.Assert(err, qt.IsNil)
	sources := changedSource(debtProse)
	options := unswell.ChangeOptions{TrustedPolicy: true, PolicyChanges: []unswell.PolicyChange{policyResourceChange("config")}}
	result, err := engine.AnalyzeChangedWithOptions(t.Context(), sources, sources, options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Gate.Reasons, qt.HasLen, 3)
	c.Assert(result.PolicyComparison.Complete, qt.IsTrue)
	c.Assert(result.PolicyComparison.FullScan, qt.IsTrue)
	c.Assert(result.Changes.Documents[0].FullReason, qt.Equals, "policy_changed")
	c.Assert(result.Changes.SelectedUnits, qt.Equals, result.Changes.ComparedUnits)
	options.PolicyChanges[0].Path = "mutated.yaml"
	c.Assert(result.PolicyComparison.Changes[0].Path, qt.Equals, "policy.yaml")
}

func TestTrustedPolicyIgnoresNewPermissionsAndRetainsMatchingOnes(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy)})
	c.Assert(err, qt.IsNil)
	permission := "<!-- unswell-disable-next-block filler.announced-importance -- Required contract wording. -->\n\n"
	before := changedSource("# Retries\n\n" + debtProse)
	after := changedSource("# Retries\n\n" + permission + debtProse)
	options := unswell.ChangeOptions{TrustedPolicy: true}
	result, err := engine.AnalyzeChangedWithOptions(t.Context(), before, after, options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Suppressions, qt.HasLen, 1)
	c.Assert(result.Suppressions[0].TrustState, qt.Equals, "untrusted")
	c.Assert(result.PolicyComparison.Changes[0].Kind, qt.Equals, "source-permissions")
	c.Assert(result.Findings[0].Suppressed, qt.IsFalse)
	c.Assert(result.Findings[0].SuppressionIDs, qt.HasLen, 0)
	c.Assert(result.Assessments[len(result.Assessments)-1].EffectiveSlopScore > 0, qt.IsTrue)
	moved := changedSource("# Retries\n\nA clean introduction.\n\n" + permission + debtProse)
	result, err = engine.AnalyzeChangedWithOptions(t.Context(), after, moved, options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Suppressions[0].TrustState, qt.Equals, "trusted")
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	c.Assert(result.PolicyComparison.Changes, qt.HasLen, 0)
}

func TestTrustedPermissionUnionRequiresEveryContributingPermission(t *testing.T) {
	c := qt.New(t)
	policy := "version: 1\nextends: [builtin:custom]\nrules:\n  repetition.exact-sentence: {enabled: true, gate: forbid}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	permission := "<!-- unswell-disable-next-block repetition.exact-sentence -- Quoted contract wording. -->\n\n"
	prose := "The client opens a new connection after the server closes the previous connection."
	before := changedSource("# First\n\n" + permission + prose + "\n\n# Second\n\n" + permission + prose)
	after := changedSource(string(before[0].Bytes) + "\n\n# Third\n\n" + permission + prose)
	result, err := engine.AnalyzeChangedWithOptions(t.Context(), before, after, unswell.ChangeOptions{TrustedPolicy: true})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Suppressions, qt.HasLen, 3)
	c.Assert(result.Suppressions[0].TrustState, qt.Equals, "trusted")
	c.Assert(result.Suppressions[1].TrustState, qt.Equals, "trusted")
	c.Assert(result.Suppressions[2].TrustState, qt.Equals, "untrusted")
	for _, finding := range result.Findings {
		c.Assert(finding.Suppressed, qt.IsFalse)
		c.Assert(finding.SuppressionIDs, qt.HasLen, 0)
	}
}

func TestTrustedPermissionsCannotInheritAmbiguousTargets(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy)})
	c.Assert(err, qt.IsNil)
	permission := "<!-- unswell-disable-next-block filler.announced-importance -- Quoted wording. -->\n\n"
	before := changedSource(permission + debtProse + "\n\n" + permission + debtProse)
	options := unswell.ChangeOptions{TrustedPolicy: true}
	result, err := engine.AnalyzeChangedWithOptions(t.Context(), before, before, options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	result, err = engine.AnalyzeChangedWithOptions(t.Context(), before,
		changedSource("A clean introduction.\n\n"+string(before[0].Bytes)), options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	for _, permission := range result.Suppressions {
		c.Assert(permission.TrustState, qt.Equals, "ambiguous")
	}
}

func TestTrustedBaselineChangeCannotAcceptNewDebt(t *testing.T) {
	c := qt.New(t)
	before := changedSource("# First\n\n" + debtProse)
	data, _ := captureDebt(t, debtPolicy, before[0])
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	after := changedSource(string(before[0].Bytes) + "\n\n# Second\n\n" + debtProse)
	options := unswell.ChangeOptions{TrustedPolicy: true, PolicyChanges: []unswell.PolicyChange{policyResourceChange("baseline")}}
	result, err := engine.AnalyzeChangedWithOptions(t.Context(), before, after, options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Gate.Accepted, qt.HasLen, 3)
	c.Assert(result.Gate.Reasons, qt.HasLen, 3)
}

func TestTrustedComparisonIsConcurrentAndKeepsOperationalErrors(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), NoGate: true})
	c.Assert(err, qt.IsNil)
	sources := changedSource(debtProse)
	options := unswell.ChangeOptions{TrustedPolicy: true}
	var results [4]unswell.RunResult
	var failures [4]error
	var workers sync.WaitGroup
	for i := range results {
		workers.Go(func() {
			results[i], failures[i] = engine.AnalyzeChangedWithOptions(t.Context(), sources, sources, options)
		})
	}
	workers.Wait()
	for i := range results {
		c.Assert(failures[i], qt.IsNil)
		c.Assert(results[i], qt.DeepEquals, results[0])
	}
	invalid := changedSource(string([]byte{0xff}))
	result, err := engine.AnalyzeChangedWithOptions(t.Context(), invalid, sources, options)
	c.Assert(err, qt.IsNotNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.PolicyComparison.Complete, qt.IsFalse)
	c.Assert(len(result.Findings) > 0, qt.IsTrue)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = engine.AnalyzeChangedWithOptions(ctx, sources, sources, options)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
