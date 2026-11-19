# Saved reports

All writers receive the same completed `unswell.RunResult`; they never reopen
source files or execute rules. `unswell report` validates a saved JSON result
against the embedded [versioned schema](../report/schema.json), including required
fields, and rejects contradictory completion states. Schema loading is offline.

JSON and SARIF preserve every finding regardless of display filters. The schema
is generated from the public Go types by `cmd/genschema`; `make schema` detects
drift. SARIF output is validated against the vendored official OASIS 2.1.0 errata
01 schema, and editorial scores are not mapped to security severity. Columns use
Unicode code points (byte spans remain original UTF-8 offsets). Related
occurrences retain their own locations.

HTML is one file with embedded styling and optional filtering script; it remains
readable without JavaScript. Source and paths pass through Go's HTML escaping;
user text never becomes trusted HTML. Text output removes terminal controls, and
Markdown escapes table and formatting characters. All writers return I/O errors.

Source text is omitted unless `--include-source` is selected: without it, reports
still retain paths, coordinates, messages, metrics and score traces. With it,
saved reports can show highlighted paragraphs after originals change or disappear.

For each Markdown or MDX page, the text, Markdown, and JSON reports carry the
[AI-sounding index](ai-sounding-index.md) with the places of its marks. A mark
holds its position and no source text, except that a `stock-word` mark holds its
word of the fixed list. A run of sentences is one mark: it holds the number of
its sentences.

Source [suppressions](suppressions.md) retain their reasons and audit records even
without source inclusion. Raw findings link to the permissions that cover them.
SARIF records accepted `inSource` suppressions with reasons. HTML, Markdown, and
text identify permitted findings and show raw and effective unit scores. A
reporter never resolves permissions again: a saved result contains those
calculations.

The CLI writes regular-file reports through temporary files in the destination
directory and renames them after successful serialization; it rejects duplicate
destinations, direct input overwrites, symlink destinations and existing hard-link
aliases to inputs. When a requested output fails, the exit code is 2.

[SARIF output and consumers](sarif.md) records how the SARIF report is checked
against the published schema and against the reference consumer.

## Verified formats and platforms

The end-to-end suite scans every input format in `document.Formats()`; a format
without an executed scenario fails its own test.

One analysis writes all five reports named by `report.Formats()`. The saved JSON
decodes through `report.Read`, which accepts at most `report.MaxBytes`; the
SARIF document validates against the published schema, and the text, Markdown and
HTML reports carry the finding.

The native CI job runs this package on Linux, macOS and Windows: CI checks the
formats on each of them, not on one machine.
