# #349 complementary semantic pass: two more events, both gates still fail

This preregistered exploratory pilot added one semantic and rhetorical pass to
the frozen four-page source-function pilot. The exact packet was signed in
commit `82a804e` **before** the new model calls. Its SHA-256 is
`018d92918fbc6e2513e096aea880cf633dcf8a2c2020a1b6876b0108696fdd93`.
The public source revisions and eight unchanged source-only editorial events
are in the [control pilot](../2026-11-08-source-function-pilot/README.md).
The same assistant reviewed the complete new stream against the complete
pages. This is not human annotation or an AI-authorship study.

Four sequential GPT-6-Astra medium calls through the existing Codex
subscription each completed in one attempt, with no tools and valid original
UTF-8 bindings. The [journal](run/run.json) SHA-256 is
`1d1b8387d10767a401ecf7bac8e4e744849e23578af72b7b7260f7594404ef98`.
The model saw only each complete numbered source and the frozen treatment
prompt: no labels, prior responses, or grades. No replacement edits were
requested or supplied.

| Measure | Prior pass | Combined stream | Registered advance |
| --- | ---: | ---: | ---: |
| Fully recalled frozen events | 3/8 | **5/8 (62.5%)** | At least 7/8 |
| Accepted actual delivered diagnoses | 16/17 (94.1%) | **28/33 (84.8%)** | At least 85% |
| New-pass diagnoses alone | — | 12/16 accepted | — |
| Valid complete pages | 4/4 | 4/4 | 4/4 |
| Protected technical counterexamples flagged | 0 | 0 | 0 |
| Supplied edits / edit safety | 0 / not applicable | 0 / not applicable | Separately report |

The new pass supplied FULL diagnoses for restic's absolute claim about user
behavior and Gitea's rhetorical screenshot lead. The latter combines with the
prior pass's accepted grammar diagnosis to cover its frozen composite event.
The prior Caddy plugin finding remains PARTIAL: neither stream explicitly
diagnoses the `ridiculously extensible` intensifier. Gitea's unsupported
`easiest, fastest, and most painless` claim and the Collector's `Performant`
tautology remain MISSED. The new pass also repeated several prior diagnoses
and produced four uncertain findings; all 16 new findings count in the combined
delivery denominator. The [complete review](review.json) records every
disposition and event grade.

The two additional events are a **directional exposed-page gain only**. The
prompt categories were chosen after seeing the prior control's misses. One
Gitea event is inside an HTML summary that the current engine excludes, so
the model's direct source access is not a product comparison. The independent
[result verifier](verify_result.py) checks the frozen packet, prior control,
tool-free traces, source binding, complete finding review, and result
arithmetic. Negative tests reject a tampered packet, unbound quote, tool use,
missing review row, and fabricated positive result. Run:

```sh
make check-complementary-semantic-pilot
```

The registered advance failed both recall and accepted-delivery gates. No
runtime detector, rule, default, or frozen label changed. #349 remains open.
Any further use of this complement needs a new repository-disjoint page set
before the required complete development-cohort evaluation and separate
prospective confirmation.
