package figures

import (
	"fmt"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

// reliability plots observed positive rate against mean response for each
// nonempty bin. The diagonal is perfect calibration; distance from it is the
// term ECE sums. An empty bin has no rate and is left out rather than drawn at
// zero, which would read as a confident wrong prediction.
func reliability(summary evaluation.Summary) []byte {
	c := newCanvas("Reliability", "mean response", "observed positive rate")
	first, last := c.position(0, 0), c.position(1, 1)
	fmt.Fprintf(&c.buffer, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#999999" stroke-dasharray="4 3"/>`,
		first.x, first.y, last.x, last.y)
	drawn := 0
	for _, bin := range summary.Micro.Bins {
		if bin.Count == 0 || bin.MeanResponse == nil || bin.PositiveRate == nil {
			continue
		}
		at := c.position(*bin.MeanResponse, *bin.PositiveRate)
		fmt.Fprintf(&c.buffer, `<circle cx="%d" cy="%d" r="%d" fill="#1f4e79"/>`, at.x, at.y, pointRadius)
		c.text(at.x, at.y-8, "middle", fmt.Sprintf("%d", bin.Count))
		drawn++
	}
	if drawn == 0 {
		c.note(emptyExplanator)
	}
	return c.finish()
}

// riskCoverage plots the error rate among accepted decisions against the share
// of eligible targets they cover. A curve that rises to the right means the
// model's confident answers are its better ones.
func riskCoverage(summary evaluation.Summary) []byte {
	c := newCanvas("Risk against coverage", "coverage of eligible targets", "error rate among accepted")
	points := summary.Micro.RiskCoverage
	if len(points) == 0 {
		c.note(emptyExplanator)
		return c.finish()
	}
	c.buffer.WriteString(`<polyline fill="none" stroke="#1f4e79" stroke-width="2" points="`)
	for index, entry := range points {
		at := c.position(entry.Coverage, entry.Risk)
		if index > 0 {
			c.buffer.WriteString(" ")
		}
		fmt.Fprintf(&c.buffer, "%d,%d", at.x, at.y)
	}
	c.buffer.WriteString(`"/>`)
	for _, entry := range points {
		at := c.position(entry.Coverage, entry.Risk)
		fmt.Fprintf(&c.buffer, `<circle cx="%d" cy="%d" r="%d" fill="#1f4e79"/>`, at.x, at.y, pointRadius-1)
	}
	final := points[len(points)-1]
	at := c.position(final.Coverage, final.Risk)
	c.text(at.x, at.y-8, "end", fmt.Sprintf("%d of %d wrong", final.Errors, final.Accepted))
	return c.finish()
}
