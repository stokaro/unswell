package training

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

type fixture struct {
	manifest corpus.Manifest
	files    map[string][]byte
	round    map[string]json.RawMessage
}

func fixtureInput(t *testing.T) fixture {
	t.Helper()
	c := qt.New(t)
	data, err := os.ReadFile("testdata/manifest.json")
	c.Assert(err, qt.IsNil)
	manifest, err := corpus.LoadManifest(t.Context(), data)
	c.Assert(err, qt.IsNil)
	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil)
	needed, err := corpus.Files(t.Context(), plan)
	c.Assert(err, qt.IsNil)
	root, err := os.OpenRoot("testdata/sources")
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Assert(root.Close(), qt.IsNil) })
	files := make(map[string][]byte)
	for _, file := range needed {
		data, err := root.ReadFile(file.Path)
		c.Assert(err, qt.IsNil)
		files[file.Path] = data
	}
	data, err = os.ReadFile("testdata/round.json")
	c.Assert(err, qt.IsNil)
	var round map[string]json.RawMessage
	c.Assert(json.Unmarshal(data, &round), qt.IsNil)
	return fixture{manifest, files, round}
}

func encode(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	return data
}

func (f fixture) compile(t *testing.T) (corpus.Artifact, *annotation.Round) {
	t.Helper()
	c := qt.New(t)
	plan, err := corpus.MakePlan(t.Context(), f.manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, f.files)
	c.Assert(err, qt.IsNil)
	units := make([]annotation.Unit, 0, len(artifact.Units))
	for _, candidate := range artifact.Units {
		units = append(units, candidate.Unit)
	}
	f.round["units"] = encode(t, units)
	var judgments []annotation.Judgment
	c.Assert(json.Unmarshal(f.round["judgments"], &judgments), qt.IsNil)
	f.round["judgments"] = json.RawMessage(`[]`)
	unloaded, err := annotation.Load(t.Context(), encode(t, f.round))
	c.Assert(err, qt.IsNil)
	packet, err := unloaded.Packet(t.Context())
	c.Assert(err, qt.IsNil)
	for i := range judgments {
		judgments[i].PacketSHA256 = packet.SHA256
	}
	f.round["judgments"] = encode(t, judgments)
	round, err := annotation.Load(t.Context(), encode(t, f.round))
	c.Assert(err, qt.IsNil)
	return artifact, round
}

func (f *fixture) replaceSource(index int, text string) {
	source := &f.manifest.Sources[index]
	f.files[source.Path] = []byte(text + "\n")
	source.Bytes = len(f.files[source.Path])
	source.SHA256 = fmt.Sprintf("%x", sha256.Sum256(f.files[source.Path]))
}

func fittingOptions() Options {
	return Options{Kind: "paragraph", Features: []string{"prose-words"}, MissingFeatures: "reject", Calibration: "isotonic",
		AllowSimulation: true, Fit: Fit{L2: 1, Tolerance: 1e-8, MaxIterations: 100, MaxOperations: 10_000_000}}
}
