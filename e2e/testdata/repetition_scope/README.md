# Repetition scope fixtures

`changes.md.txt` retains the repeated Literal entry and intervening entries from
pydantic/pydantic HISTORY.md at 00a128a3609dac82dfe0cdb4200bbf2011aa5f83,
release 0.31. The #999 entry and release 0.30 entry are negative controls added
for this test. `LICENSE.pydantic__pydantic` retains the upstream MIT notice.

`tables.md.txt` retains the equal capability wording from stokaro/ptah
`docs/site/src/content/docs/schema/export.mdx` at
654eae5591392278e6c8bce8e54737f780766f19, in two independently useful provider
rows. `LICENSE.ptah` retains the upstream MIT notice. These are the source cases
p014-d01 and the p007 table control in the frozen full-page audit.

`variants.md.txt` is a synthetic control for the installation and repeated
release-credit structures observed in p017 and p014. Each introduction has a
different audience or release scope. The expected result is silence, not a
blanket exclusion of changelogs or tables. Root blackbox tests also require
within-section repetition and a repeated sentence inside one table cell to fire.

The CLI applies CRLF and BOM before checking. Golden reports preserve the exact
primary and related positions, including the inline identifier and issue number.
