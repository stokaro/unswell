package training_test

import (
	"encoding/json"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/training"
)

// cohortBank seeds a historical, a contemporary, and a mixed reference cohort
// from the first training unit of each cohort.
func cohortBank(c *qt.C, artifact corpus.Artifact, decisions annotation.DecisionSet,
	files map[string][]byte,
) corpus.CompressionBank {
	c.Helper()
	seeds := map[string]string{}
	for _, target := range artifact.Units {
		if target.Partition != "training" || target.Unit.Kind != "paragraph" {
			continue
		}
		if _, taken := seeds[target.Cohort]; !taken {
			seeds[target.Cohort] = target.Unit.ID
		}
	}
	c.Assert(seeds, qt.HasLen, 2)
	options := corpus.CompressionBankOptions{Version: corpus.CompressionBankVersion, Kind: "paragraph",
		Compression: feature.CompressionOptions{Level: 6, MaxInputBytes: 1 << 20}, Cohorts: []corpus.CompressionCohort{
			{ID: "historical", Origin: "historical", UnitIDs: []string{seeds["historical-2012"]}},
			{ID: "contemporary", Origin: "contemporary", UnitIDs: []string{seeds["contemporary"]}},
			{ID: "mixed", Origin: "mixed", UnitIDs: []string{seeds["contemporary"], seeds["historical-2012"]}},
		}}
	bank, err := corpus.BuildCompressionBankDecisions(c.Context(), artifact, decisions, files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(bank.ReservedGroups, qt.HasLen, 2)
	return bank
}

// The compression baseline measures every target against the bank's cohorts,
// excludes the reserved groups, joins prepared features when asked, restores,
// and predicts with the same bank; every other use of a bank is refused.
func TestRunCompressionDecisionsFitsReferenceColumns(t *testing.T) {
	c := qt.New(t)
	artifact, files := cohortFixture(c)
	decisions, err := corpus.CohortDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	bank := cohortBank(c, artifact, decisions, files)
	options := fittingOptions()
	options.AllowSimulation, options.Features, options.MissingFeatures = false, nil, "exclude"

	fitted, err := training.RunCompressionDecisions(c.Context(), artifact, decisions, files, options, bank)
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Identity.FeatureSource, qt.Equals, "compression_bank")
	c.Assert(fitted.Identity.CompressionBankHash, qt.Equals, bank.SHA256)
	c.Assert(fitted.Identity.FeatureContract, qt.Equals, feature.CompressionContract)
	c.Assert(fitted.Options.Features, qt.DeepEquals, []string{
		"compression.incremental-bytes/contemporary", "compression.incremental-bytes/historical",
		"compression.incremental-bytes/mixed", "compression.reference-gain/contemporary",
		"compression.reference-gain/historical", "compression.reference-gain/mixed"})
	c.Assert(fitted.Options.Reservation, qt.DeepEquals, training.BankReservation(bank))
	c.Assert(fitted.Partitions[0].Excluded["reserved_reference_group"], qt.Equals, 4)
	c.Assert(len(fitted.Partitions[0].Rows) > 0, qt.IsTrue)
	for _, partition := range fitted.Partitions {
		for _, row := range partition.Rows {
			for _, group := range bank.ReservedGroups {
				c.Assert(row.GroupID, qt.Not(qt.Equals), group.ID)
			}
		}
	}

	encoded, err := json.Marshal(fitted)
	c.Assert(err, qt.IsNil)
	restored, err := training.Load(c.Context(), encoded)
	c.Assert(err, qt.IsNil)
	c.Assert(restored.SHA256, qt.Equals, fitted.SHA256)

	threshold := 0.5
	plan := training.PredictionPlan{Version: training.PredictionVersion, ID: "compression-trial", ProtocolSHA256: protocolDigest(),
		ModelSHA256: fitted.SHA256, CorpusSHA256: artifact.SHA256, Partition: "final_test",
		Context: "prepared_piece", Response: "isotonic", Threshold: &threshold}
	predictions, err := training.PredictWith(c.Context(), artifact, files, fitted, plan, training.PredictionResources{Bank: &bank})
	c.Assert(err, qt.IsNil)
	c.Assert(len(predictions.Rows) > 0, qt.IsTrue)
	available := 0
	for _, row := range predictions.Rows {
		if row.Status == "available" {
			available++
			continue
		}
		c.Assert(row.Reason, qt.Equals, "calibration/out_of_range")
	}
	c.Assert(available > 0, qt.IsTrue)
	_, err = training.Predict(c.Context(), artifact, files, fitted, plan, nil)
	c.Assert(err, qt.ErrorMatches, "compression prediction requires the fitted reference bank")

	baseline := fittingOptions()
	baseline.AllowSimulation, baseline.Reservation = false, training.BankReservation(bank)
	reserved, err := training.RunDecisions(c.Context(), artifact, decisions, files, baseline)
	c.Assert(err, qt.IsNil)
	c.Assert(reserved.Partitions[0].Excluded["reserved_reference_group"], qt.Equals, 4)
	c.Assert(reserved.Identity.CompressionBankHash, qt.Equals, "")
	plan.ModelSHA256, plan.Context = reserved.SHA256, "prepared_piece"
	_, err = training.PredictWith(c.Context(), artifact, files, reserved, plan, training.PredictionResources{Bank: &bank})
	c.Assert(err, qt.ErrorMatches, "only compression prediction takes a reference bank")
}

// A joint fit carries the prepared features beside the reference columns, in
// one sorted column order, on the same rows as the reference columns alone.
func TestRunCompressionDecisionsJoinsPreparedFeatures(t *testing.T) {
	c := qt.New(t)
	artifact, files := cohortFixture(c)
	decisions, err := corpus.CohortDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	bank := cohortBank(c, artifact, decisions, files)
	options := fittingOptions()
	options.AllowSimulation, options.Features, options.MissingFeatures = false, []string{"prose-words"}, "exclude"
	joint, err := training.RunCompressionDecisions(c.Context(), artifact, decisions, files, options, bank)
	c.Assert(err, qt.IsNil)
	c.Assert(joint.Options.Features, qt.HasLen, 7)
	c.Assert(joint.Options.Features[6], qt.Equals, "prose-words")
	c.Assert(joint.Identity.Columns[6].ID, qt.Equals, "prose-words")
	c.Assert(joint.Identity.FeatureSource, qt.Equals, "compression_bank")
	c.Assert(joint.Identity.FeatureContract, qt.Not(qt.Equals), feature.CompressionContract)
	c.Assert(joint.Partitions[0].Excluded["reserved_reference_group"], qt.Equals, 4)
	encoded, err := json.Marshal(joint)
	c.Assert(err, qt.IsNil)
	_, err = training.Load(c.Context(), encoded)
	c.Assert(err, qt.IsNil)
	threshold := 0.5
	plan := training.PredictionPlan{Version: training.PredictionVersion, ID: "joint-trial", ProtocolSHA256: protocolDigest(),
		ModelSHA256: joint.SHA256, CorpusSHA256: artifact.SHA256, Partition: "final_test",
		Context: "prepared_piece", Response: "isotonic", Threshold: &threshold}
	predictions, err := training.PredictWith(c.Context(), artifact, files, joint, plan, training.PredictionResources{Bank: &bank})
	c.Assert(err, qt.IsNil)
	c.Assert(len(predictions.Rows) > 0, qt.IsTrue)
}

// A compression fit refuses the zero policy, activation columns, a bank of
// another kind or corpus, and a reservation other than the bank's.
func TestRunCompressionDecisionsRefusesOtherBanks(t *testing.T) {
	c := qt.New(t)
	artifact, files := cohortFixture(c)
	decisions, err := corpus.CohortDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	bank := cohortBank(c, artifact, decisions, files)
	options := fittingOptions()
	options.AllowSimulation, options.Features, options.MissingFeatures = false, nil, "exclude"
	for _, row := range []struct {
		name    string
		change  func(*training.Options, *corpus.CompressionBank)
		message string
	}{
		{"zero-policy", func(o *training.Options, _ *corpus.CompressionBank) { o.MissingFeatures = "zero" },
			"the zero missing-feature policy applies to rule activations only"},
		{"activation", func(o *training.Options, _ *corpus.CompressionBank) {
			o.Features = []string{"activation/hype.vague-praise"}
		},
			"compression training adds reference columns to prepared features only"},
		{"other-kind", func(_ *training.Options, b *corpus.CompressionBank) { b.Options.Kind = "sentence" },
			"compression training requires a sealed bank.*"},
		{"other-corpus", func(_ *training.Options, b *corpus.CompressionBank) {
			b.Verification.ArtifactSHA256 = strings.Repeat("0", 64)
		}, "compression training requires a sealed bank.*"},
		{"other-reservation", func(o *training.Options, _ *corpus.CompressionBank) {
			o.Reservation = &training.Reservation{BankSHA256: strings.Repeat("1", 64), Groups: []string{"none"}}
		}, "compression training reserves the bank's groups and no others"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			changed, other := options, bank
			row.change(&changed, &other)
			_, err := training.RunCompressionDecisions(c.Context(), artifact, decisions, files, changed, other)
			c.Assert(err, qt.ErrorMatches, row.message)
		})
	}
}
