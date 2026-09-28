# Bounded definition-benefit restatements

The complete exposed Ptah glossary now produces one source-bound diagnostic for
its repeated single-source meaning-consistency benefit in both profiles. The two
highlighted assertions are at bytes 590–723 and 1314–1468. The instructions
between them remain outside the diagnostic.

| Complete pages | Profile | Before findings | After findings | Added | Removed |
| --- | --- | ---: | ---: | ---: | ---: |
| One exposed glossary and three frozen controls | technical | 14 | 15 | 1 | 0 |
| Same pages | strict | 14 | 15 | 1 | 0 |

The added finding is accepted by the OpenAI assistant under ADR 0041.
[All added and removed dispositions](results/dispositions.json) retain that
reviewer's identity. The three confirmation pages contain no target positives
and produce no new findings. Their annotations and source hashes were
[frozen before detector output](confirmation/manifest.json). This is a small
negative-only check from the same project, not repository holdout, independent
human annotation, or a recall measurement. It does not qualify defaults (#26)
or complete the wider contextual-recall work (#349).

## Supported construction

Version 3 of `repetition.repeated-claim` compares complete assertions over the
same opaque definition-source operand and the same cross-page meaning invariant.
It recognizes bounded positive and prevention forms. The file identity retains
case, path, and delimiters. The original sentence window and candidate budget
apply. Multiple recognized assertions form one event with all locations.

The complete glossary also supplies a narrow source-binding bridge: an add
instruction names the source, the immediately following rendering sentence calls
it a map, and a linking instruction precedes the repeated map benefit. This does
not resolve general pronouns or infer that any nearby map is the same source.
Only contribution headings can continue that comparison; other sections,
containers, and excluded quote/code boundaries interrupt it.

The frozen shortened glossary test without the binding remains silent. Different
sources, conditions, modals, exceptions, quantities, spelling invariants, and
actual contribution actions also remain distinct. Other vocabulary, paraphrase
forms, and general semantic equivalence are unsupported.

## Validation and reproduction

[Blackbox controls](../../../testdata/definition-benefit/controls.json) contain the
complete exposed example, two synthetic positives, and eleven synthetic negative
controls. Additional public-API tests cover binding changes, excluded boundaries,
window limits, grouped locations, BOM/CRLF/emphasis mapping, and hidden link
operands. Existing long-reference and unsupported-code budget tests remain
unchanged. They exposed an unnecessary extra mapping pass in the first prototype;
the final implementation shares the claim traversal and only performs extra
mapping work for eligible constructions.

The baseline is `c18213636bcc1bcd5b46769a56c068adb66f0b94`. Build each version
with `CGO_ENABLED=0 go build -trimpath -o /absolute/output ./cmd/unswell`, then run:

```sh
python3 -B research/reviews/2026-09-28-definition-benefit/measure.py \
  --before /absolute/before --after /absolute/after --output /new/output
```

The script verifies the frozen inputs, refuses to overwrite output, scans all four
complete pages with both profiles, and retains complete JSON reports. It compares
actual locations and evidence rather than version-dependent fingerprints.
[The measurement record](results/measurement.json) contains commands, binary and
source hashes, per-run times, and every added/removed finding. Compressed reports
and their hashes are in [results](results/). Source documents are from Ptah under
its [MIT license](PTAH-LICENSE); revision-specific links are in the manifest.

On this Darwin arm64 host, first invocations took 6.98 seconds before and
6.17 seconds after; subsequent strict scans took 0.266 and 0.256 seconds.
Those single observations include process startup and are not a throughput
benchmark. The candidate binary is 34,512 bytes larger. Peak memory was not
measured. No model, network request, or GPU participates in these scans.
