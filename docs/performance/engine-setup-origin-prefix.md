# Engine setup allocation cost

[#338](https://github.com/stokaro/unswell/issues/338) records ordinary root tests
exceeding the ten-minute package deadline on Windows. A local unchanged root
suite also exceeded that deadline; its bounded rerun passed in 1,120.163 seconds.
The sampled stack was constructing configuration origins, rather than waiting
on a lock or testing the new MCP failure-report code.

`recordNodeOrigins` checked each existing key against `prefix + "/"` inside a
map loop. Long prefixes caused a new string allocation on each iteration. The
fix computes the same prefix once before that loop. It preserves the traversal,
descendant deletion, origin values, and policy identity. No test case is removed
or shared across previously independent engines.

The public `BenchmarkEngineConstruction` benchmark constructs fresh default and
strict engines without analyzing text. The [raw samples](engine-setup-origin-prefix.json)
record the baseline tree, compiler, host, command, and all three five-iteration
runs for each profile. The host was shared and loaded; timing samples are not
an isolated throughput claim.

| Median per construction | Default before | Default after | Strict before | Strict after |
| --- | ---: | ---: | ---: | ---: |
| Allocated bytes | 163,481,302 | 10,060,604 | 455,172,132 | 16,330,606 |
| Allocations | 2,494,904 | 45,488 | 7,091,281 | 76,339 |

The benchmark measures actual allocations, not peak resident memory. Existing
configuration and engine tests verify policy behavior. Final acceptance for
#338 also requires the complete Windows suite on the minimum compiler with
measured margin; these macOS samples alone do not establish that result.
Race, active fuzzing, and coverage remain deferred under #123.
