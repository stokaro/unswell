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
  --report json:before.json --report text:before.txt

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
