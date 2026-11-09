# Cue-enriched, page-disjoint source-role transfer

Frozen before search execution. This challenge set targets the selector's
stated admission cues. It is not a prevalence sample, a complete-page recall
estimate, or #349 product qualification. The assistant will label original
source text before executing the unchanged selector from commit `42dd619`.

## Fixed search and selection

Run GitHub code search on November 9, 2026 with `per_page=30`,
`sort=indexed`, and `order=desc`, in this order:

1. `"we hope this guide" extension:md`
2. `"we hope this" extension:md`
3. `"let's explore" extension:md`
4. `"this is the place to start" extension:md`

Preserve each raw response and its SHA-256. For each query, take the first five
results from distinct public repositories not already selected by an earlier
query. Exclude repositories whose final path component is `ptah`,
`containerd`, `prometheus`, `tokio`, `pydantic`, `bat`, `ripgrep`, `imgui`,
`fzf`, `curl`, `marked`, `etcd`, or `buildkit`, and the repositories in
`2026-11-08-source-role-rescue` and the first page-disjoint transfer. Skip a
result only when its repository or path duplicates an earlier selection or it
matches an exclusion; continue to the next result until five are selected or
the 30-result page ends. No content-based replacement is allowed.

For each selected file, capture the default branch's current commit SHA and
fetch that exact revision. Record source SHA-256, Git blob SHA-1, fetch time,
and GitHub URL. If the file is missing, not UTF-8, or over 200 KiB, record an
omission without replacement. Keep exact source bytes outside the public repo;
commit only the manifest, selected original quotes, and labels.

## Targets and review

Inventory every LF-delimited line containing the query phrase (case
insensitive). Retain the first three matching lines per selected file when
the line is at most 500 UTF-8 bytes and does not start, after whitespace,
with `#`, `>`, `-`, `*`, `+`, `|`, backtick, `<`, `[`, or `!`. Record original
byte start/end and complete quote. A target crossing a protected Markdown
span is `withhold`, not silently deleted. Do not add or replace targets after
seeing selector decisions.

The assistant labels every inventoried target `admit`, `withhold`, or
`abstain` against the full source page and neighboring paragraph. `Admit`
means that deleting the target would preserve technical and editorial meaning
because it is empty document or authorial narration. `Withhold` means it
contains a concrete step, safety condition, definition, fact, comparison,
meaningful transition, or deliberate author voice. `Abstain` means the role
is unresolved. Freeze labels and source identities in a signed commit before
running the selector. No label is a human annotation.

## Fixed decision rule

Report the full three-way matrix by query and repository. A provisional
within-cue transfer pass needs at least five `admit` targets spanning at
least three repositories, at least 80% admitted positives, and zero admitted
`withhold` or `abstain` targets. Fewer positives make the test underpowered;
any false admission fails this selector version. Regardless of the result,
do not infer whole-page detection, standalone generation, safe supplied edits,
or #349 qualification from this targeted challenge set.
