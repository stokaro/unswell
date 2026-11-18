# The data of the research report

The [research report](https://unswell.dev/edge/research-report/) on the site
tells what the project tried from September to November 2026 and what it
measured. Its text is `docs/site/src/content/docs/research-report.mdx`. This
directory keeps the report tied to the records.

`make check-research-report` runs the program below without `--write`, and its
tests.

## What is checked

- **Figures.** `figures.py` names a record file and a pointer for each plotted
  value, and holds no number. `build.py --write` reads the values and writes
  `docs/site/src/data/research-report.json`, which the site draws the figures
  and their tables from. Without `--write` the program builds the file again
  and fails when the committed one differs.
- **Numbers in the prose.** `claims.json` lists each number of the text with
  its source: a pointer into a record, or the words of a document that hold
  it. The program reads the source again, and it fails on a number of the
  prose that no claim covers. Years, dates, versions, and the numbers of
  issues and decisions are let through.
- **References.** `references.json` lists each external work with the address
  that was opened, the day it was checked against that page, and how the
  project used it: `ran`, `adapted`, `compared`, `corpus`, or `read`. Every
  external link of the report has to be in the list, and every entry has to be
  cited. The check needs no network; it reads the record of the check.
- **Chart colors.** The series colors in `docs/site/src/styles/unswell.css`
  have to be the same in both dark blocks, stand out from the surfaces, and
  stay apart for a reader with a color-vision deficiency.
- **History.** `data/history.json` freezes the commits by week and the dates
  of the releases at one commit. The program counts them again from git and
  fails when the clone does not hold that commit.

## Limits

A claim proves where a number came from. It cannot prove that the sentence
around it uses the number correctly; that is read by a person or a reviewer.
The geometry of a figure is covered by a self-test of the site, and each
figure carries its numbers as a table so that a wrong drawing can be seen.
