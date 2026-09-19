# Count-centered section framing

The maintainer identified an unreported numerical hook in the Ptah provider
guide: “Four lines of a specification decide where your corpus goes,” followed
by “The four lines.” The existing document-announcement warning covered the next
sentence and did not address this construction.

`filler.numbered-section-framing` now reports the opening clause with the heading
as related evidence. Both technical and strict profiles produce one new warning
on the complete pinned page. Name the configuration topic in the introduction
and heading; preserve the endpoint, model, credential, and privacy instructions.
The warning does not count YAML fields or assert that four is factually wrong.

| Input | Before | After | Review |
| --- | --- | --- | --- |
| User-supplied provider page | Missed pair | One paired warning | Actionable: use a configuration heading and direct introduction |
| Four separate complete pages | No target-family finding | No new finding | One broader numerical opening remains missed |

Existing diagnostics are unchanged in location, wording, and evidence. Total
findings across the five pages are 18 → 19 for technical and 21 → 22 for strict.
Every scan completed without errors, rule skips, or abstention. The gate passes
in both phases; its thresholds are unchanged.

The confirmation miss is the opening “Four answers, and the right one depends
on whether you control the writes” on the consistency-mode page. Its following
heading does not repeat the number. It is retained as a miss, and the frozen
candidate was not broadened after inspecting confirmation outputs. This bounded
repair does not establish good recall, population precision, or AI authorship.

Concrete counts remain controls: two credential schemes, two probe strings,
three simultaneous data-race conditions, and a named menu of four Mermaid usage
options. Tests retain explicit format requirements, negation, named task
headings, changed counts, code, quotations, lists, intervening content, and
vocabulary exceptions. Digit and word forms are supported within the catalog's
stated limits. Existing CLI goldens changed only in catalog identity hashes.

The Codex assistant read all four confirmation pages before implementation and
reviewed the new diagnostic afterward under ADR0041. The maintainer supplied the
development example. These are not independent human labels. Review covers only
count-centered framing, not every possible editorial defect on each page.

[inputs.json](inputs.json) records the source identities and deterministic
metadata selection. Earlier reviewed repository/path identities were excluded;
source versions are pinned, with notices retained in `inputs.tar.gz`.
[source-review.json](source-review.json) contains the original spans, repairs,
controls, and whole-page coverage. Input and runtime freezes precede confirmation
outputs; [summary.json](summary.json) is derived from the retained reports.

Build the baseline and candidate commits in [code-freeze.json](code-freeze.json)
with `CGO_ENABLED=0`, `-trimpath`, and an explicit `unswell.BuildCommit`, then run:

```sh
python3 tools/measure.py --binary /path/to/unswell --output /new/report-directory
python3 tools/verify.py
python3 tools/test_evidence.py
```

The replay restores sources outside a Git checkout and checks that all five
pages are present. Six evidence tests reject missing locations, credit for a
neighboring sentence, unreviewed new warnings, changed old diagnostics, and
incomplete scans. Public API and annotated CLI tests cover the production rule.
Research classification revision 10 records it as general style and preserves
all prior classifications. Race, active fuzzing, and coverage remain deferred
to #123. Runtime source files remain unchanged after confirmation.
[validation.json](validation.json) records completed repository checks, CLI/MCP
parity on 1122 documents, and the host-access retry for resource reporting.
