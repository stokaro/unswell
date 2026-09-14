package probability_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/probability"
)

// lexicalPack builds a pack whose columns are n-gram keys rather than named
// catalog features. The keys are fixtures and carry no editorial meaning.
func lexicalPack(c *qt.C, terms ...string) probability.File {
	c.Helper()
	options := feature.LexicalOptions{WordMin: 1, WordMax: 2, CharMin: 3, CharMax: 5}
	columns := make([]feature.Descriptor, len(terms))
	for i, term := range terms {
		columns[i] = feature.Descriptor{ID: feature.LexicalColumnID(term), Version: "1", Family: "lexical-ngram",
			Type: "number", Unit: "log1p-count", Scope: "paragraph",
			Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences}}
	}
	preparation, err := nlp.PreparationHash(extractionHash, false, false)
	c.Assert(err, qt.IsNil)
	file := probability.File{
		Version: probability.Version, ID: "fixture-lexical", DeclaredStatus: "experimental",
		HumanCorpus: "not_qualified", Task: probability.TaskOrigin, Rubric: "fixture-rubric-v1", Kind: "paragraph",
		Contract: probability.Contract{FeatureContract: feature.LexicalCountContract, UnitContract: nlp.UnitContract,
			Columns: columns, ColumnsSHA256: columnDigest(c, columns), NLP: providerIdentity(c),
			Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences}, PreparationHash: preparation},
		Limits: probability.Limits{MinWords: 25, MaxWords: 89}, Estimator: "logistic",
		Logistic: &probability.Logistic{Means: make([]float64, len(terms)), Scales: onesOf(len(terms)),
			Weights: onesOf(len(terms)), Intercept: 0.1},
		Calibration: probability.Calibration{Algorithm: "isotonic", Scores: []float64{-1, 0, 1},
			Responses: []float64{0.1, 0.4, 0.9}},
		Vocabulary: &probability.Vocabulary{Options: options, Terms: terms},
	}
	return sealed(c, file)
}

func onesOf(n int) []float64 {
	values := make([]float64, n)
	for i := range values {
		values[i] = 1
	}
	return values
}

func TestLexicalPackLoadsWithItsVocabulary(t *testing.T) {
	c := qt.New(t)
	pack, err := probability.Load(c.Context(), encoded(c, lexicalPack(c, "c:ati", "c:ent")))
	c.Assert(err, qt.IsNil)
	vocabulary := pack.Vocabulary()
	c.Assert(vocabulary, qt.IsNotNil)
	c.Assert(vocabulary.Terms, qt.DeepEquals, []string{"c:ati", "c:ent"})
	c.Assert(pack.Columns(), qt.HasLen, 2)
}

// A pack of named features that also carried keys, or a pack of keys that
// carried none, would leave a caller guessing how to measure it.
func TestPackRefusesAMismatchedVocabulary(t *testing.T) {
	for _, row := range []struct {
		name    string
		message string
		edit    func(*probability.File)
	}{
		{"named features with a vocabulary", ".*carries no vocabulary.*", func(f *probability.File) {
			f.Contract.FeatureContract = feature.UnitContract
		}},
		{"keys without a vocabulary", ".*requires its vocabulary.*", func(f *probability.File) {
			f.Vocabulary = nil
		}},
		{"a term the columns do not name", ".*does not name its vocabulary term.*", func(f *probability.File) {
			f.Vocabulary.Terms[0] = "c:xyz"
		}},
		{"fewer terms than columns", ".*terms for 2 columns.*", func(f *probability.File) {
			f.Vocabulary.Terms = f.Vocabulary.Terms[:1]
		}},
		{"a key the counter cannot produce", ".*not a countable n-gram key.*", func(f *probability.File) {
			f.Vocabulary.Options.CharMin, f.Vocabulary.Options.CharMax = 4, 4
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			file := lexicalPack(c, "c:ati", "c:ent")
			row.edit(&file)
			_, err := probability.Load(c.Context(), encoded(c, sealed(c, file)))
			c.Assert(err, qt.ErrorMatches, row.message)
		})
	}
}

// A model fitted inside a word band says nothing above it, so the pack abstains
// rather than extrapolating.
func TestPackAbstainsOutsideItsWordBand(t *testing.T) {
	c := qt.New(t)
	pack, err := probability.Load(c.Context(), encoded(c, lexicalPack(c, "c:ati", "c:ent")))
	c.Assert(err, qt.IsNil)
	values := []feature.Value{numberValue(pack.Columns()[0], 0), numberValue(pack.Columns()[1], 0)}
	for _, row := range []struct {
		name   string
		words  int
		status string
	}{
		{"below the floor", 24, probability.StatusInsufficientEvidence},
		{"at the floor", 25, probability.StatusAvailable},
		{"at the ceiling", 89, probability.StatusAvailable},
		{"above the ceiling", 90, probability.StatusInsufficientEvidence},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			estimate, err := pack.Estimate(c.Context(),
				probability.Unit{Kind: "paragraph", Words: row.words, Values: values})
			c.Assert(err, qt.IsNil)
			c.Assert(estimate.Status, qt.Equals, row.status)
		})
	}
}

func numberValue(column feature.Descriptor, number float64) feature.Value {
	return feature.Value{ID: column.ID, Version: column.Version, Unit: column.Unit, Number: &number}
}

func TestPackRefusesAnInvertedWordBand(t *testing.T) {
	c := qt.New(t)
	file := lexicalPack(c, "c:ati", "c:ent")
	file.Limits.MaxWords = 10
	_, err := probability.Load(c.Context(), encoded(c, sealed(c, file)))
	c.Assert(err, qt.ErrorMatches, ".*maximum at or above the minimum.*")
}
