package corpus_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func duplicateFixture(c *qt.C) (corpus.Manifest, map[string][]byte) {
	c.Helper()
	m, files := sample()
	base := "# Guide\n\n" + strings.Repeat("The scheduler retries a failed job after the configured delay expires. ", 12)
	files["guide-a.md"] = []byte(base + "\nSee the retry section.\n")
	files["guide-b.md"] = []byte(base + "\nSee the backoff section instead.\n")
	files["short.md"] = []byte("# Short\n\nTwo words.\n")
	for i, name := range []string{"guide-a.md", "guide-b.md", "short.md"} {
		m.Sources = append(m.Sources, corpus.Source{ID: "n" + string(rune('0'+i)), Path: name, SHA256: hash(files[name]),
			Bytes: len(files[name]), Format: document.Markdown, ProseLanguage: "en", Repository: "repo-" + name,
			Document: "repo-" + name + "/" + name, Reference: "fixture:" + name, Topic: "scheduler", Purpose: "Explain retries",
			Role: "documentation", Origin: m.Sources[0].Origin, Rights: m.Sources[0].Rights, Notices: m.Sources[0].Notices})
	}
	return m, files
}

func TestNearDuplicatesAreReportedAndApplied(t *testing.T) {
	c := qt.New(t)
	m, files := duplicateFixture(c)
	report, err := corpus.DetectDuplicates(t.Context(), m, files, corpus.DefaultDuplicateThreshold)
	c.Assert(err, qt.IsNil)
	c.Assert(report.Version, qt.Equals, corpus.DuplicatesVersion)
	c.Assert(report.Sources, qt.Equals, 5)
	c.Assert(report.TooShort, qt.DeepEquals, []string{"n2"})
	c.Assert(report.Clusters, qt.HasLen, 1)
	c.Assert(report.Clusters[0].Sources, qt.DeepEquals, []string{"n0", "n1"})
	c.Assert(report.Clusters[0].Key, qt.Matches, `near-duplicate-v1:[0-9a-f]{16}`)
	var pair corpus.DuplicatePair
	for _, item := range report.Pairs {
		if item.A == "n0" && item.B == "n1" {
			pair = item
		}
	}
	c.Assert(pair.ExactJaccard >= corpus.DefaultDuplicateThreshold, qt.IsTrue, qt.Commentf("%+v", pair))
	c.Assert(pair.ExactJaccard < 1, qt.IsTrue)
	c.Assert(pair.EstimatedJaccard > 0, qt.IsTrue)
	for _, item := range report.Pairs {
		c.Assert(item.A < item.B, qt.IsTrue)
		if item.A == "d0" || item.B == "d0" {
			c.Assert(item.ExactJaccard < corpus.DefaultDuplicateThreshold, qt.IsTrue)
		}
	}
	// The same inputs produce the same report; the planner ignores the report
	// until the curator applies its keys, after which the pair shares a group.
	again, err := corpus.DetectDuplicates(t.Context(), m, files, corpus.DefaultDuplicateThreshold)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, report)
	before, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	c.Assert(groupOf(before, "n0"), qt.Not(qt.Equals), groupOf(before, "n1"))
	applied, err := corpus.ApplyDuplicates(t.Context(), m, report)
	c.Assert(err, qt.IsNil)
	after, err := corpus.MakePlan(t.Context(), applied)
	c.Assert(err, qt.IsNil)
	c.Assert(groupOf(after, "n0"), qt.Equals, groupOf(after, "n1"))
	c.Assert(groupOf(after, "n2"), qt.Not(qt.Equals), groupOf(after, "n0"))
	for _, source := range applied.Sources {
		if source.ID == "n0" || source.ID == "n1" {
			c.Assert(source.Related, qt.DeepEquals, []string{report.Clusters[0].Key})
		} else {
			c.Assert(source.Related, qt.HasLen, len(sourceByID(m, source.ID).Related))
		}
	}
}

func groupOf(plan corpus.Plan, id string) string {
	for _, group := range plan.Groups {
		if slices.Contains(group.Sources, id) {
			return group.ID
		}
	}
	return ""
}

func sourceByID(m corpus.Manifest, id string) corpus.Source {
	for _, source := range m.Sources {
		if source.ID == id {
			return source
		}
	}
	return corpus.Source{}
}

func TestNearDuplicatesRejectBadInputs(t *testing.T) {
	c := qt.New(t)
	m, files := duplicateFixture(c)
	for _, threshold := range []float64{0, -0.1, 1.5} {
		_, err := corpus.DetectDuplicates(t.Context(), m, files, threshold)
		c.Assert(err, qt.IsNotNil)
	}
	report, err := corpus.DetectDuplicates(t.Context(), m, files, 1)
	c.Assert(err, qt.IsNil)
	c.Assert(report.Clusters, qt.HasLen, 0)
	// Tampered bytes and a report for another manifest are rejected.
	files["guide-a.md"] = append(files["guide-a.md"], '!')
	_, err = corpus.DetectDuplicates(t.Context(), m, files, corpus.DefaultDuplicateThreshold)
	c.Assert(err, qt.IsNotNil)
	other := m
	other.ID = "other"
	_, err = corpus.ApplyDuplicates(t.Context(), other, report)
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = corpus.DetectDuplicates(ctx, m, files, corpus.DefaultDuplicateThreshold)
	c.Assert(err, qt.IsNotNil)
}
