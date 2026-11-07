# Scoped degree evidence for #349

This is an exposed development replay of a narrow `filler.unscoped-assurance`
change. The editor was a Codex assistant under ADR 0041. The original 36 pages,
123 reference events, complete baseline reports, and editorial decisions had
already been seen. This is neither independent confirmation nor a claim that
the #349 quality gates pass. No model was called, no reserved page was opened,
and no label or threshold was changed.

The prior rule treated any `CD` token in an adjacent sentence as evidence for
a capability's degree. A neighboring pronoun `one` therefore suppressed
`perfectly adoptable` in the Ptah adoption page. The revision lets a bare
count ground a claim only in its own sentence; explicit nearby measurements
and criteria still ground it. The prior rule also treated every `if` as a
reason to suppress a degree warning. A reader's stated wish to perform an
operation does not establish the ease of the tool used for it. The revision
permits the narrow `if you want/wish to ...` reader-intent construction only
when the quality claim follows a main-clause boundary. A degree term inside
the hypothetical intent remains excluded. Other conditions, attribution,
negation, protected code, and
measurement boundaries. Rule version 11 records the changed output contract.

The [replay ledger](runs.json) binds the CLI binary and complete compressed
[technical](technical.json.gz) and [strict](strict.json.gz) reports. It used
Go 1.27.3 on darwin/arm64 with `CGO_ENABLED=0`, `-trimpath`, and
`-buildvcs=false`. The existing offline `replay.py` verified all archived
source hashes before running both profiles. Each report has 36 documents,
zero errors, zero skipped rules, and complete source bytes.

The [auditor](analyze.py) checks both report hashes and every document source
hash. It compares every finding against the frozen current-main baseline by
rule and original byte range. All old finding content is unchanged; the nine
old `filler.unscoped-assurance` findings have only the expected version,
identity, and fingerprint changes. The 176 earlier strict editorial decisions
are reused only at those unchanged finding locations and texts. The two new
warnings were read in their full source pages and accepted by the same Codex
assistant as matching the already frozen `unjustified_intensifiers` targets:

| Source | New warning | Frozen event | Editorial reason |
| --- | --- | --- | --- |
| `context/c02.md` bytes 1002–1021 | `perfectly adoptable` | `context/c02-d03` | `perfectly` adds a degree to adoptability without a criterion. |
| `context/c05.md` bytes 301–319 | `easy to use client` | `context/c05-d02` | Ease of use is asserted without a user task or measure; the preceding `if` only states the reader's intent to integrate. |

Both events were previously `missed` and are now FULL on this review. The
complete stream adds exactly two findings and removes none. This changes the
contextual cohort from 13/46 to **15/46 FULL** and from 18/24 to **20/26**
accepted technical findings (strict: 20/27). Whole-page and historical
confirmation cohorts are unchanged. All six cohort/profile cells still fail
the unchanged 80% FULL-event and 85% accepted-delivery thresholds. The
[result](result.json) contains the exact arithmetic; the negative test rejects
a changed new-finding range even if the report hash is updated.

The [constructed-control screen](check_controls.py) replayed the 18 already
frozen external controls with the candidate CLI. No case changed its warning
locations or messages. This is a close-control check, not a population
precision estimate. The screened binary SHA-256 was
`9c24900190a1fef0d215869b6397df350bc6df254d171ce23659d684f49d96e4`.

Reproduce the audit offline from the repository root:

```sh
python3 -B research/reviews/2026-11-07-scoped-degree/analyze.py --check
python3 -B -m unittest discover -s research/reviews/2026-11-07-scoped-degree -p 'test_*.py'
python3 -B research/reviews/2026-11-07-scoped-degree/check_controls.py --binary /path/to/candidate-unswell
```

This does not support widening the adverb vocabulary from these pages. TETRA's
professional edit pairs are academic-paper revisions, not Unswell
`needs_revision` or technical-safety labels. #349 still needs new operators
across several missed categories and selective reduction of existing warning
noise, followed by an untouched prospective confirmation.
