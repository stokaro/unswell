# TETRA import

`cmd/importtetra` imports explicitly supplied TETRA source bytes into the
`unswell-tetra-import-v1` research artifact. It uses Go, reads stdin, writes
stdout, and performs no filesystem access, network requests, subprocess calls,
training, or model inference. The product CLI and MCP server do not import it.

TETRA records professional editors' revisions to academic papers. Those
preferences can support an auxiliary revision experiment; they do not establish
Unswell's `needs_revision` label or the technical safety of a change.
See [Mita et al., BEA 2024](https://aclanthology.org/2024.bea-1.21/) and the
[pinned upstream source](https://github.com/chemicaltree/tetra/tree/054ba8b308b7ed5a3a2aca9619427bc8f8fbd405).
The pinned repository declares
[CC BY 4.0](https://github.com/chemicaltree/tetra/blob/054ba8b308b7ed5a3a2aca9619427bc8f8fbd405/README.md#license).
Imported records retain that declaration. Cite the authors and retain attribution
when distributing corpus-derived material.

## Explicit input

The [input schema](input.schema.json) requires the contract version, canonical
repository URL, lowercase 40-character commit, license declaration,
`include_xml`, and an ordered inventory of sources. Each source contains its
relative `original/Pxx-xxxx-A.xml` or `original/Wxx-xxxx-B.xml` path, SHA-256,
Git blob SHA-1, and lossless `xml_base64`. Editors A, B, and C are supported.

Both hashes are checked against the decoded bytes before any pairs are returned.
SHA-1 implements the upstream Git object identity; SHA-256 checks content.
These checks cannot authenticate an acquisition manifest or prove that a file
belongs to the declared commit. The caller must acquire and validate that inventory.
The [pinned replay](../../../reviews/2026-09-30-tetra-import/README.md)
checks all source entries before constructing input.

Unknown or duplicate JSON fields, case aliases, unknown versions, incorrect
hashes, duplicate paths, and invalid paths fail without an import artifact.
Limits are 256 files, 256 KiB per XML source, 8 MiB of decoded sources, 16 MiB of
input JSON, and 64 MiB of output JSON. Cancellation and writer failures return
errors.

## Sections, revisions, and ranges

The importer supports the ordered `title`, `abstract`, and `introduction`
sections and their flat `text`/`edit` leaves. It preserves leaf contents and
joins adjacent leaves with one ASCII space. XML character references are decoded;
whitespace inside a leaf is preserved. XML comments contribute no text.
Sections retain separate original/revised hashes and every same-editor change.
Do not apply an isolated edit while discarding the other changes needed by its
complete revised section.

Each edit retains its upstream before/after text, raw type, normalized type list,
comment, and UTF-8 byte ranges in both reconstructed sections. These are not
positions in the XML file, publication, or an Unswell extracted document.
Any later extraction must establish its own mapping.

Empty corrected text is a deletion; empty original text is an insertion. An
identical correction remains an unchanged observation. Missing type or comment
differs from an explicitly empty value. Duplicate normalized types are removed
and sorted, while the raw attribute remains available.

DTD/entity declarations, namespaces, unknown or repeated attributes, nested
markup, processing instructions inside content, reordered sections, and
non-whitespace text outside leaves are unsupported. Literal tabs and line breaks
inside attribute values are also unsupported because the Go XML decoder does
not apply XML attribute whitespace normalization. Numeric character references
remain supported. Unsupported source is recorded, never silently repaired.

## Group quarantine and exits

Filename identity determines the paper group. Declared XML identity is checked
exactly; trailing spaces, an empty ID, or a conflicting editor are not repaired.
A source failure quarantines every supplied filename version of that paper.
All supplied files stay in the output, including parsed peers and source failures.
`include_xml: true` additionally preserves the exact raw bytes for every record.

File status distinguishes `valid`, `identity_conflict`, `invalid_xml`, and
`quarantined`. Source issue codes distinguish XML syntax errors from unsupported
structure. `source_eligible` means integrity and supported shape passed; it is
not a quality label, permission for every downstream use, or an independent split.
Group related versions before any future data partition.

Exit 0 means every supplied group passed import checks. Exit 2 means a source
group was quarantined or an operational failure occurred. A quarantine still
produces the full accounting artifact with `complete: false`; an operational
failure does not produce a successful artifact. Exit 130 means interruption.
Partial output from a writer failure must not be consumed.
`task_labels_assigned` and `product_qualified` are always false.

## Teaching example

From `research/annotation`:

```sh
mkdir -p ../../artifacts
CGO_ENABLED=0 go build -o ../../bin/importtetra ./cmd/importtetra
../../bin/importtetra < internal/tetra/testdata/tutorial-input.json > ../../artifacts/tetra-tutorial.json
go test ./internal/tetra ./cmd/importtetra
```

Create the output directory before redirecting. The input and
[expected output](testdata/tutorial-output.json) are tested against the readable
[XML fixture](testdata/tutorial.xml). An assistant constructed this example,
including a zero revision declaration; it is not an upstream paper, editor
judgment, or training example. The output golden was independently reconstructed
with the replay verifier before comparison with Go.
