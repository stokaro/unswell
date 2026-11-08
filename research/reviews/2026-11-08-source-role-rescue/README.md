# #349 source-role rescue: source-only development review

Status: frozen source-only development packet, November 8, 2026. A Codex
assistant selected and reviewed these four complete public pages before running
the Unswell CLI on them. The assistant had already seen the original #349
development labels and the retrospective operator-triage result, so this packet
is **not** prospective confirmation. No model was called to create this packet.

## Question and boundary

Can a source-role decision distinguish empty narration about a document from a
real task step, safety qualification, author voice, or another kind of claim?
This is a possible triage mechanism for the archived six-stream proposal bank,
not a new authorship detector. A diagnostic must point to the actual redundant
source bytes. A supplied edit is safe only if it preserves instructions,
prerequisites, links, technical scope, and neighboring section structure.

The `review.json` cases were judged on full pages. Positive cases are narrow
editorial criticisms, not permission to delete a whole paragraph. Controls
must not receive the *empty authorial narration* diagnosis; they may have a
different problem. An `unresolved` case cannot be counted as either a positive
or a negative success. The review is by a Codex assistant under ADR 0041, not a
human annotation. A later correction must be a dated amendment, not a rewrite
of these labels.

| Page | Exact upstream revision | Source license |
| --- | --- | --- |
| [Caddy README](https://github.com/caddyserver/caddy/blob/fcfdef43c566a00aa5b7d6c50329b90ffe00e241/README.md) | `fcfdef43c566a00aa5b7d6c50329b90ffe00e241` | Apache-2.0 |
| [Carbon icon guide](https://github.com/carbon-design-system/carbon/blob/f5bdc21a0933baca91875e618f07141ff708cbe7/docs/guides/building-an-icon-library.md) | `f5bdc21a0933baca91875e618f07141ff708cbe7` | Apache-2.0 |
| [Animal-AI YAML guide](https://github.com/Kinds-of-Intelligence-CFI/animal-ai/blob/9d74c9a7a0a8d60861c47058c9e86581bb71eb52/docs/configGuide/YAML-Config-Syntax.md) | `9d74c9a7a0a8d60861c47058c9e86581bb71eb52` | Apache-2.0 |
| [Discourse remote-theme tutorial](https://github.com/discourse/discourse/blob/444e62cc26e96c7da97f08fa45df219f03f60114/docs/developer-guides/docs/07-theme-developer-tutorial/02-remote-theme.md) | `444e62cc26e96c7da97f08fa45df219f03f60114` | GPL-2.0 |

The exact source and license bytes are in `sources/`. The Animal-AI source is
gzip-compressed to preserve upstream trailing spaces without introducing a
repository whitespace failure; verification operates on its decompressed
bytes. `review.json` pins source SHA-256, Git blob IDs, and original UTF-8 byte
spans. Run
`python3 -B verify.py` to check those identities and the review contract.

## Predeclared test

1. Record the current main CLI's findings on all four complete pages before
   changing a runtime rule. Compare diagnosis and exact source location, not
   mere lexical overlap. A warning for Caddy's unsupported comparison cannot
   earn credit for empty authorial narration.
2. A candidate selector must be specified from source-visible roles and tested
   without consulting the frozen editorial grades during a new page's decision.
   Keep the current rule and policy unchanged until that selector passes
   source-role controls and preserves safe advice.
3. The three positive events here all come from one repository. They can only
   falsify the candidate and reveal integration defects; a pass cannot show
   cross-repository transfer. After construction, freeze new page-disjoint
   pages and labels before executing the candidate on them.
4. Even a page-disjoint pass does not change #349's 57/46/20 event denominators
   or replace the complete 36-page replay, the 80% FULL-event recall and 85%
   accepted-delivery gates, suggestion-safety review, or prospective
   confirmation. The archived six-stream bank remains exposed and its five
   unresolved supplied edits remain unresolved.

Fail this narrow screen if any of its controls receives the empty-narration
diagnosis, if a positive is credited without its actual defect, or if advice
deletes a task step, safety qualification, link, or required section content.
Do not repair a failure by changing these labels or picking another threshold
on the same pages.
