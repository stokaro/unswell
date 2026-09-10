package figures_test

import (
	"encoding/xml"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/evaluation"
	"github.com/stokaro/unswell/research/annotation/figures"
)

func summarize(c *qt.C, rows []evaluation.Observation) evaluation.Summary {
	c.Helper()
	summary, err := evaluation.Summarize(c.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	return summary
}

func observed(id, group string, label int, response float64) evaluation.Observation {
	positive := response >= 0.5
	return evaluation.Observation{UnitID: id, GroupID: group, Label: label, Response: &response, Positive: &positive}
}

// A figure is evidence only if it comes from the record it claims to describe,
// so each chart is checked against values the summary actually holds.
func TestChartsPlotTheSavedRecord(t *testing.T) {
	c := qt.New(t)
	summary := summarize(c, []evaluation.Observation{
		observed("a", "one", 1, 0.9), observed("b", "one", 0, 0.2), observed("c", "two", 0, 0.8),
	})
	for _, name := range figures.Names() {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			svg, err := figures.Render(name, summary)
			c.Assert(err, qt.IsNil)
			c.Assert(xml.Unmarshal(svg, new(any)), qt.IsNil)
			c.Assert(string(svg), qt.Contains, figures.Version)
			c.Assert(string(svg), qt.Not(qt.Contains), "no covered decisions")
			again, err := figures.Render(name, summary)
			c.Assert(err, qt.IsNil)
			c.Assert(again, qt.DeepEquals, svg)
		})
	}
}

// An empty record must say so. A blank plot area reads as a perfect result.
func TestChartsStateWhenThereIsNothingToPlot(t *testing.T) {
	c := qt.New(t)
	summary := summarize(c, []evaluation.Observation{{UnitID: "a", GroupID: "one", Label: 1}})
	for _, name := range figures.Names() {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			svg, err := figures.Render(name, summary)
			c.Assert(err, qt.IsNil)
			c.Assert(string(svg), qt.Contains, "no covered decisions to plot")
		})
	}
}

// The reliability chart labels each bin with its count, so a point supported by
// one observation cannot be read as a stable rate.
func TestReliabilityShowsBinSupport(t *testing.T) {
	c := qt.New(t)
	summary := summarize(c, []evaluation.Observation{
		observed("a", "one", 1, 0.95), observed("b", "one", 1, 0.96), observed("c", "one", 0, 0.05),
	})
	svg, err := figures.Render("reliability", summary)
	c.Assert(err, qt.IsNil)
	c.Assert(string(svg), qt.Contains, ">2</text>")
	c.Assert(string(svg), qt.Contains, ">1</text>")
}

// The risk-coverage chart names the accepted count behind its final point.
func TestRiskCoverageLabelsItsFinalPoint(t *testing.T) {
	c := qt.New(t)
	summary := summarize(c, []evaluation.Observation{
		observed("a", "one", 1, 0.9), observed("b", "one", 0, 0.9), observed("c", "one", 0, 0.1),
	})
	svg, err := figures.Render("risk-coverage", summary)
	c.Assert(err, qt.IsNil)
	c.Assert(string(svg), qt.Contains, "1 of 3 wrong")
}

func TestUnknownFigureIsRejected(t *testing.T) {
	c := qt.New(t)
	summary := summarize(c, []evaluation.Observation{observed("a", "one", 1, 0.9)})
	_, err := figures.Render("scatter", summary)
	c.Assert(err, qt.ErrorMatches, `unknown figure "scatter"`)
}
