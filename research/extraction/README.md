# Literal selection on Ptah

Issue [#63](https://github.com/stokaro/unswell/issues/63) was observed on Ptah commit
`8873cca616bcd021e5d6588037076fc9d7001c1a`. Keep that source snapshot fixed when
comparing selection policies. Extract an archive from an existing Ptah clone into
a new directory; do not run its programs or alter the working checkout.

The [broad policy](ptah-broad.yaml) uses the technical profile's normal context
selection. The [focused policy](ptah-focused.yaml) adds named exceptions for
workflow shell values, SQL values in three package directories, and two identifier
fixtures. Comments and neighboring messages remain selected. These examples do
not classify every embedded language or supply a recommended project-wide policy.

Run both policies with the same compiled Unswell binary, source snapshot, and
resource limit. For each policy, save the JSON report, terminal report, and exit
code. A 30-second scan of this snapshot can time out; the comparison uses 10 minutes.
Replace the absolute paths below with those of the source archive and this checkout:

```sh
unswell check /path/to/ptah-snapshot --timeout 10m \
  --project-root /path/to/ptah-snapshot \
  --config /path/to/unswell/research/extraction/ptah-broad.yaml \
  --allow-config-outside-root \
  --report json:broad.json --report text:broad.txt

unswell check /path/to/ptah-snapshot --timeout 10m \
  --project-root /path/to/ptah-snapshot \
  --config /path/to/unswell/research/extraction/ptah-focused.yaml \
  --allow-config-outside-root \
  --report json:focused.json --report text:focused.txt
```

An incomplete result must remain visible. This snapshot also contains the Bash
parser failure tracked by [#62](https://github.com/stokaro/unswell/issues/62).
The broad policy additionally selects non-English SQL identifier fixtures, which
can fail English applicability. Neither case is a reason to rewrite source or
remove the parser and applicability checks.

Compare selected source hashes, findings, exclusions and their source ranges,
prose counts, errors, and manifest identities. A newly excluded literal must have
a recorded reason. Fewer findings measure a selection change, not improved prose,
a false-positive rate, or an authorship percentage. Ptah's repository-level origin
is maintainer-reported; these inputs have no independent editorial labels.

The root CLI fixtures and blackbox extraction tests cover invalid source encoding,
decoded non-UTF-8 bytes, valid escaped Unicode, data-only input, and explicit SQL,
JSON, script, identifier, and protocol exceptions alongside checked prose.

## Recorded comparison

The [saved comparison](ptah-comparison.json) records the September 9, 2026 runs on
the pinned snapshot. The initial broad run used Unswell commit
`bc627f079acce48963662a9d3a914fc1718071c0`. Both subsequent runs used the same binary
built from `d4eae11f524c33d997bd47e8605f577405eca929`. The technical profile, rules,
and NLP identities were unchanged. No Ptah source was edited or executed.

| Run | Documents analyzed | Prose words | Findings | File errors | Exit |
| --- | ---: | ---: | ---: | ---: | ---: |
| Before fix, broad | 4,064 | 2,546,834 | 10,992 | 5 | 2 |
| After fix, broad | 4,067 | 2,549,315 | 10,996 | 2 | 2 |
| After fix, focused | 4,068 | 2,545,831 | 10,957 | 1 | 2 |

Every previously analyzed document retained the same source hash, document
summary, exclusions, and findings in the broad run after the fix. Three byte-valued
literals received `non-utf8-literal` exclusions, allowing analysis of the remaining
prose in `core/renderer/identifier_bytes_test.go`,
`internal/agentpatch/agentpatch_test.go`, and
`internal/atlasschema/planfile_hcl_test.go`.

The focused policy changed 39 of the 4,067 common documents and recovered
`internal/parser/parser_test.go` by excluding its non-English identifier fixture.
The other 4,028 common documents retained identical results. The configured
exclusions covered 199 workflow scalars, 559 SQL values, three literals in the
Unicode fixture, and four identifier constants. Each changed document had a
named exception, and every exclusion span was checked against the pinned source.
The Bash grammar error remained in all runs.

The JSON summary retains report hashes, exact exclusions, manifest identities,
counts, and remaining errors. The three full JSON reports, terminal reports, and
exit records are retained in the local experiment directory; they are not bundled
with this summary. Repeating the commands above regenerates reports, though a
different configuration path changes the recorded configuration identity. Config
paths in the summary are represented by their basenames. This comparison measures
source selection and error recovery; it supplies no editorial quality labels.
