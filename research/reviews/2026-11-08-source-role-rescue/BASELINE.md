# Exact-source baseline on the frozen development packet

The unchanged technical-profile CLI was built from the source-only commit
`4ad716ae96b7f5295f88a711e7f95078a4440ea3` (parent: main
`87f50fe67eaf793758b5e65468b3f85ec46da00e`). The binary SHA-256 was
`37a214bc62d5db17ceff75ab16cb5c4a6694313d8aff8857dfed91e2e2630329`.
Each complete page was supplied as original UTF-8 bytes over stdin with
`--format markdown --profile technical --no-gate --include-source` and a JSON
report. No runtime rule or policy was changed, and no model was called.

| Page | Findings | Assessments | Full source/report |
| --- | ---: | ---: | --- |
| Animal-AI | 5 | 269 | [animal.json.gz](baseline-reports/animal.json.gz) |
| Caddy | 4 | 192 | [caddy.json.gz](baseline-reports/caddy.json.gz) |
| Carbon | 3 | 101 | [carbon.json.gz](baseline-reports/carbon.json.gz) |
| Discourse | 0 | 84 | [discourse.json.gz](baseline-reports/discourse.json.gz) |

All four reports have `complete` status, one full document with the pinned
source SHA-256, no errors, and no gate failure. `baseline.json` pins the exact
compressed report digests and the two relevant finding identities;
`verify_baseline.py` checks the reports against the frozen source packet.

## Source-role review

| Frozen case | Result on current main | Reason |
| --- | --- | --- |
| Animal-AI intro kickoff | Missed | No finding points to the redundant final sentence. |
| Animal-AI repeated example openings | Missed | Neither repeated announcement is diagnosed. |
| Animal-AI empty conclusion | Partial | `filler.evaluative-closure` identifies the outcome-wish sentence, but its advice does not say to remove the now-empty `Conclusion` heading. |
| Seven source-role controls | No empty-narration diagnosis | The baseline does not confuse the reviewed steps, safety condition, concrete links, or Caddy's different claim type with this target category. |
| Carbon enthusiastic introduction | Unresolved | The assistant review did not decide whether to change its deliberate author voice. |

Caddy's unbounded comparison has a `filler.unscoped-assurance` finding at the
correct sentence. This is a separate quality-claim diagnosis and cannot count
as source-role rescue. Findings elsewhere in these four pages have not been
editorially adjudicated, so 12 total findings are **not** an accepted-delivery
estimate. The one partial hit is **not** a FULL-event pass. On this narrow
packet, source-role FULL is 0/3; its only purpose is to expose two misses and
one incomplete suggestion before a candidate is built. These are new
development cases selected by an assistant who knew the original #349
research. They do not change the 57/46/20 frozen cohort results or qualify
#349.
