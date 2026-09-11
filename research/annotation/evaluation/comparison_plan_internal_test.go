package evaluation

// White-box tests: Vary individual identities at the private comparison boundary;
// RunComparison validates complete hashed prediction artifacts before reaching this projection.

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation/training"
)

func comparisonInputs() (ComparisonPlan, training.Predictions, training.Predictions) {
	plan := ComparisonPlan{Version: ComparisonVersion, ID: "frozen", ProtocolSHA256: strings.Repeat("a", 64),
		CandidateSHA256: strings.Repeat("b", 64), ComparatorSHA256: strings.Repeat("c", 64)}
	a := training.Predictions{SHA256: plan.CandidateSHA256, Plan: training.PredictionPlan{
		ProtocolSHA256: plan.ProtocolSHA256, CorpusSHA256: strings.Repeat("d", 64), Partition: "final_test", Context: "prepared_piece"}}
	b := a
	b.SHA256 = plan.ComparatorSHA256
	return plan, a, b
}

func TestComparisonRejectsUnmatchedScopes(t *testing.T) {
	for _, row := range []struct {
		name   string
		change func(*training.Predictions)
	}{
		{"selection", func(p *training.Predictions) { p.SHA256 = strings.Repeat("0", 64) }},
		{"protocol", func(p *training.Predictions) { p.Plan.ProtocolSHA256 = strings.Repeat("0", 64) }},
		{"corpus", func(p *training.Predictions) { p.Plan.CorpusSHA256 = strings.Repeat("0", 64) }},
		{"partition", func(p *training.Predictions) { p.Plan.Partition = "development" }},
		{"context", func(p *training.Predictions) { p.Plan.Context = "source_document" }},
		{"kind", func(p *training.Predictions) { p.Model.Identity.Kind = "sentence" }},
		{"profile", func(p *training.Predictions) { p.Model.Identity.ProfileSHA256 = strings.Repeat("0", 64) }},
		{"NLP", func(p *training.Predictions) { p.Model.Identity.NLP.Version = "different" }},
		{"policy", func(p *training.Predictions) { p.Model.Identity.ExtractionPolicyHash = "different" }},
		{"preparation", func(p *training.Predictions) { p.Model.Identity.PreparationHash = "different" }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			plan, a, b := comparisonInputs()
			_, err := comparisonPlanIdentity(t.Context(), plan, a, b)
			c.Assert(err, qt.IsNil)
			for _, source := range []string{"", "lexical_ngrams", "compression_bank", "llmdet_tables"} {
				left, right := a, b
				left.Model.Identity.FeatureContract, right.Model.Identity.FeatureContract = feature.UnitContract, feature.UnitContract
				switch source {
				case "lexical_ngrams":
					right.Model.Identity.FeatureSource = source
					right.Model.Identity.FeatureContract = feature.LexicalCountContract
					right.Model.Identity.LexicalVocabularyHash = strings.Repeat("f", 64)
				case "compression_bank":
					right.Model.Identity.FeatureSource, right.Model.Identity.Context = source, "prepared_piece"
					right.Model.Identity.FeatureContract = feature.CompressionContract
					right.Model.Identity.CompressionBankHash = strings.Repeat("f", 64)
				case "llmdet_tables":
					right.Model.Identity.FeatureSource, right.Model.Identity.Context = source, "prepared_piece"
					right.Model.Identity.FeatureContract = training.LLMDetContract
					right.Model.Identity.LLMDetPackHash = strings.Repeat("f", 64)
				}
				_, err := comparisonPlanIdentity(t.Context(), plan, left, right)
				c.Assert(err, qt.IsNil)
				row.change(&right)
				hash, err := comparisonPlanIdentity(t.Context(), plan, left, right)
				c.Assert(err, qt.IsNotNil)
				c.Assert(hash, qt.Equals, "")
			}
		})
	}
}

// Two trials compare only under one reservation, because the reserved groups
// define the rows both fits saw.
func TestComparisonRequiresOneReservation(t *testing.T) {
	c := qt.New(t)
	plan, a, b := comparisonInputs()
	reservation := &training.Reservation{BankSHA256: strings.Repeat("a", 64), Groups: []string{"g1"}}
	a.Model.Options.Reservation = reservation
	_, err := comparisonPlanIdentity(t.Context(), plan, a, b)
	c.Assert(err, qt.ErrorMatches, "compared trials require the same reservation")
	b.Model.Options.Reservation = &training.Reservation{BankSHA256: strings.Repeat("a", 64), Groups: []string{"g1"}}
	_, err = comparisonPlanIdentity(t.Context(), plan, a, b)
	c.Assert(err, qt.IsNil)
}

func TestComparisonAllowsDifferentModelFeatures(t *testing.T) {
	c := qt.New(t)
	plan, a, b := comparisonInputs()
	b.Model.Identity.ColumnsSHA256 = strings.Repeat("f", 64)
	b.Model.Identity.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS}
	b.Plan.Response = "isotonic"
	_, err := comparisonPlanIdentity(t.Context(), plan, a, b)
	c.Assert(err, qt.IsNil)
}
