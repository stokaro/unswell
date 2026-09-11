package cli

// White-box tests: the memory envelope is installed by an unexported helper at
// the start of Run, and the property under test is that an explicit
// GOMEMLIMIT in the environment wins over the built-in envelope.

import (
	"runtime/debug"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestMemoryEnvelopeDefersToEnvironment(t *testing.T) {
	c := qt.New(t)
	previous := debug.SetMemoryLimit(-1)
	t.Cleanup(func() { debug.SetMemoryLimit(previous) })

	t.Setenv("GOMEMLIMIT", "")
	debug.SetMemoryLimit(1 << 40)
	applyMemoryEnvelope()
	c.Assert(debug.SetMemoryLimit(-1), qt.Equals, int64(memoryEnvelope))

	t.Setenv("GOMEMLIMIT", "1GiB")
	debug.SetMemoryLimit(1 << 40)
	applyMemoryEnvelope()
	c.Assert(debug.SetMemoryLimit(-1), qt.Equals, int64(1<<40))
}
