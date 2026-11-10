# NLI evidence-state screen v2: explicit no-truncation setup

The first frozen run stopped before ONNX session creation or any inference.
The tokenizer JSON supplied by the pinned upstream revision has truncation
enabled at 512 token IDs. The v1 runner rejected that setting as specified.
Its terminal record is `terminal-v1.json`; it has no scored result and must not
be retried or overwritten.

V2 keeps the same 16 source/target pairs, key, source identities, published
model revision and asset hashes, argmax label mapping, admission mapping,
directional gates, and first-error stop policy from `PROTOCOL.md`. The sole
inference change is calling `tokenizer.no_truncation()` before encoding, then
rejecting any pair over 512 token IDs. A tokenizer-only preflight, without
model inference, found all 16 pairs to be at most 74 token IDs. V2 has one
run, no retry, no fit, and no threshold search. `freeze-v2.json` and a commit
must exist before the first v2 inference.

The v2 result is `result-v2.json`. The independent scorer is `verify_v2.py`.
V1's failure remains terminal and does not contribute a diagnosis, a model
quality estimate, or any change to #349's 123-event denominators.
