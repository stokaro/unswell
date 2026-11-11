package extract

// White-box tests: Check which run of lines recovery blanks and that a run
// holding half of a fenced block extends to the closing fence. The public
// parser only shows the final document, and no input makes the pinned grammar
// fail inside a fence reliably enough to pin this choice from outside.

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestRecoveryRegionIsTheRunOfLinesAtTheDefect(t *testing.T) {
	c := qt.New(t)
	source := []byte("x\n\nFirst paragraph\nsecond line.\n\nThird paragraph.\n")
	region, ok := nextRegion(source, 20, nil)
	c.Assert(ok, qt.IsTrue)
	c.Assert(string(source[region.Start:region.End]), qt.Equals, "First paragraph\nsecond line.\n")
	// A run that is already blanked gives way to the next one.
	next, ok := nextRegion(source, 20, []document.Span{region})
	c.Assert(ok, qt.IsTrue)
	c.Assert(string(source[next.Start:next.End]), qt.Equals, "Third paragraph.\n")
}

func TestRecoveryRegionClosesAnOpenFence(t *testing.T) {
	c := qt.New(t)
	source := []byte("x\n\nText before.\n```sh\nmake build\n\nmake test\n```\n\nText after.\n")
	region, ok := nextRegion(source, 5, nil)
	c.Assert(ok, qt.IsTrue)
	c.Assert(string(source[region.Start:region.End]), qt.Equals, "Text before.\n```sh\nmake build\n\nmake test\n```\n")
}

func TestRecoveryNeverBlanksTheFrame(t *testing.T) {
	c := qt.New(t)
	source := []byte("x\n\nOnly paragraph.\n")
	region, ok := nextRegion(source, 0, nil)
	c.Assert(ok, qt.IsTrue)
	c.Assert(region.Start >= markdownFramePrefix, qt.IsTrue)
	blankRegion(source, region)
	c.Assert(string(source[:markdownFramePrefix]), qt.Equals, "x\n\n")
}
