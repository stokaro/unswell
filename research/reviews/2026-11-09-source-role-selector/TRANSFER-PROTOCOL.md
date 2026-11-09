# Page-disjoint source-role transfer protocol

Frozen before source retrieval or selector execution on these pages. This is a
targeted falsification check for the source-role admission selector, not a full
#349 qualification or a detector recall estimate. The assistant reviews the
source roles; no independent human annotation is claimed.

## Sources

Fetch `README.md` at the current default-branch commit of each repository,
record the commit SHA, Git blob SHA, raw source SHA-256, fetch time, and exact
URL. The list and order are fixed:

1. `nats-io/nats-server`
2. `denoland/deno`
3. `neovim/neovim`
4. `syncthing/syncthing`
5. `astral-sh/ruff`
6. `duckdb/duckdb`
7. `qdrant/qdrant`
8. `grafana/loki`

If a README is unavailable or exceeds 200 KiB, record its omission without
replacement. Do not inspect selector output during annotation.

## Deterministic target inventory

For each exact UTF-8 source, split on LF. A candidate is a single nonempty
line of 30–500 UTF-8 bytes that does not begin, after whitespace, with `#`,
`>`, `-`, `*`, `+`, `|`, backtick, `<`, `[`, `!`, or a digit followed by a period.
It must contain one of these whole-word patterns, case-insensitively:
`I`, `we`, `our`, `let's`, `this README`, `this page`, `this guide`,
`this document`, `in this`, `now`, or `next`. For each source retain the first
10 and last 10 eligible lines in byte order, deduplicating overlap. Include
the original byte start/end, quote, and SHA-256 of the complete source. Do not
manually replace or add targets after viewing them.

Before running the selector, the assistant labels every inventoried line as
`admit`, `withhold`, or `abstain` from the full source page. `Admit` means the
line is only empty authorial or document narration and its removal would
preserve all technical and editorial meaning. `Withhold` means it has a
concrete reader action, technical claim, limitation, safety qualification,
distinct comparison, protected expression, or intentional author voice.
`Abstain` means the role cannot be decided confidently. Save labels, reasons,
and source hashes in a separate signed freeze commit before selector execution.

## Decision rule

Run the exact selector from the first signed commit on all inventoried lines.
Report the full three-way confusion matrix by source, not just positive cases.
A provisional transfer pass needs at least five independently labeled
`admit` lines across at least three repositories, at least 80% recall on those
lines, zero `admit` decisions on any `withhold` or `abstain` line, and no
source/target integrity failure. If the positive count is too small, report
the test as underpowered; if a false admission occurs, retire this selector
version. No threshold can turn this into a full-page or end-to-end #349 pass.
