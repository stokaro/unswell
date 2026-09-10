package probability_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/probability"
)

const extractionHash = "3ba0e2ab9e5d2b7dbb0e2d61e2f4a1a3f3ec0a04bfc3f5c6d9e6a17d2c5b8f04"

// packFile builds a syntactically valid experimental pack over real columns.
// Its numerical parameters are fixtures; they carry no editorial meaning.
func packFile(c *qt.C, kind string, ids ...string) probability.File {
	c.Helper()
	file := probability.File{
		Version: probability.Version, ID: "fixture-pack", DeclaredStatus: "experimental", HumanCorpus: "not_qualified",
		Task: probability.Task, Rubric: "fixture-rubric-v1", Kind: kind, Contract: packContract(c, kind, ids),
		Limits: probability.Limits{MinWords: 5}, Estimator: "logistic",
		Logistic:    &probability.Logistic{Means: []float64{10, 0.5}, Scales: []float64{5, 0.25}, Weights: []float64{0.5, -0.25}, Intercept: 0.1},
		Calibration: probability.Calibration{Algorithm: "isotonic", Scores: []float64{-1, 0, 1}, Responses: []float64{0.1, 0.4, 0.9}},
	}
	return sealed(c, file)
}

func packContract(c *qt.C, kind string, ids []string) probability.Contract {
	c.Helper()
	catalog, err := feature.UnitCatalog(kind)
	c.Assert(err, qt.IsNil)
	columns := make([]feature.Descriptor, 0, len(ids))
	for _, id := range ids {
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == id })
		c.Assert(index >= 0, qt.IsTrue)
		columns = append(columns, catalog[index])
	}
	preparation, err := nlp.PreparationHash(extractionHash, false, false)
	c.Assert(err, qt.IsNil)
	return probability.Contract{FeatureContract: feature.UnitContract, UnitContract: nlp.UnitContract, Columns: columns,
		ColumnsSHA256: columnDigest(c, columns), NLP: providerIdentity(c), Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences},
		PreparationHash: preparation}
}

// sealed recomputes the digest so a test edit stays loadable unless it targets
// the digest itself.
func sealed(c *qt.C, file probability.File) probability.File {
	c.Helper()
	digest, err := probability.Digest(file)
	c.Assert(err, qt.IsNil)
	file.SHA256 = digest
	return file
}

func columnDigest(c *qt.C, columns []feature.Descriptor) string {
	c.Helper()
	encoded, err := json.Marshal(columns)
	c.Assert(err, qt.IsNil)
	return fmt.Sprintf("%x", sha256.Sum256(encoded))
}

func providerIdentity(c *qt.C) nlp.Identity {
	c.Helper()
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	return provider.Identity()
}

func encoded(c *qt.C, value any) []byte {
	c.Helper()
	data, err := json.Marshal(value)
	c.Assert(err, qt.IsNil)
	return data
}

func loaded(c *qt.C, file probability.File) *probability.Pack {
	c.Helper()
	pack, err := probability.Load(c.Context(), encoded(c, file))
	c.Assert(err, qt.IsNil)
	return pack
}

func matchingRun(c *qt.C, file probability.File) probability.Run {
	c.Helper()
	return probability.Run{FeatureContract: file.Contract.FeatureContract, UnitContract: file.Contract.UnitContract,
		NLP: file.Contract.NLP, Capabilities: slices.Clone(file.Contract.Capabilities),
		PreparationHash: file.Contract.PreparationHash}
}

func number(value float64) *float64 { return &value }

func measured(c *qt.C, file probability.File, numbers ...*float64) []feature.Value {
	c.Helper()
	values := make([]feature.Value, 0, len(numbers))
	for i, value := range numbers {
		column := file.Contract.Columns[i]
		reason := ""
		if value == nil {
			reason = "no_prose_words"
		}
		values = append(values, feature.Value{ID: column.ID, Version: column.Version, Unit: column.Unit, Number: value, Reason: reason})
	}
	return values
}

func TestLoadKeepsDeclarationsAndDetachesParameters(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "sentence", "prose-words", "type-token-ratio")
	pack := loaded(c, file)
	c.Assert(pack.Kind(), qt.Equals, "sentence")
	c.Assert(pack.Accepted(), qt.IsFalse)
	c.Assert(pack.SHA256(), qt.Equals, file.SHA256)
	c.Assert(pack.File(), qt.DeepEquals, file)
	c.Assert(pack.Columns(), qt.DeepEquals, file.Contract.Columns)
	changed := pack.File()
	changed.Logistic.Weights[0] = 99
	changed.Contract.Columns[0].ID = "changed"
	changed.Calibration.Responses[0] = 0.99
	pack.Columns()[0].Requires[0] = "changed"
	c.Assert(pack.File(), qt.DeepEquals, file)
	c.Assert(pack.Columns()[0].Requires, qt.DeepEquals, file.Contract.Columns[0].Requires)
}

func TestLoadRequiresAnAcceptedPackToNameItsEvidence(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "paragraph", "prose-words", "type-token-ratio")
	file.DeclaredStatus, file.HumanCorpus, file.Evaluation = "accepted", "qualified", "fixture evaluation record"
	pack := loaded(c, sealed(c, file))
	c.Assert(pack.Accepted(), qt.IsTrue)
	c.Assert(pack.Kind(), qt.Equals, "paragraph")
	for _, edit := range []func(*probability.File){
		func(f *probability.File) { f.HumanCorpus = "not_qualified" },
		func(f *probability.File) { f.Evaluation = "" },
		func(f *probability.File) { f.Evaluation = strings.Repeat("x", 257) },
	} {
		changed := file
		edit(&changed)
		result, err := probability.Load(c.Context(), encoded(c, sealed(c, changed)))
		c.Assert(err, qt.ErrorMatches, ".*accepted probability pack requires.*")
		c.Assert(result, qt.IsNil)
	}
}

func TestPackSupportsConcurrentEstimates(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "sentence", "prose-words", "type-token-ratio")
	pack := loaded(c, file)
	unit := probability.Unit{Kind: "sentence", Words: 12, Values: measured(c, file, number(12), number(0.6))}
	var group sync.WaitGroup
	results := make([]probability.Estimate, 8)
	for i := range results {
		group.Go(func() {
			estimate, err := pack.Estimate(context.Background(), unit)
			c.Check(err, qt.IsNil)
			results[i] = estimate
		})
	}
	group.Wait()
	for _, estimate := range results {
		c.Assert(estimate.Status, qt.Equals, probability.StatusAvailable)
		c.Assert(math.Abs(*estimate.Probability-0.5) < 1e-12, qt.IsTrue)
	}
}
