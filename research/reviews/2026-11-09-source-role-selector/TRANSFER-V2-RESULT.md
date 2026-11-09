# Cue-enriched source-role transfer: selector retired

The fixed searches selected 20 public Markdown pages from 20 repositories.
All default-branch commit SHAs were captured and signed before reading the
files. Exact bytes were fetched at those revisions, and every local Git blob
SHA-1 matched the search result. The four complete search response pages are
archived losslessly in `search_v2/*.json.gz`; their uncompressed SHA-256 values
are in `transfer_v2_selection.json`.

The frozen line inventory selected 12 targets. The assistant labeled one
`admit` and 11 `withhold` against complete pages before running the unchanged
selector from commit `42dd619`. The matrix was:

| Frozen role | Selector admit | Selector withhold | Selector abstain |
| --- | ---: | ---: | ---: |
| Admit | 1 | 0 | 0 |
| Withhold | 5 | 5 | 1 |

The five false admissions are source-bound, not score disagreements:

- `llmware-ai/llmware`: the target also contains an intentional closing
  sign-off, `Happy coding!`.
- `go-nunu/nunu`: the target also makes a product-function and benefit claim.
- `google-research/tuning_playbook`: the target is one wrapped line of a
  substantive sentence about the state of the field.
- `reapp/reapp`: the target starts a sentence about included components that
  continues on the next line.
- `HuolalaTech/page-spy-web`: the target states when the tool is useful and
  introduces the concrete use cases below.

The one correct admission is the empty guide-benefit wish in
`quoid/userscripts`. Its neighboring paragraph contains a link to a separate
development guide, but the selected line itself has no instruction or claim.

The frozen rule fails on **any** false admission, so this selector version is
**retired**. The set also has fewer than five positives across three
repositories, so it cannot estimate within-cue recall. This negative result
overrides the exposed archive's apparent 85.03% numerical threshold crossing.
No runtime rule changed and #349 remains open. Do not repair this version by
changing these labels, narrowing the target set, or tuning a threshold on the
same files. Any replacement needs a new frozen protocol that reasons about
complete source units and mixed technical/editorial content, followed by full
event and delivery review against the unchanged #349 gates.
