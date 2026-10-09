# Partially observed source reviews

`annotation_contract.normalize_review` converts a source review into separate
necessity and intent labels, masks, and exact original UTF-8 edit spans. It uses
only the Python standard library. It calls no annotator and fits no model.

The eight editorial intents are nonexclusive. Naming `wordiness` does not
declare `empty_framing` absent. An unmentioned intent stays `unknown`, with a
`None` loss label and a false loss mask. A known negative requires an explicit
`not_necessary` decision and its own reason.

```python
from annotation_contract import normalize_review

result = normalize_review(
    {
        "necessity": {
            "status": "necessary",
            "reason": "The subject and verb disagree.",
        },
        "edits": [
            {
                "quote": "settings starts",
                "replacement": "settings start",
                "intents": ["fluency"],
                "status": "necessary",
                "reason": "The plural subject takes start.",
            }
        ],
    },
    source="Make sure your debugger settings starts the executable correctly.",
    source_start=100,
)
assert result["intent_loss_labels"]["fluency"] == 1
assert result["intent_loss_labels"]["wordiness"] is None
```

Necessity uses `necessary`, `retain`, `optional`, and `unknown`. The first two
supply binary labels 1 and 0; the others remain masked. Intent decisions use
`necessary`, `not_necessary`, `optional`, and `unknown`. A whole-unit retention
decision does not invent negative labels for every intent.

Each edit has an exact nonempty `quote`, one or more `intents`, a separate
`necessary` or `optional` status, and a reason. `replacement` is optional and
may be `None` when a diagnostic has no safe proposed rewrite. A repeated quote
requires a zero-based `occurrence`. Overlapping edits must be represented
together. Protected spans are absolute byte ranges supplied by the existing
engine; this adapter contains no Markdown parser. A necessary decision needs a
localized mandatory target, and malformed or conflicting responses fail.

One unit can have required and optional edits of the same intent. The mandatory
edit establishes the positive intent label without making the optional repair
mandatory. Review the revised text separately: a successful repair does not
automatically make the after text a clean example. Unmarked tokens remain
unknown rather than background negatives.

A trainer must use only true mask entries and skip an intent-loss row with no
observed dimensions. It must not replace `None` with zero before computing the
loss. Global necessity can supply a separate binary loss while all intent
dimensions remain unknown. Reasons, annotation identities, and teacher outputs
are training metadata, not inference inputs.

This adapter does not decide whether a review is semantically correct. A valid
quote and response schema do not prove necessity or technical preservation.
The caller records source hashes, engine units, context, reviewer identity, and
annotation provenance and freezes any training protocol separately.

Run the public-contract regression tests from the repository root:

```bash
make check-source-review-contract
```

The [completed acquisition study](../../../docs/research/partial-source-labels.md)
records the historical default-negative problem and the separately reviewed
100-case packet. Earlier frozen experiments retain their original annotations
and outcomes.
