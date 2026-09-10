// Package figures renders evaluation records as SVG charts. Every value comes
// from a saved record, so a published figure cannot disagree with the numbers
// beside it.
package figures

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

// Version identifies the chart layout, so a regenerated figure that looks
// different is a deliberate change rather than a silent one.
const Version = "unswell-research-figure-v1"

// Names returns the supported chart names in a stable order.
func Names() []string { return []string{"reliability", "risk-coverage"} }

const (
	width, height   = 480, 360
	left, right     = 60, 20
	top, bottom     = 30, 45
	plotW, plotH    = width - left - right, height - top - bottom
	tickCount       = 5
	pointRadius     = 4
	emptyExplanator = "no covered decisions to plot"
)

// Render draws one named chart for a summary. An unknown name is an error, not
// an empty image.
func Render(name string, summary evaluation.Summary) ([]byte, error) {
	switch name {
	case "reliability":
		return reliability(summary), nil
	case "risk-coverage":
		return riskCoverage(summary), nil
	default:
		return nil, fmt.Errorf("unknown figure %q", name)
	}
}

type canvas struct {
	buffer bytes.Buffer
}

func newCanvas(title, xLabel, yLabel string) *canvas {
	c := &canvas{}
	fmt.Fprintf(&c.buffer, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" `+
		`font-family="sans-serif" font-size="11" data-figure="%s">`, width, height, width, height, Version)
	c.buffer.WriteString(`<rect width="100%" height="100%" fill="#ffffff"/>`)
	c.text(width/2, 18, "middle", title)
	c.text(left+plotW/2, height-10, "middle", xLabel)
	fmt.Fprintf(&c.buffer, `<text x="14" y="%d" text-anchor="middle" transform="rotate(-90 14 %d)">%s</text>`,
		top+plotH/2, top+plotH/2, escape(yLabel))
	fmt.Fprintf(&c.buffer, `<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="#333333"/>`,
		left, top, plotW, plotH)
	for i := 0; i <= tickCount; i++ {
		value := float64(i) / tickCount
		x, y := c.position(value, 0), c.position(0, value)
		fmt.Fprintf(&c.buffer, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#cccccc"/>`, x.x, top, x.x, top+plotH)
		fmt.Fprintf(&c.buffer, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#cccccc"/>`, left, y.y, left+plotW, y.y)
		c.text(x.x, top+plotH+16, "middle", fmt.Sprintf("%.1f", value))
		c.text(left-8, y.y+4, "end", fmt.Sprintf("%.1f", value))
	}
	return c
}

type point struct{ x, y int }

// position maps the unit square onto the plot area with y growing upward.
func (c *canvas) position(x, y float64) point {
	return point{x: left + int(x*float64(plotW)+0.5), y: top + plotH - int(y*float64(plotH)+0.5)}
}

func (c *canvas) text(x, y int, anchor, value string) {
	fmt.Fprintf(&c.buffer, `<text x="%d" y="%d" text-anchor="%s">%s</text>`, x, y, anchor, escape(value))
}

func (c *canvas) note(value string) {
	c.text(left+plotW/2, top+plotH/2, "middle", value)
}

func (c *canvas) finish() []byte {
	c.buffer.WriteString("</svg>\n")
	return c.buffer.Bytes()
}

func escape(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return replacer.Replace(value)
}
