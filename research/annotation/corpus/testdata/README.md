# Ptah source candidates

The `ptah/` directory contains eight unchanged files from `stokaro/ptah` commit
`8873cca616bcd021e5d6588037076fc9d7001c1a`, plus its original root `LICENSE`.
The root license is MIT, copyright Denis Voytyuk 2025, 2026. The frozen snapshot
contains no additional `LICENSE`, `COPYING`, or `NOTICE` files. Each imported file
and retained notice have exact byte counts, SHA-256 hashes, and pinned source links
in [ptah-manifest.json](ptah-manifest.json). Linked resources were not imported.
The Go files are data fixtures and are not compiled or executed by these tests.
Preserve their original bytes; do not format or rewrite them to update a golden.

The maintainer authorized Ptah as an Unswell evaluation subject and described it
as nearly entirely AI-generated. The manifest records that assertion at repository
scope. Unit origin, quality, individual authors, author-language background, and
generation histories remain unknown. The imported text has no human judgments.

The selection covers README, documentation, package documentation, ordinary code
comments, and strings in SQL schema and migration code. Whole-file doc-comment
regions and nine error-string ranges were reviewed against the original source.
Generic comments and strings retain their roles when more specific evidence is
absent. This is an engineering selection for extraction and annotation workflow
testing, not a representative or randomly sampled editorial evaluation corpus.

The current extraction yields 378 unlabeled candidates from eight source documents:

| Kind | Count |
| --- | ---: |
| Fragment | 213 |
| Sentence | 100 |
| Paragraph | 65 |

| Role | Count |
| --- | ---: |
| README | 72 |
| Documentation | 194 |
| Doc comment | 77 |
| Ordinary comment | 23 |
| Error message | 9 |
| Generic string | 3 |

Some sentence and paragraph units overlap, so 378 is not a count of independent
observations. Every source belongs to one connected group pinned to `development`:
Ptah has already been inspected while developing Unswell. Related versions and
rewrites must remain together. Training, calibration, and final-test counts here
are zero. No unseen-generator, author-language, release-note, or quality subgroup
claim follows from this fixture.

Allowed uses in this batch are annotation and source redistribution, based on the
retained notice and maintainer authorization. Future training/evaluation uses need
an explicit updated declaration and their own corpus protocol. This preparation
does not satisfy #22's 5,000 labeled units, two independent human annotators, or
1,000 final-test units.
