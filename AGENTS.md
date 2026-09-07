# Unswell development

Use American English. Keep runtime messages and user documentation in English.

The CLI calls the public engine. Rules, scoring, and policy belong to the library;
reporters consume a completed result. Library packages must not read the working
directory or environment, print, exit, start processes, or use the network.

Use original UTF-8 byte spans with an explicit source map. Preserve negation,
numbers, versions, identifiers, and protected code boundaries. A score is an index,
never a probability. Missing capabilities and operational errors cannot pass a gate.

Use standard testing and quicktest imported as `qt`; do not use testify. Maintain
package and exported API comments. Keep the public package ledger current.
Do not add blanket lint exclusions or weaken limits for an individual algorithm.

Run `make check` before publishing. Run `make race` for concurrency changes.
Every repository policy check needs a negative test proving it rejects a violation.
Document alpha limitations and preserve later requirements in the roadmap.

Docker work must use an explicit remote context according to the user's global
instructions. Remove only resources created for this task.
