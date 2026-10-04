# Contrast frequency and default review load

The default contrast note usually points at necessary technical alternatives.
On the 36 complete exposed development pages, unchanged frozen assistant
judgments accepted one of its 30 findings, rejected 27, and left two uncertain.
Counting contrast markers does not identify a repeated proposition or the
wording that needs revision.

`syntax.paired-contrast-density` version 3 therefore stays opt-in in every
builtin profile. The matcher and all parameters are unchanged from version 2.
Explicit configuration preserves its results, measurements, zero score, and
nonblocking gate. The existing construction rules retain their diagnostics.
[ADR 0043](../../../docs/adr/0043-contrast-frequency-opt-in.md) records the policy.

| Complete original stream | Before | Default after | Explicit opt-in after |
| --- | ---: | ---: | ---: |
| technical | 173 | 143 | 173 |
| strict | 194 | 164 | 194 |

All other diagnostics, exact source bytes, protected regions, source locations,
and unit scores are unchanged. The default stream removes 30 zero-score notes
in each profile, including the one accepted note. The [accepted loss](accepted-loss.json)
concerns repeated reporting-versus-failure framing in `whole/p004.md`; it remains
part of the research denominator. No new event credit or improved recall is
claimed. Shared unit assessments remain available at all 10,553 locations.

The comparison retains all 36 sources, 4,549 original units, 804 original
criticism judgments, and 123 reference events. Frozen labels are neither
changed nor promoted. Review is by the maintainer-accepted root assistant under
ADR 0041, on exposed development sources. This is not independent human review,
separate confirmation, an authorship result, or qualification of the broader
80% recall / 85% accepted-finding objective. Related: [#349](https://github.com/stokaro/unswell/issues/349).

The [protocol](protocol.json) predates implementation. [Comparison](comparison.json)
records aggregate losses, complete source scope, explicit matcher preservation,
and ruleset identity. [Measurement](measurement.json) binds six complete compressed
reports. The [source manifest](source-manifest.json) preserves the original
identities. The private original judgment index is not republished.

An initial local comparison failed because a rule-version change renames finding
IDs in assessment contributions. The corrected comparison binds each new ID to
its exact unchanged rule and source locations before comparing every contribution.
Default assessments omit only the removed zero-score contributions. The failed
attempt remains separate locally; it produced no successful aggregate.

Build the preceding engine from the recorded base commit and the candidate
engine from this change, then reproduce in a new output directory:

```sh
python3 -B reproduce.py --before /path/to/base-unswell \
  --after /path/to/current-unswell --output /path/to/new-comparison
```

The replay runs only the offline CLI. No model call, prospective-source access,
threshold change, calibration, or automatic edit is part of this comparison.
