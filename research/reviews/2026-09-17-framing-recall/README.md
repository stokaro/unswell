# Contextual framing and certainty: a bounded improvement

This change detects seven previously missed defects on the development pages.
On the 12 sampled Ptah pages, event credit rises from **1/27 to 8/27** in both
technical and strict. Historical development pages remain **0/21**; the two
exposed anchors remain **4/9**. All gates still pass.

The separate confirmation result is negative: **1/11 before and after**, with
no additional findings on nine pages. The implementation addresses known
construction families but does not establish improved recall on new pages.
Do not use the development gain to claim that technical prose is now thoroughly
checked. Ten confirmation defects and 45 development defects remain missed.

## Changes and boundaries

Version 2 of `filler.evaluative-closure` recognizes bare whole-point/design
announcements and short anaphoric assertions of value or verification. Version 2
of `filler.document-justification` adds document parts described as earning their
place. The new `filler.unscoped-assurance` recognizes intentional quality
announcements, nearly-always-right judgments and universal learning promises.

These are bounded clause constructions using existing tokens, POS annotations,
protected spans, budgets and source mapping. Complements that name a purpose,
condition or measurement prevent the added matches. An assurance followed by a
colon is left alone because the following clause may supply its mechanism.
A clause over 48 tokens is outside this matcher. These boundaries account for
some remaining misses, and general paraphrases are still unsupported.

The new rule has experimental warning defaults: weight 12, cap 24, no rule
prohibition. Existing weights and gate thresholds are unchanged. The diagnostics
request an editorial review and do not determine whether a claim is false or who
wrote it. No default promotion or model qualification is claimed.

The implementation and unit controls were written before confirmation outputs
were opened. A lint-only decomposition retained their behavior. No matcher was
tuned from confirmation outputs. Development diagnostic spans sometimes include
an adjoining conjunction or Markdown admonition marker; the saved locations
show that limitation instead of hiding it.

## Evidence

- [Protocol](protocol.md), [input freeze](input-freeze.json) and
  [annotation freeze](annotation-freeze.json).
- [Results](RESULTS.md): both profiles, every page, complete scan time and memory.
- [Added/removed diagnostic review](CHANGES.md) and
  [machine-readable dispositions](dispositions.json).
- [Confirmation misses](CONFIRMATION-MISSES.md): exact passages, reasons and repairs.
- [Confirmation annotations](confirmation/annotations.json): 11 defects,
  six uncertain judgments, ten acceptable controls and complete prose coverage.
- Eight compressed JSON reports under `reports/`, with binary hashes, engine
  revisions, command lines, host and per-process resource measurements.

Before is `cc79256188310f9d63a86f259044df4c3a27af1d`; after is
`dbae2c86af6b1906ecf6f592c1dcc6c40af4ba13`. The reproduced development baseline
matches the earlier saved report exactly. The code revision predates these result
files; subsequent changes in this PR concern evaluation tooling and evidence.

## Review and selection limits

The reviewer was the same Codex assistant who implemented the rules. The
maintainer accepts that editorial review under ADR 0041. This is neither a human
annotation nor an independent or blinded experiment, and no agreement statistic
is invented. Uncertain labels remain unresolved and receive no positive credit.

Confirmation sources were selected by pinned source identity, before
implementation, after excluding all 26 development sources. One page was selected
per Ptah format/length cell and one historical page per length cell. The long MDX
cell had only one remaining candidate. Source files differ; repositories can
overlap. These nine pages are not a representative population estimate.

The confirmation rubric covers the three targeted families, whereas the earlier
development annotation covers seven defect categories. Their recall fractions
must not be pooled. Complete source files are retained with their notices and
rights records. Imported components and text inside images were not expanded;
source-file coverage does not establish rendered-site coverage.

Every added or removed diagnostic was reviewed. Retained development findings
inherit the earlier individual dispositions. On confirmation pages, the evaluator
checks all possible source-overlap credits for the targeted events; it does not
relabel the other existing warnings or estimate their precision.

## Reproduce

The standard-library Python scripts are optional research tools, not runtime or
Go-training dependencies. From the repository root:

```sh
record=research/reviews/2026-09-17-framing-recall
python3 -B "$record/tools/render.py"
python3 -B -m unittest discover -s "$record/tools" -p 'test_*.py' -v
python3 -B "$record/tools/measure.py" --binary /path/to/unswell \
  --set confirmation --output /new/output/directory
```

Use `--set development` for the earlier 26-page archive. The output directory must
be new. Selectors and annotations are not regenerated from diagnostics. To replay
selection, run `tools/select_inputs.py --ptah-sources PATH --output NEW_PATH`
with the Ptah source snapshot named in the manifest. Artifact comparison verifies
hashes, original UTF-8 spans, notices, source disjointness, complete documents,
policy identity, every diagnostic delta and semantic eligibility of event credit.

Blackbox Go tests cover positive constructions, negation, questions, numerical
scope, protected code, operational complements, source mapping, exemptions,
contexts and occurrence limits. Existing budget/cancellation tests include the new
rule. The source-bound rewrite fixture preserves the concrete error behavior.

Further work belongs to recall improvement on broader contextual constructions
and to repetition scope in #300. The new confirmation misses provide concrete
next candidates; another iteration needs new confirmation pages.
