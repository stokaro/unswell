package probability_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/probability"
)

func TestLoadRejectsInconsistentDeclarations(t *testing.T) {
	for _, row := range []struct {
		name, message string
		edit          func(*probability.File)
	}{
		{"version", ".*unsupported probability pack version.*", func(f *probability.File) { f.Version = "unswell-probability-pack-v0" }},
		{"task", ".*unsupported probability pack task.*", func(f *probability.File) { f.Task = "origin" }},
		{"kind", ".*unsupported probability pack unit kind.*", func(f *probability.File) { f.Kind = "fragment" }},
		{"identifier", ".*printable ID and rubric.*", func(f *probability.File) { f.ID = "" }},
		{"control character", ".*printable ID and rubric.*", func(f *probability.File) { f.ID = "pack\x00" }},
		{"rubric", ".*printable ID and rubric.*", func(f *probability.File) { f.Rubric = strings.Repeat("r", 129) }},
		{"status", ".*experimental or accepted.*", func(f *probability.File) { f.DeclaredStatus = "draft" }},
		{"corpus", ".*not_qualified or qualified.*", func(f *probability.File) { f.HumanCorpus = "unknown" }},
		{"absent minimum", ".*minimum word count.*", func(f *probability.File) { f.Limits.MinWords = 0 }},
		{"unbounded minimum", ".*minimum word count.*", func(f *probability.File) { f.Limits.MinWords = 10001 }},
		{"estimator", ".*logistic estimator.*", func(f *probability.File) { f.Estimator = "forest" }},
		{"absent parameters", ".*logistic estimator.*", func(f *probability.File) { f.Logistic = nil }},
		{"width", ".*must match its 2 columns.*", func(f *probability.File) { f.Logistic.Weights = []float64{1} }},
		{"calibration algorithm", ".*isotonic calibration.*", func(f *probability.File) { f.Calibration.Algorithm = "platt" }},
		{"calibration knots", ".*at least 2 matching knots.*", func(f *probability.File) {
			f.Calibration.Scores, f.Calibration.Responses = []float64{0}, []float64{0.5}
		}},
		{"calibration responses", ".*at least 2 matching knots.*", func(f *probability.File) {
			f.Calibration.Responses = []float64{0.1, 0.4}
		}},
		{"calibration order", ".*increasing scores.*", func(f *probability.File) {
			f.Calibration.Responses = []float64{0.9, 0.4, 0.1}
		}},
		{"scale", ".*positive scale.*", func(f *probability.File) { f.Logistic.Scales[0] = 0 }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			file := packFile(c, "sentence", "prose-words", "type-token-ratio")
			row.edit(&file)
			result, err := probability.Load(c.Context(), encoded(c, sealed(c, file)))
			c.Assert(err, qt.ErrorMatches, row.message)
			c.Assert(result, qt.IsNil)
		})
	}
}

func TestLoadRejectsIncompatibleMeasurementContracts(t *testing.T) {
	for _, row := range []struct {
		name, message string
		edit          func(*probability.File)
		resealColumns bool
	}{
		{"feature contract", ".*requires feature contract.*", func(f *probability.File) {
			f.Contract.FeatureContract = "other-v1"
		}, false},
		{"unit contract", ".*requires feature contract.*", func(f *probability.File) {
			f.Contract.UnitContract = "other-v1"
		}, false},
		{"preparation", ".*preparation hash and a complete NLP identity.*", func(f *probability.File) {
			f.Contract.PreparationHash = "0011"
		}, false},
		{"provider", ".*preparation hash and a complete NLP identity.*", func(f *probability.File) {
			f.Contract.NLP.Name = ""
		}, false},
		{"unknown column", ".*does not match this build's sentence measurement.*", func(f *probability.File) {
			f.Contract.Columns[1].ID = "invented-ratio"
		}, true},
		{"altered column", ".*does not match this build's sentence measurement.*", func(f *probability.File) {
			f.Contract.Columns[0].Version = "9"
		}, true},
		{"repeated column", ".*repeats column.*", func(f *probability.File) {
			f.Contract.Columns[1] = f.Contract.Columns[0]
		}, true},
		{"column digest", ".*column digest does not cover.*", func(f *probability.File) {
			f.Contract.ColumnsSHA256 = strings.Repeat("0", 64)
		}, false},
		{"absent capabilities", ".*requires the capabilities its columns need.*", func(f *probability.File) {
			f.Contract.Capabilities = []nlp.Capability{}
		}, false},
		{"repeated capability", ".*is unknown or repeated.*", func(f *probability.File) {
			f.Contract.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Tokens}
		}, false},
		{"unknown capability", ".*is unknown or repeated.*", func(f *probability.File) {
			f.Contract.Capabilities = []nlp.Capability{nlp.Tokens, "entities"}
		}, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			file := packFile(c, "sentence", "prose-words", "type-token-ratio")
			row.edit(&file)
			if row.resealColumns {
				file.Contract.ColumnsSHA256 = columnDigest(c, file.Contract.Columns)
			}
			result, err := probability.Load(c.Context(), encoded(c, sealed(c, file)))
			c.Assert(err, qt.ErrorMatches, row.message)
			c.Assert(result, qt.IsNil)
		})
	}
}

func TestLoadRequiresDeclaredCapabilitiesForEveryColumn(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "sentence", "prose-words", "noun-token-ratio")
	result, err := probability.Load(c.Context(), encoded(c, file))
	c.Assert(err, qt.ErrorMatches, `.*column "noun-token-ratio" requires the pos capability.*`)
	c.Assert(result, qt.IsNil)
	file.Contract.Capabilities = append(file.Contract.Capabilities, nlp.POS)
	c.Assert(loaded(c, sealed(c, file)).Columns(), qt.HasLen, 2)
}

func TestLoadRejectsMalformedBytes(t *testing.T) {
	c := qt.New(t)
	valid := string(encoded(c, packFile(c, "sentence", "prose-words", "type-token-ratio")))
	for _, row := range []struct{ name, data, message string }{
		{"empty", "", ".*must contain 1 to .* bytes"},
		{"oversized", strings.Repeat(" ", probability.MaxBytes+1), ".*must contain 1 to .* bytes"},
		{"null", strings.Replace(valid, `"fixture-rubric-v1"`, "null", 1), ".*cannot contain null.*"},
		{"duplicate key", strings.Replace(valid, `"task":`, `"id":"other","task":`, 1), ".*duplicate or invalid.*"},
		{"unknown field", strings.Replace(valid, `"task":`, `"unknown":1,"task":`, 1), ".*unknown field.*"},
		{"two objects", valid + valid, ".*exactly one JSON object.*"},
		{"digest", strings.Replace(valid, `"declared_status":"experimental"`, `"declared_status":"accepted"`, 1),
			".*(accepted probability pack requires|digest does not cover).*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result, err := probability.Load(c.Context(), []byte(row.data))
			c.Assert(err, qt.ErrorMatches, row.message)
			c.Assert(result, qt.IsNil)
		})
	}
}

// The decoder's own wording for unfinished input differs between supported Go
// releases, so truncation is checked as a failure rather than as a message.
func TestLoadRejectsTruncatedBytes(t *testing.T) {
	c := qt.New(t)
	valid := encoded(c, packFile(c, "sentence", "prose-words", "type-token-ratio"))
	result, err := probability.Load(c.Context(), valid[:len(valid)-1])
	c.Assert(err, qt.IsNotNil)
	c.Assert(result, qt.IsNil)
}

func TestLoadRejectsAlteredDigestAndCancellation(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "sentence", "prose-words", "type-token-ratio")
	file.Limits.MinWords++
	result, err := probability.Load(c.Context(), encoded(c, file))
	c.Assert(err, qt.ErrorMatches, ".*digest does not cover its contents.*")
	c.Assert(result, qt.IsNil)
	ctx, cancel := context.WithCancel(c.Context())
	cancel()
	result, err = probability.Load(ctx, encoded(c, packFile(c, "sentence", "prose-words", "type-token-ratio")))
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.IsNil)
}
