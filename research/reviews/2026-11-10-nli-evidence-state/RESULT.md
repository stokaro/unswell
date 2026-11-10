# Local NLI evidence-state screen: negative

The v1 frozen runner stopped before the first inference because the published
tokenizer JSON had truncation enabled; `terminal-v1.json` preserves that
operational stop. A separately committed v2 protocol disabled tokenizer
truncation explicitly and retained a hard 512-token-ID limit. The v2 packet
and runner were frozen in commit `425f92d` before the single local CPU run.
All 16 pairs were 74 token IDs or fewer. No hosted model call, fit, threshold
search, retry, product change, or prospective-page access occurred.

| Predeclared measure | Result | Advance requirement |
| --- | ---: | ---: |
| Correct evidence relations | **9/16** | At least 14/16 |
| Correct editorial admission decisions | **9/16** | At least 14/16 |
| Original support-flip pairs fully correct | **1/4** | 4/4 |
| Contradiction cases falsely admitted | **2/4** | 0 |
| Operational errors in v2 | 0 | 0 |

All six no-evidence cases were labeled neutral. Only one of six support cases
was labeled entailment, and only two of four contradictions were labeled
contradiction. The model called the explicit comparison table neutral even
when the table supported the comparison; it likewise called a matching
heading plus table neutral. A usability observation supporting a reader need
was neutral, as was an observation contradicting that need. These are
source-function distinctions that matter to the frozen editorial decisions.

The directional gate failed. Retire this exact NLI checkpoint, target-removed
premise construction, argmax mapping, and admission rule on these exposed
cases. Do not select a new cutoff or adjust the relation key using these
outputs. This is a small synthetic screen, not a real-page transfer test and
not evidence that all NLI models or evidence representations fail. It adds no
FULL-event witness or delivered finding to the 36-page/123-event #349
development cohorts and supplies no meaning-preserving edit. The current
editorial detector and prospective confirmation boundary remain unchanged.

The v2 freeze SHA-256 is
`f8d79e32fb74e347a08af0b5c8d36ff6642002c4938c61ef8ee4e8976797c593`.
The raw logits and argmax decisions in `result-v2.json` have SHA-256
`23a554a6700dc783323298bd85e5639acb61e89204e121c42af6dd6dd01440bd`;
the independent summary has SHA-256
`4348bd830a783c47fd920b472d1ec96ebedc6c5a7f07d0568e731ec819420e39`.
The verifier reconstructs all decisions and counts from the frozen packet,
key, and logits. Its negative tests reject changed logits, a forged relation,
missing cases, and false identity claims. The model and tokenizer are pinned
by revision and SHA-256 but retained outside Git.

Verify the committed evidence without the model:

```sh
python3 -B research/reviews/2026-11-10-nli-evidence-state/prepare.py
python3 -B research/reviews/2026-11-10-nli-evidence-state/verify_v2.py
python3 -B -m unittest discover \
  -s research/reviews/2026-11-10-nli-evidence-state -p 'test_*.py'
```

The next evidence-state hypothesis needs explicit, source-bound propositions
and document roles before an NLI-like relation score can be meaningful. Any
such replacement needs new page-disjoint controls and the unchanged full
editorial-cohort and prospective gates; this result does not authorize an
automatic diagnosis filter.
