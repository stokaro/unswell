// Package testfixture builds simulated research inputs for tests.
package testfixture

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// TrainingInput holds caller-owned source and annotation fixtures.
type TrainingInput struct {
	Manifest corpus.Manifest
	Files    map[string][]byte
	Round    map[string]json.RawMessage
}

// Load reads simulated training fixtures under the given directory.
func Load(t *testing.T, directory string) TrainingInput {
	t.Helper()
	c := qt.New(t)
	fixtureRoot, err := os.OpenRoot(directory)
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Assert(fixtureRoot.Close(), qt.IsNil) })
	data, err := fixtureRoot.ReadFile("manifest.json")
	c.Assert(err, qt.IsNil)
	manifest, err := corpus.LoadManifest(t.Context(), data)
	c.Assert(err, qt.IsNil)
	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil)
	needed, err := corpus.Files(t.Context(), plan)
	c.Assert(err, qt.IsNil)
	root, err := fixtureRoot.OpenRoot("sources")
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Assert(root.Close(), qt.IsNil) })
	files := make(map[string][]byte)
	for _, file := range needed {
		data, err := root.ReadFile(file.Path)
		c.Assert(err, qt.IsNil)
		files[file.Path] = data
	}
	data, err = fixtureRoot.ReadFile("round.json")
	c.Assert(err, qt.IsNil)
	var round map[string]json.RawMessage
	c.Assert(json.Unmarshal(data, &round), qt.IsNil)
	return TrainingInput{manifest, files, round}
}

// Encode marshals a test value and fails the test on error.
func Encode(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	return data
}

// Compile binds simulated judgments to the current source packet.
func (f TrainingInput) Compile(t *testing.T) (corpus.Artifact, *annotation.Round) {
	t.Helper()
	c := qt.New(t)
	plan, err := corpus.MakePlan(t.Context(), f.Manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, f.Files)
	c.Assert(err, qt.IsNil)
	units := make([]annotation.Unit, 0, len(artifact.Units))
	for _, candidate := range artifact.Units {
		units = append(units, candidate.Unit)
	}
	f.Round["units"] = Encode(t, units)
	var judgments []annotation.Judgment
	c.Assert(json.Unmarshal(f.Round["judgments"], &judgments), qt.IsNil)
	f.Round["judgments"] = json.RawMessage(`[]`)
	unloaded, err := annotation.Load(t.Context(), Encode(t, f.Round))
	c.Assert(err, qt.IsNil)
	packet, err := unloaded.Packet(t.Context())
	c.Assert(err, qt.IsNil)
	for i := range judgments {
		judgments[i].PacketSHA256 = packet.SHA256
	}
	f.Round["judgments"] = Encode(t, judgments)
	round, err := annotation.Load(t.Context(), Encode(t, f.Round))
	c.Assert(err, qt.IsNil)
	return artifact, round
}

// ReplaceSource updates one source and its manifest digest.
func (f *TrainingInput) ReplaceSource(index int, text string) {
	source := &f.Manifest.Sources[index]
	f.Files[source.Path] = []byte(text + "\n")
	source.Bytes = len(f.Files[source.Path])
	source.SHA256 = fmt.Sprintf("%x", sha256.Sum256(f.Files[source.Path]))
}
