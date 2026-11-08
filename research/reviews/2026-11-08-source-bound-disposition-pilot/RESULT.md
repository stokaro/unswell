# #349 paired disposition pilot: no recall gain

The frozen packet SHA-256 is
`66b31ff2408fd835de17c37f8fc56ca3ccd41e9836024d218019e148ae136598`.
All nine tool-free calls completed once, in the preregistered order, without a
schema, source-binding, trace, or timeout error. The complete raw requests,
responses, traces, and bound findings are in [run](run), with hashes in
[run.json](run/run.json). No response was repaired or retried.

| Arm | FULL / frozen events | Accepted / delivered | Protected controls flagged | Uncertain units |
| --- | ---: | ---: | ---: | ---: |
| Free-form control | 3 / 4 | 3 / 3 | 0 / 5 | Not requested |
| Every-unit disposition | 3 / 4 | 3 / 3 | 0 / 5 | 3 / 128 |

The treatment did account for every eligible unit, but **did not find an
additional frozen event**. Both arms found the two Dagger events and the
Prefect `any data process` event. Both missed the Prefect recovery event.
Airbyte produced no findings in either arm. The preregistered directional
advance rule failed because treatment gained zero FULL events over control.
The [machine result](result.json) and [complete-stream review](review.json)
preserve all denominators and reviewer identity. No edits were supplied, so
edit safety was not measured.

The frozen Prefect recovery event has a substantive label ambiguity: "you can
build ... pipelines that ... recover from unexpected changes" can be read as a
capability rather than a promise to recover from every unexpected change. This
was noticed only during outcome review. The label and 4-event denominator
were not changed. If that event were excluded as a sensitivity check, both
arms would be 3/3 FULL, so the conclusion of no between-arm gain is unchanged.
This is not a post-hoc rescore or a replacement qualification denominator.

This selected three-page pilot is too small and source-specific to estimate
general recall or precision. Its 3/3 accepted findings in each arm do not
establish the 85% gate at the required scale. The result argues against
expanding the expensive every-unit-disposition path as the next #349 method.
The next development design should target missed source-grounded event types
directly and use a larger independently reviewed, prospectively frozen cohort
before any product-quality claim. The 36-page, 123-event development gate and
separate prospective confirmation remain open.
