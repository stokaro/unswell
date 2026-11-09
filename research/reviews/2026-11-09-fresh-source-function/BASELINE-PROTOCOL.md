# Current-main transfer baseline on sealed pages

The source-only v2 labels were sealed in commit `84fd42e` before this
baseline. Run the unchanged public CLI from base main `14722f0` on each of
the nine complete pinned source pages with the technical profile and no gate.
Use `--stdin --filename page-NNN.md --profile technical --no-gate --report
json:<path>`. Preserve complete report JSON and stderr in the local artifact
directory. Do not select only paragraphs containing search cues.

Match a delivered finding to a positive only when its original-byte span
overlaps the labeled span and its diagnosis addresses the recorded reader
burden. A location hit alone is not FULL. Inspect every delivered finding for
acceptance, not only hits near positives. Controls require source-context
review of any overlapping finding. This is an exposed research baseline, not
product qualification or a prospective confirmation. Do not use its output
to revise sealed source-only labels.
