package e2e_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

type preparedCorpusTarget struct {
	Kind, TextHash, ContextHash string
	Segments                    []document.Span
}

func TestPreparedCollectionMatchesFrozenCorpusTargets(t *testing.T) {
	c := qt.New(t)
	sources, paths, config := preparedCorpusInputs(t)
	engine, err := unswell.New(unswell.Options{Config: config, PreparedFeatures: []string{"prose-words"},
		PreparedKinds: []string{"sentence", "paragraph", "fragment"}})
	c.Assert(err, qt.IsNil)
	result, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	var frozen []corpusExpectedUnit
	decodeFile(t, "corpusdata/ptah-units.golden.json", &frozen)
	c.Assert(frozen, qt.HasLen, 378)
	want := make(map[string][]preparedCorpusTarget)
	for _, unit := range frozen {
		path, exists := paths[unit.Source]
		c.Assert(exists, qt.IsTrue)
		want[path] = append(want[path], preparedCorpusTarget{unit.Kind,
			fmt.Sprintf("%x", sha256.Sum256([]byte(unit.Text))), fmt.Sprintf("%x", sha256.Sum256([]byte(unit.Context))), unit.Segments})
	}
	got := make(map[string][]preparedCorpusTarget)
	for _, source := range result.PreparedFeatures.Sources {
		for _, unit := range source.Units {
			b := unit.Binding
			got[source.Path] = append(got[source.Path], preparedCorpusTarget{b.Kind, b.TextSHA256, b.ContextSHA256, b.Segments})
		}
	}
	c.Assert(got, qt.DeepEquals, want)
}

func preparedCorpusInputs(t *testing.T) ([]document.Source, map[string]string, []byte) {
	t.Helper()
	c := qt.New(t)
	var manifest struct {
		Policy  extract.Policy `json:"extraction_policy"`
		Sources []struct {
			ID, Path, SHA256 string
			Format           document.Format
		}
	}
	data, err := os.ReadFile("../research/annotation/corpus/testdata/ptah-manifest.json")
	c.Assert(err, qt.IsNil)
	c.Assert(json.Unmarshal(data, &manifest), qt.IsNil)
	c.Assert(manifest.Sources, qt.HasLen, 8)
	var sources []document.Source
	paths := make(map[string]string)
	for _, source := range manifest.Sources {
		data, err := os.ReadFile(filepath.Join("..", "research", "annotation", "corpus", "testdata", "ptah", source.Path))
		c.Assert(err, qt.IsNil)
		c.Assert(fmt.Sprintf("%x", sha256.Sum256(data)), qt.Equals, source.SHA256)
		sources = append(sources, document.Source{Name: source.Path, Format: source.Format, Bytes: data})
		paths[source.ID] = source.Path
	}
	config, err := json.Marshal(map[string]any{"version": 1, "extends": []string{"builtin:custom"}, "extraction": manifest.Policy})
	c.Assert(err, qt.IsNil)
	return sources, paths, config
}
