# TETRA source import

The Go developer importer preserves complete editor revisions, exact source
identity, and failed source groups for the auxiliary experiment in
[#365](https://github.com/stokaro/unswell/issues/365). This record checks import
integrity. It contains no model result, quality qualification, or recall claim.

The source is TETRA by Masato Mita, Keisuke Sakaguchi, Masato Hagiwara,
Tomoya Mizumoto, Jun Suzuki, and Kentaro Inui:
[Towards Automated Document Revision: Grammatical Error Correction, Fluency
Edits, and Beyond](https://aclanthology.org/2024.bea-1.21/) (BEA 2024).
The [pinned repository](https://github.com/chemicaltree/tetra/tree/054ba8b308b7ed5a3a2aca9619427bc8f8fbd405)
declares CC BY 4.0. Its editor preferences concern academic-paper revision;
they do not supply Unswell's editorial labels or establish technical safety.

## Pinned inventory and observed accounting

[acquisition-manifest.json](acquisition-manifest.json) binds all 194 public
source entries at commit `054ba8b308b7ed5a3a2aca9619427bc8f8fbd405`: the README,
edit-type map, and 192 XML files. Its SHA-256 is
`7893eeed77d4e9390581c3bf6b6d6e63954a232f93414903181bb6a3b6dd4465`.
File hashes and Git blob identities are verified before Go runs. Corpus prose
and the full import remain external; the repository stores the public inventory,
verifier, teaching fixture, and [verification summary](verification.json).

| Accounting item | Count |
| --- | ---: |
| Supplied XML sources | 192 |
| Filename paper groups | 64 |
| Parsed sources, including identity conflicts | 187 |
| XML syntax or unsupported-structure failures | 5 |
| Declared identity conflicts | 6 |
| Quarantined sources, including affected peers | 24 |
| Eligible paper groups | 56 |
| Retained edits from parsed sources | 3,523 |
| Retained edits from eligible groups | 2,984 |

The source failures remain in the artifact with their exact bytes. No identity,
empty correction, or unsupported punctuation is repaired. A successful integrity
verification therefore accompanies importer exit 2 and `complete: false`.
The counts concern the pinned public corpus, not detection performance.

Five paper groups have an XML syntax or declared-identity problem:
`P07-1091`, `P13-1062`, `P14-1008`, `P15-1026`, and `P17-1188`.
Three more have non-whitespace characters outside text/edit leaves:

| Public source | Unowned content |
| --- | --- |
| `original/W14-5132-C.xml` | A comma after an introduction leaf |
| `original/W16-0501-B.xml` | A greater-than sign before the first section |
| `original/W16-0501-C.xml` | A greater-than sign before the first section |
| `original/W17-1603-A.xml` | A greater-than sign after a title leaf |

The earlier local importer read leaf text but omitted non-whitespace root text
and leaf tails. It excluded only the first five groups. This stricter import
excludes all eight groups and preserves the unsupported sources for inspection.
That discrepancy does not retroactively change earlier frozen research artifacts
or turn their results into a result under this import.

## Reproduction

Acquire the exact upstream revision into an explicit source directory. Acquisition
uses the network separately; the following replay uses only local files and a
Go binary. It requires Python 3.9 or newer for the optional independent verifier.
Ordinary Go builds and product tests do not require Python or this corpus.

From the repository root:

```sh
(cd research/annotation && CGO_ENABLED=0 go build -trimpath -o ../../bin/importtetra ./cmd/importtetra)
PYTHONDONTWRITEBYTECODE=1 python3 research/reviews/2026-09-30-tetra-import/replay.py \
  --source-root /path/to/pinned/tetra \
  --binary bin/importtetra \
  --output-dir /path/to/new/replay-directory
(cd research/reviews/2026-09-30-tetra-import && PYTHONDONTWRITEBYTECODE=1 python3 -m unittest -v test_replay.py)
```

The output directory must not exist. It receives the input, complete output,
stderr, and verification record. Keep these source-containing files within the
dataset's allowed uses. Binary hashes depend on compiler and platform; record
them for each run rather than requiring another platform to match the observed
binary. Input and source bindings remain fixed.

`replay.py` verifies the complete acquisition inventory, runs the actual Go
command, and independently reconstructs every supported source with Python's
ElementTree. It checks source membership, raw bytes, metadata, every section,
edit, UTF-8 range, missing/empty fields, unchanged edits, group quarantine,
counts, completion, and CLI exit. Negative checks prove rejection of a changed
range, source hash, group membership, omitted file/edit, invented task label,
and omitted peer quarantine. No model is loaded or called.

The importer and its supported XML contract are documented
[here](../../annotation/internal/tetra/README.md). Go blackbox tests include a
readable teaching fixture and independent golden, malformed and unsupported XML,
identity conflicts, hashes, deterministic ordering, strict JSON, cancellation,
and writer failures.

## Remaining experiment

The importer supplies editor-preference pairs with complete source groups.
It does not establish whether a change preserves a technical contract, whether
it would earn `needs_revision`, or whether transferred features improve
contextual recall. Training, semantic validity, controlled ablations, frozen
full-event evaluation, and an accept/reject decision remain separate parts of
#365. The product engine, MCP results, rules, and default gate are unchanged.
