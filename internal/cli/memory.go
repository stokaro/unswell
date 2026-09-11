package cli

import (
	"os"
	"runtime/debug"
)

// memoryEnvelope is the soft limit for the runtime's total memory. The
// documented envelope of one scan is 512 MiB on a 2-vCPU host. Three quarters
// of it leaves room for the mapped binary and the runtime's own bookkeeping:
// a host limit counts them, the runtime does not. Near the limit the collector
// runs more often. Without the limit the heap grew to twice its live size, and
// a tree whose live data fit still died on a host that enforces the envelope.
// Live data above the limit still runs: the collector works harder and the
// scan goes on.
const memoryEnvelope = 384 << 20

// applyMemoryEnvelope installs the soft limit unless the environment set one.
func applyMemoryEnvelope() {
	if os.Getenv("GOMEMLIMIT") != "" {
		return
	}
	debug.SetMemoryLimit(memoryEnvelope)
}
