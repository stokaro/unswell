package baseline_test

import (
	"bytes"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/baseline"
)

func TestLoadRejectsUntrustedArtifacts(t *testing.T) {
	c := qt.New(t)
	file, err := baseline.Create(t.Context(), sampleSnapshot())
	c.Assert(err, qt.IsNil)
	encoded, err := baseline.Encode(t.Context(), file)
	c.Assert(err, qt.IsNil)
	for _, change := range []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"null", []byte("null")},
		{"array", []byte("[]")},
		{"trailing", append(bytes.Clone(encoded), []byte("{}")...)},
		{"unknown-field", bytes.Replace(encoded, []byte(`"version":`), []byte(`"extra": 1, "version":`), 1)},
		{"duplicate-field", bytes.Replace(encoded, []byte(`"version":`), []byte(`"version": "wrong", "version":`), 1)},
		{"nested-duplicate", bytes.Replace(encoded, []byte(`"content_hash":`), []byte(`"content_hash": "wrong", "content_hash":`), 1)},
		{"bad-version", bytes.ReplaceAll(encoded, []byte(baseline.Version), []byte("unswell-baseline-v0"))},
		{"bad-fingerprint-version", bytes.ReplaceAll(encoded, []byte(baseline.FingerprintVersion), []byte("unknown"))},
		{"bad-key", bytes.ReplaceAll(encoded, []byte(file.Entries[0].Fingerprint), []byte(hash("tampered")))},
		{"missing-acceptance", bytes.ReplaceAll(encoded, []byte(`"accepted"`), []byte(`""`))},
		{"null-acceptance", bytes.ReplaceAll(encoded, []byte(`"accepted"`), []byte(`null`))},
		{"oversized", bytes.Repeat([]byte{' '}, baseline.MaxBytes+1)},
		{"too-deep", []byte(strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34))},
	} {
		t.Run(change.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := baseline.Load(t.Context(), change.data)
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestInvalidFilesAndCoverage(t *testing.T) {
	for _, change := range []struct {
		name string
		edit func(*baseline.File)
	}{
		{"duplicate-entry", func(f *baseline.File) { f.Entries = append(f.Entries, f.Entries[0]) }},
		{"duplicate-document", func(f *baseline.File) { f.Documents = append(f.Documents, f.Documents[0]) }},
		{"missing-document", func(f *baseline.File) { f.Documents[0].Path = "other.md" }},
		{"missing-entries", func(f *baseline.File) { f.Entries = nil }},
		{"missing-compatibility", func(f *baseline.File) { f.Compatibility.ModelHash = "" }},
		{"entry-limit", func(f *baseline.File) { f.Entries = make([]baseline.Entry, baseline.MaxEntries+1) }},
		{"document-limit", func(f *baseline.File) { f.Documents = make([]baseline.Document, baseline.MaxDocuments+1) }},
	} {
		t.Run(change.name, func(t *testing.T) {
			c := qt.New(t)
			file, err := baseline.Create(t.Context(), sampleSnapshot())
			c.Assert(err, qt.IsNil)
			change.edit(&file)
			_, err = baseline.Encode(t.Context(), file)
			c.Assert(err, qt.IsNotNil)
		})
	}
	c := qt.New(t)
	snapshot := sampleSnapshot()
	snapshot.Candidates = nil
	file, err := baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNil)
	c.Assert(file.Entries, qt.HasLen, 0)
	_, err = baseline.Encode(t.Context(), file)
	c.Assert(err, qt.IsNil)
	snapshot.Documents = nil
	_, err = baseline.Create(t.Context(), snapshot)
	c.Assert(err, qt.IsNotNil)
}

func FuzzLoad(f *testing.F) {
	file, err := baseline.Create(f.Context(), sampleSnapshot())
	if err != nil {
		f.Fatal(err)
	}
	data, err := baseline.Encode(f.Context(), file)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(data)
	f.Add([]byte(`{"version":null,"version":"duplicate"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		file, err := baseline.Load(t.Context(), data)
		if err != nil {
			return
		}
		c := qt.New(t)
		encoded, err := baseline.Encode(t.Context(), file)
		c.Assert(err, qt.IsNil)
		roundTrip, err := baseline.Load(t.Context(), encoded)
		c.Assert(err, qt.IsNil)
		c.Assert(roundTrip, qt.DeepEquals, file)
	})
}
