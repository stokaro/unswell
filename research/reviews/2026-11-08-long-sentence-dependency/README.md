# Long-sentence dependency feasibility audit

November 8, 2026. This is the bounded source-only follow-up selected in the
[#421 decision](../2026-11-03-issue421/README.md), using the current
November 7 assistant review. The [protocol](PROTOCOL.md), input hashes, and
pair-selection script were frozen in commit `c7818cf` before the packet was
generated. The six original pages and A/B packet were generated from the
tracked source archives and remain in the local ignored artifact directory;
the packet SHA-256 is
`49575965ce44b0254eeac56fc2aaa04e2169a34f7895a7ef45951432ed7138a4`.

The assistant read each complete original page and recorded twelve
source-grounded decisions in [review.json](review.json), then wrote
[review-seal.json](review-seal.json) before opening the separate
[key.json](key.json). The sealed review SHA-256 is
`aada766c3dfc8dd4ab8a011914f3f39ac9c342b4c2efa826c64877b1da3d1d80`.
This is a label-hidden review within an exposed development attempt. The same
assistant had seen other review grades earlier, and the seal is a recorded
sequence, not independent proof of blindness or human annotation.

## Result

| Pair | Page/topic | Actionable member | Uncertain member | Separates? |
| --- | --- | --- | --- | --- |
| 01 | Ptah database roles | split needed | retain | yes |
| 02 | ripgrep installation/build | retain | retain | no |
| 03 | MySQL/MariaDB name and key rules | split needed | split needed | no |
| 04 | Atlas project config | retain | retain | no |
| 05 | BuildKit solver | retain | retain | no |
| 06 | Liquibase migration | split needed | retain | yes |

Only **2/6** pairs separated. The sole FULL-event witness,
`whole/p026-d02`, was judged `split_needed`, but the preregistered screen
required at least **5/6** separated pairs as well as that witness. The screen
failed. The source-local distinction between independent parts and a coherent
dependency does not justify a new long-sentence restriction on this evidence.
No threshold was tuned, no rule was changed, and no full-stream replay was
started. [result.json](result.json) is the machine-readable outcome.

The failure is informative. Pair 02's previously actionable MUSL/PCRE2
sentence remains one prerequisite with a distribution-specific qualification.
Pair 03's previously uncertain MySQL/MariaDB sentence separates an engine's
column-name comparison from foreign-key binding. Pair 04's previously
actionable exporter explanation gives two reasons for one design choice.
Pair 05's previously actionable cache sentence needs both input types to
explain one solver decision. These examples cross the old grade boundary, so
the proposed distinction is not a reliable selector for it.

This does not change [#349](https://github.com/stokaro/unswell/issues/349)
qualification. The November 7 complete review still has 123 frozen events
and remains below its 80% FULL-event recall and 85% accepted actual-finding
gates. The existing baseline has no overlapping finding for 76 of the 80
missed events, so selecting or suppressing existing findings cannot recover
most misses. The next candidate should address missing source-local
diagnostics, with its own frozen hypothesis and full-stream safety check.

## Reproduce

From the repository root, run:

```sh
python3 research/reviews/2026-11-08-long-sentence-dependency/prepare.py \
  --output /tmp/unswell-long-sentence-packet
python3 research/reviews/2026-11-08-long-sentence-dependency/evaluate.py \
  --packet-dir /tmp/unswell-long-sentence-packet
python3 -m unittest discover \
  -s research/reviews/2026-11-08-long-sentence-dependency -p 'test_*.py'
```

The preparer rejects a preexisting output directory and verifies the frozen
inputs, all 36 original source pages, and the exact UTF-8 target spans. The
evaluator checks the packet, sources, key, review seal, pair identities, and
the sole FULL-event witness. The negative tests change a review decision or
source page and verify rejection. The commands are offline and make no model
calls.
