# #349 unit-relation V1 stopped at its first validation error

The exact source-only labels, prompts, schemas, model, effort, nine-call limit,
and failure policy were frozen in signed commit `697e893` before execution.
The packet SHA-256 is
`2d63d69607562a582a7540f541dfd707ea07a12e05baca4819f774cf76ca3827`.
Four public README pages contain 388 eligible units and five frozen assistant
review events. This is selected development evidence, not prospective
confirmation, a human annotation, or an AI-authorship study.

One GPT-6-Astra medium call completed and returned 88 records for the first
Fastfetch chunk, with no tool use. The response obeyed the JSON schema and
unit count, but `units[0].function` was `Title`. The frozen validator allowed
only eight function categories, while the frozen prompt did not tell the model
that vocabulary. The first call was recorded as failed, and the run stopped
before all eight remaining calls. Nothing was retried, edited in place, or
scored as editorial quality. The raw response, request, trace, and journal are
preserved under `run/`; `terminal.json` reports the exact stop.

The first response still provides a bounded observation: its relation field
marked both frozen Fastfetch targets in that first chunk, ordinals 70 and 71,
as `unestablished`. This is source-location exposure only. No second-pass
verdict, accepted diagnosis, complete-page recall, delivery fraction, or edit
safety was measured. The third Fastfetch target and all other pages remained
unexecuted.

The V1 run is terminal. A new version may reuse the immutable first response
under a corrected, separately frozen contract, but must not call the same
request again or reinterpret this run as complete.

Reproduce the source and first-error checks:

```sh
python3 -B research/reviews/2026-11-08-unit-relation-pilot/verify_terminal.py --check
python3 -B -m unittest discover -s research/reviews/2026-11-08-unit-relation-pilot -p 'test_*.py'
```
