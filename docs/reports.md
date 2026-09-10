# Saved reports

All writers receive the same completed `unswell.RunResult`. They never reopen
source files or execute rules. `unswell report` validates a saved JSON result
against the embedded [versioned schema](../report/schema.json), including required
fields, and rejects contradictory completion states. Schema loading is offline.

JSON and SARIF preserve every finding regardless of display filters. The schema
is generated from the public Go types by `cmd/genschema`; `make schema` detects
drift. SARIF output is validated against the vendored official OASIS 2.1.0 errata
01 schema. Columns use Unicode code points, while byte spans remain original
UTF-8 offsets. Related occurrences retain their own locations. Editorial scores
are not mapped to security severity.

HTML is one file with embedded styling and optional filtering script. It remains
readable without JavaScript. Source and paths pass through Go's HTML escaping;
user text never becomes trusted HTML. Text output removes terminal controls.
Markdown escapes table and formatting characters. All writers return I/O errors.

Source text is omitted unless `--include-source` is selected. Without it, reports
still retain paths, coordinates, messages, metrics and score traces. With it,
saved reports can show highlighted paragraphs after originals change or disappear.

Source [suppressions](suppressions.md) retain their reasons and audit records even
without source inclusion. Raw findings link to the permissions that cover them.
SARIF records accepted `inSource` suppressions with reasons. HTML, Markdown, and
text identify permitted findings and show raw and effective unit scores. A saved
result contains those calculations; a reporter never resolves permissions again.

The CLI writes regular-file reports through temporary files in the destination
directory and renames them after successful serialization. It rejects duplicate
destinations, direct input overwrites, symlink destinations and existing hard-link
aliases to inputs. A failed requested output gives exit code 2.

[SARIF output and consumers](sarif.md) records how the SARIF report is checked
against the published schema and against the reference consumer.

## Verified formats and platforms

The end-to-end suite scans every input format in `document.Formats()`. A format
without an executed scenario fails its own test.

One analysis writes all five reports named by `report.Formats()`. The saved JSON
decodes through `report.Read`. The SARIF document validates against the
published schema. The text, Markdown and HTML reports carry the finding.

The native CI job runs this package on Linux, macOS and Windows. CI checks the
formats on each of them, not on one machine.
