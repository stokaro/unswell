package corpus_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func sample() (corpus.Manifest, map[string][]byte) {
	files := map[string][]byte{
		"LICENSE": []byte("Test-owned source and notice fixture.\n"),
		"readme.md": []byte("\xef\xbb\xbf# Cache\r\n\r\nThe **cache** may retry &amp; wait." +
			"\r\n\r\nKeep `protected_identifier` unchanged.\r\n\r\n```go\n// Hidden code text.\n```\n"),
		"sample.go": []byte("// Package sample defines the cache.\npackage sample\n\nconst Message = \"Cache \\u0065ntry cannot be decoded.\"\n"),
	}
	notice := corpus.Notice{Path: "LICENSE", SHA256: hash(files["LICENSE"]), Bytes: len(files["LICENSE"])}
	manifest := corpus.Manifest{Version: corpus.Version, ID: "test-round", Seed: "frozen-seed",
		Weights: corpus.Weights{Training: 6000, Development: 1500, Calibration: 1500, FinalTest: 1000},
		Policy:  extract.Policy{}, UnitKinds: []string{"paragraph", "sentence", "fragment"}}
	for i, name := range []string{"readme.md", "sample.go"} {
		format := document.Markdown
		if i == 1 {
			format = document.Go
		}
		manifest.Sources = append(manifest.Sources, corpus.Source{ID: fmt.Sprintf("d%d", i), Path: name, SHA256: hash(files[name]),
			Bytes: len(files[name]), Format: format, ProseLanguage: "en", Repository: "test-project", Document: "test-project/" + name,
			Reference: "fixture:" + name, Topic: "cache", Purpose: "Explain cache behavior", Role: "documentation",
			Origin:  annotation.Origin{Label: "unknown", Scope: "repository", Evidence: "Teaching fixture; origin is not a quality label."},
			Rights:  annotation.Rights{License: "test-fixture", Evidence: "Test-owned bytes", AllowedUses: []string{"annotation"}},
			Notices: []corpus.Notice{notice}})
	}
	return manifest, files
}

func hash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func encoded(c *qt.C, value any) []byte {
	c.Helper()
	data, err := json.Marshal(value)
	c.Assert(err, qt.IsNil)
	return data
}

func TestFrozenGroupsAndOrder(t *testing.T) {
	c := qt.New(t)
	m, _ := sample()
	m.Sources[0].Partition = "development"
	plan, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	c.Assert(plan.Groups, qt.HasLen, 1)
	c.Assert(plan.Groups[0].Partition, qt.Equals, "development")
	c.Assert(plan.Groups[0].Sources, qt.DeepEquals, []string{"d0", "d1"})
	slices.Reverse(m.Sources)
	slices.Reverse(m.UnitKinds)
	again, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, plan)
	m.Sources[0].Partition = "final_test"
	_, err = corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.ErrorMatches, ".*conflicting partition pins.*")
}

func TestRelationshipFamilies(t *testing.T) {
	for _, family := range []string{"document", "author", "template", "related", "generation", "hash"} {
		t.Run(family, func(t *testing.T) {
			c := qt.New(t)
			m, _ := sample()
			m.Sources[1].Repository = "second-project"
			connect(&m, family)
			m.Sources[0].Partition, m.Sources[1].Partition = "training", "final_test"
			_, err := corpus.MakePlan(t.Context(), m)
			c.Assert(err, qt.ErrorMatches, ".*conflicting partition pins.*")
		})
	}
}

func connect(m *corpus.Manifest, family string) {
	switch family {
	case "document":
		m.Sources[1].Document = m.Sources[0].Document
	case "author":
		m.Sources[0].Authors, m.Sources[1].Authors = []string{"author:a"}, []string{"author:a"}
	case "template":
		m.Sources[0].Templates, m.Sources[1].Templates = []string{"template:a"}, []string{"template:a"}
	case "related":
		m.Sources[0].Related, m.Sources[1].Related = []string{"revision:a"}, []string{"revision:a"}
	case "generation":
		m.Sources[0].GenerationTasks, m.Sources[1].GenerationTasks = []string{"prompt:a"}, []string{"prompt:a"}
	case "hash":
		m.Sources[1].SHA256 = m.Sources[0].SHA256
	}
}

func TestTransitiveGroups(t *testing.T) {
	c := qt.New(t)
	m, _ := sample()
	third := m.Sources[1]
	third.ID, third.Path, third.SHA256, third.Document, third.Repository = "d2", "third.go", hash([]byte("third")), "third-doc", "third-repo"
	third.Bytes = 5
	m.Sources[1].Related, third.Related = []string{"bridge"}, []string{"bridge"}
	m.Sources[0].Partition, third.Partition = "training", "final_test"
	m.Sources = append(m.Sources, third)
	_, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.ErrorMatches, ".*conflicting partition pins.*")
}

func TestPlanCannotBeReassignedOrReused(t *testing.T) {
	c := qt.New(t)
	m, _ := sample()
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	p.Groups[0].Partition = "invented"
	c.Assert(corpus.ValidatePlan(t.Context(), p), qt.ErrorMatches, ".*does not match.*")
	p, err = corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	p.Manifest.Sources[0].Purpose = "Changed context"
	c.Assert(corpus.ValidatePlan(t.Context(), p), qt.ErrorMatches, ".*does not match.*")
}

func TestExtractionAndVerification(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	start := strings.Index(string(files["sample.go"]), "Cache ")
	m.Sources[1].Roles = []corpus.RoleRegion{{Span: document.Span{Start: start, End: len(files["sample.go"])}, Role: "error_message"}}
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	a, err := corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	c.Assert(a.Status, qt.Equals, "unlabeled_candidates")
	found := make(map[string]corpus.Candidate)
	for _, unit := range a.Units {
		c.Assert(strings.ContainsRune(unit.Unit.Text, 0), qt.IsFalse)
		c.Assert(unit.Unit.Text, qt.Not(qt.Contains), "protected_identifier")
		c.Assert(unit.Unit.Text, qt.Not(qt.Contains), "Hidden code")
		c.Assert(unit.Unit.Origin.Label, qt.Equals, "unknown")
		for _, span := range unit.Unit.Source.Segments {
			c.Assert(span.Valid(unit.Unit.Source.Bytes), qt.IsTrue)
		}
		found[unit.Unit.Kind+":"+unit.Unit.Text] = unit
	}
	c.Assert(found["paragraph:The cache may retry & wait."].Unit.Text, qt.Equals, "The cache may retry & wait.")
	message := found["fragment:Cache entry cannot be decoded."]
	c.Assert(message.Unit.Role, qt.Equals, "error_message")
	expectedSpan := document.Span{Start: start, End: start + len("Cache \\u0065ntry cannot be decoded.")}
	c.Assert(message.Unit.Source.Segments, qt.DeepEquals, []document.Span{expectedSpan})
	c.Assert(found["fragment:Keep"].Unit.Kind, qt.Equals, "fragment")
	loaded, err := corpus.LoadArtifact(t.Context(), encoded(c, a))
	c.Assert(err, qt.IsNil)
	verified, err := corpus.Verify(t.Context(), loaded, files)
	c.Assert(err, qt.IsNil)
	c.Assert(verified.Status, qt.Equals, "source_and_candidates_reproduced")
	c.Assert(verified.HumanCorpus, qt.Equals, "not_qualified")
}

func TestSourceAndNoticeTampering(t *testing.T) {
	for _, name := range []string{"sample.go", "LICENSE"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			m, files := sample()
			p, err := corpus.MakePlan(t.Context(), m)
			c.Assert(err, qt.IsNil)
			files[name][0] = '!'
			_, err = corpus.Build(t.Context(), p, files)
			c.Assert(err, qt.ErrorMatches, ".*does not match its recorded bytes and hash.*")
		})
	}
}

func TestResealedCandidateIsNotSourceEvidence(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	a, err := corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	a.Units[0].Unit.Text = "Invented replacement."
	a.SHA256 = ""
	a.SHA256 = hash(encoded(c, a))
	loaded, err := corpus.LoadArtifact(t.Context(), encoded(c, a))
	c.Assert(err, qt.IsNil)
	_, err = corpus.Verify(t.Context(), loaded, files)
	c.Assert(err, qt.ErrorMatches, ".*not reproduced.*")
}

func TestCancellationAndStrictJSON(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = corpus.Build(ctx, p, files)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	for _, input := range []string{`{"version":"a","version":"b"}`, `{"unknown":true}`, `{"id":"\ud800"}`} {
		_, err := corpus.LoadManifest(t.Context(), []byte(input))
		c.Assert(err, qt.IsNotNil)
	}
}

func TestPortablePaths(t *testing.T) {
	c := qt.New(t)
	for _, name := range []string{"../outside.go", "/absolute.go", "C:/outside.go", "a\\b.go", "a/../b.go", ".", "a\x00b"} {
		c.Assert(corpus.ValidPath(name), qt.IsFalse, qt.Commentf("path: %q", name))
	}
	c.Assert(corpus.ValidPath("docs/cache.md"), qt.IsTrue)
}

func FuzzLoadManifest(f *testing.F) {
	m, _ := sample()
	data, err := json.Marshal(m)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(data)
	f.Add([]byte(`{"version":"unswell-corpus-v1"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = corpus.LoadManifest(t.Context(), data)
	})
}
