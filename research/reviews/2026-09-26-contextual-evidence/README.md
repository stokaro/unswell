# Whole-page contextual review with a separate model

Status: prepared offline; **no model requests have run**. This is a research
candidate for the unresolved recall goal, not an integrated or qualified model.

The current engine still diagnoses only 1/20 frozen defects on the four recent
Ptah follow-up pages. Narrow construction extensions and earlier compact-model
and LLM experiments have not established adequate transfer. The next candidate
tests a different model, GPT-6-Astra at medium effort, with a complete-page
editorial task and explicit coverage. Changing both the model and request
contract means this is not a causal model-only ablation of the earlier Opus run.

## Fixed scope

Use all 36 complete pages in the preceding outcome-announcement replay: the
26-page audit, six contextual reference pages, and four Ptah follow-up pages.
Keep all 123 frozen defects and all 4,549 paragraph/heading units exported by the
existing engine. Source bytes total 639,627. Do not select just the earlier
28-event subset, snippets containing trigger phrases, or successful pages.

Every request contains its complete source, source hash, existing engine unit
ranges, and protected regions. It contains no editorial labels, baseline scores,
findings, proposed reference repairs, or target event counts. The packet builder
copies selected fields from a retained complete report; it does not implement a
second Markdown parser. Source hashes and the baseline report are bound.

All these pages are exposed development data. The previous exposure correction
for Atlas and migration overview remains in force. A successful screen still
requires a separately selected and labeled complete-page confirmation sample.
This experiment cannot establish population accuracy or lack of model-training
exposure. Review will be attributed to the actual assistant under ADR 0041.

## Execution and isolation

`protocol.json` fixes one sequential attempt per page, at most 36 model requests,
600 seconds per request, and two hours total. No automatic retry, model fallback,
or replacement of failed pages is allowed. The first invalid request stops the
run, preserving its raw output and every unattempted page in the denominator.

The runner uses the existing Codex ChatGPT login; it does not extract credentials
or reassign HOME or CODEX_HOME. It requests ignored user configuration, no project
instruction bytes, no host-skill discovery, disabled tools/plugins/connectors,
and an empty temporary working directory with a read-only sandbox. The flags
come from local `codex exec --help` and the
[official configuration schema](https://developers.openai.com/codex/config-schema.json).
These are requested settings, not a claim that a live startup has been verified.
Unknown startup behavior, tool events, or an incomplete turn fails the trial.

The CLI consumes subscription quota. Dollar telemetry is unavailable and remains
null; no zero-cost claim or hard monetary cap is made. A new model-agent execution
needs explicit permission for the named model and request scope. Offline planning
does not launch Codex or make network requests. The executable requires a separate
authorization flag and a new output directory.

## Evidence contract and decision

Every finding needs an exact quotation bound to an existing unit, a diagnosis,
a reason the change is warranted, a suggestion, and a statement of the technical
meaning to preserve. Unit coverage is mandatory but is not itself evidence of
understanding or diagnostic recall. Protected material may provide context but
cannot be the sole target. No generated edits are applied.

The evaluator requires a disposition for every emitted finding and every frozen
defect. Accepted, rejected, and uncertain findings remain in the precision
denominator. Full-event credit needs an explicit source-specific review and
accepted support at every annotated target. Overlap alone does not create credit;
partial matches remain partial. Source-grounding checks validate consistency,
not English editorial correctness. The final response must match the raw CLI
message, and collection rechecks the trace for forbidden tool activity.

Suggestion safety has its own required judgment and rationale for every finding:
safe, unsafe, uncertain, or not applicable. Correct diagnoses with unsafe advice
remain visible as such; diagnostic acceptance does not imply a safe edit. Safety
counts are reported separately and do not turn this diagnostic development screen
into edit qualification. No generated edit is applied.

The development screen requires complete valid responses for all pages and at
least 80% full-event recall and 85% accepted findings in each of the three
cohorts. Baseline-plus-candidate recall is reported separately; it cannot pass
the standalone screen. A passing screen does not qualify the product, change
the existing gate, authorize deployment, or establish combined-stream precision.
Confirmation must review the complete combined diagnostic stream before shipping.

## Prospective confirmation prepared separately

[Three newly added Ptah pages](confirmation/README.md) have been pinned and
reviewed before detector or model output: 14 definite defects, 12 acceptable
controls, and seven uncertain alternatives. The complete sources, judgments,
and section coverage are frozen. They are not part of the 36-call request packet
or its 123-event denominator, and no additional model calls are authorized by
preparing them. This small same-project cohort needs separate reporting and
cannot alone establish broad transfer.

## Reproduction

```sh
python3 -B -m unittest discover -s research/reviews/2026-09-26-contextual-evidence -p 'test_*.py' -v
python3 -B research/reviews/2026-09-26-contextual-evidence/study.py prepare /new/packet
python3 -B /new/packet/run.py /new/packet
```

The last command only prints the offline plan. After explicit execution approval:

```sh
python3 -B /new/packet/run.py /new/packet --output /new/run --authorize-model-calls
python3 -B research/reviews/2026-09-26-contextual-evidence/evaluate.py /new/packet /new/run --new-review /new/review.json
# Complete the source-specific review; unreviewed rows cannot pass.
python3 -B research/reviews/2026-09-26-contextual-evidence/evaluate.py /new/packet /new/run --review /new/review.json
```

The generated packet freezes the prompt, runner, validator, schema, protocol,
manifest, and all 36 requests. The evaluator and its tests are additionally frozen
before any model execution. Runtime code, rules, scores, and profiles are unchanged.
