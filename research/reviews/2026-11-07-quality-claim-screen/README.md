# Exposed source-local quality-claim screen for #349

The bounded source-local construction in [PROTOCOL.md](PROTOCOL.md) was frozen
before the full CLI replay. This is development evidence on pages already read
by the assistant, not independent confirmation. No model calls or new labels
were used. The source archives, 123 reference events, and technical/strict
profile policies stayed unchanged. `runs.json` pins the binary, policy, and
complete report hashes. The compressed JSON reports contain all 36 full source
pages and every finding; `analyze.py` checks them against the previous
version-11 replay and the frozen target spans.

| Cohort | Events now FULL | Technical accepted/delivered | Strict accepted/delivered |
| --- | ---: | ---: | ---: |
| Whole page | 28/57 | 82/126 | 82/145 |
| Context | 16/46 | 21/27 | 21/28 |
| Historical confirmation | 4/20 | 6/9 | 6/10 |

Five new findings, all accepted on assistant review, are unchanged between
profiles. No prior finding was removed or changed in substance. Prior
`filler.unscoped-assurance` identities changed because its version rose from
11 to 12; other finding identities and content stayed the same. The 18 frozen
external constructed controls yielded no changed findings. The five new
findings are:

| Frozen event | Finding | Editorial review |
| --- | --- | --- |
| `whole/p013-d01` | `APIs are polished and future-proofed` | Accepted. The release note promises unspecified future compatibility alongside concrete 1.0 scope. Keep the concrete stabilization statement; verify the broader promise. |
| `whole/p018-d02` | `important effect` | Accepted. The sRGB advice names the setting but not the visible rendering difference or affected behavior. Preserve the sRGB condition and request the actual effect. |
| `whole/p019-d02` | `most comprehensive feature set` | Accepted. The list supplies neither a comparison set nor a coverage criterion. Keep the listed features; name the criterion or use a scoped comparison. |
| `whole/p023-d01` | `powerful and flexible query language` | Accepted. The adjectives do not describe supported PromQL operations. Retain PromQL and its dimensional model; name the operations. |
| `context/c05-d05` | `reasonable defaults` | Accepted. The sentence does not identify the default OCI settings or why they fit. Retain the `Opt` configuration detail and request the defaults or a reference. |

Each candidate's diagnostic points to the subjective span and asks for a
criterion or concrete capability. Four construction-specific suggestions
explicitly require verification before weakening or removing a claim; the
copular-claim suggestion asks to keep technical conditions and replace only
the bare quality announcement. None asks to delete a technical condition,
identifier, or numeric value.
The new-finding acceptance screen is 5/5, above the predeclared four-event
and 85% thresholds. Because these are exposed pages, this only supports a
prospective untouched confirmation; it does not qualify #349. All three
cohorts remain below both 80% FULL recall and 85% accepted delivery in both
profiles. In particular, whole-page technical recall is 28/57 and accepted
delivery is 82/126.

Reproduce the checks from the repository root after building the CLI:

```sh
python3 -B research/reviews/2026-11-07-quality-claim-screen/analyze.py --check
python3 -B -m unittest discover -s research/reviews/2026-11-07-quality-claim-screen -p 'test_*.py'
python3 -B research/reviews/2026-11-07-scoped-degree/check_controls.py --binary /path/to/unswell
```

The reports can be replayed with
`research/reviews/2026-09-26-outcome-announcements/replay.py` using the frozen
36-page archives. Avoid interpreting the 5/5 exposed result as independent
precision or as detection of AI authorship.
