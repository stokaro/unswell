package extract

// White-box tests: the parse budget is an internal bound of the grammar parser; the package
// exposes no API that reports it.

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestParseBudgetGrowsWithTheInput(t *testing.T) {
	c := qt.New(t)
	c.Assert(parseBudget(0), qt.Equals, 5*time.Second)
	c.Assert(parseBudget(1023), qt.Equals, 5*time.Second)
	c.Assert(parseBudget(1024), qt.Equals, 5*time.Second+500*time.Millisecond)
	c.Assert(parseBudget(19576), qt.Equals, 5*time.Second+19*500*time.Millisecond)
	c.Assert(parseBudget(1<<20), qt.Equals, 5*time.Second+1024*500*time.Millisecond)
}
