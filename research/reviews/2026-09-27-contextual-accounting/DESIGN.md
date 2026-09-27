# Bounded contextual information accounting

This is an offline candidate for [#349](https://github.com/stokaro/unswell/issues/349).
It has not improved measured recall yet. No model execution is authorized.
The packet, execution dependencies, and evaluation inputs are frozen separately;
preparing or validating them does not launch a model.

## Evidence and hypothesis

The preceding Astra experiment marked all 1,879 units on its 18 valid pages
`reviewed`, with no `uncertain` coverage entries. Its two complete cohorts still
reached only 14/46 and 9/20 full-event recall. One page had 680 units and four
findings. These observations show that coverage declarations do not establish
semantic completeness. They do not prove why the model omitted a defect.

The proposed hypothesis is that bounded attention and explicit editorial
assessments will expose omissions hidden by a page-level list of selected
findings. Each request owns at most 64 existing engine units while retaining the
entire page as context. Each owned unit needs a short account of its informational
contribution and a retain/revise/uncertain decision. A revision requires a linked
diagnostic. The model must review framing clauses even beside useful content.

This changes several interacting factors: request partitioning, prompt, output
contract, and the visible raw source slices. A future comparison measures this
combined candidate, not the causal effect of any one factor. It is a testable
hypothesis, not evidence that longer output or more calls improves recall.

## Source integrity and ownership

The input comes exclusively from the previous frozen 36-page packet. Preparation
checks its identity. It copies full source, format, exclusions, and original
paragraph spans, and adds the exact raw UTF-8 slice of each unit. It does not
parse Markdown again, render text, normalize whitespace, fix quotes, or select
windows using reference labels or previous predictions. The model sees no labels,
baseline diagnostics, dispositions, expected repairs, or cohort names.

Consecutive non-overlapping windows cover every unit once. All units remain
visible for cross-window references. The earliest targeted unit owns a finding;
only that window may emit it. This is a deterministic ownership rule, not semantic
deduplication. A future evaluation must still review duplicates as emitted and
cannot silently discard false positives or infer equivalence from overlap.

Targets use the existing frozen source binder, including exact occurrences and
protected-region checks. The failed `**Deployment reports**` quote from the old
run remains invalid. Providing the correct raw slice makes a valid quotation
possible without relaxing the rule. A synthetic test of that quotation receives
no research credit and never repairs the recorded failed response.

The page assembler rejects missing windows, repeated windows, changed source,
unlinked findings, unexplained decisions, or false completion. It never labels
units outside one request's focus as reviewed in that request's stored result.
Structural validation cannot verify that an explanation is thoughtful or true.

## Execution and acceptance contract

1. The runner and offline collector preserve raw responses, invalid requests,
   command and prompt identities, and multi-request page completeness. The old
   36-call runner cannot run this packet. Tool-disable flags and the narrowly
   defined CLI startup-notice classifier are reused from the retained experiment.
2. The proposed run uses GPT-6-Astra medium, 88 sequential requests, at most ten
   minutes per request and four hours total, with no tools, retries, or repairs.
   The first malformed response or execution failure stops it. The source packet
   contains 14,871,626 request bytes before prompt overhead, including repeated
   page context; the largest request is 649,594 bytes. These are byte counts,
   not token estimates, evidence of context-window support, or cost guarantees.
3. Evaluation retains all 123 exposed reference events and all 36 pages.
   Keep 80% full-event recall and 85% accepted-findings fraction for every required
   cohort. Retain unavailable pages in planned recall denominators, identify them
   separately from observed semantic misses, and fail completeness if any owned
   unit is unavailable. Report partial-page emissions and their dispositions even
   when they do not count as complete-page evidence.
4. Review every emitted diagnostic, uncertainty, duplicate, and suggestion under
   ADR 0041. A contribution assessment alone earns no event credit. Report safety
   independently. Any baseline-plus-model claim needs the full combined stream
   reviewed; adding recall counts does not qualify its precision.
5. Obtain explicit user authorization for the concrete new execution. The
   preceding experiment is terminal; its unused calls do not transfer here.
6. If the complete development screen passes, use the separately frozen three
   prospective pages for confirmation under their declared role. Do not inspect
   their detector output during development or claim the exposed pages establish
   transfer. A failed screen rejects this candidate without shipping it.

The autonomous Go runtime and its gate are unchanged. This work does not restore
routine race, fuzz, or coverage runs, and does not require native-speaker review.
