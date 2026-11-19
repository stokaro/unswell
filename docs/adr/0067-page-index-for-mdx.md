# ADR 0067: The page index reads MDX pages

Status: accepted

## Context

[ADR 0059](0059-page-index-and-fix.md) gave every Markdown page an
AI-sounding index, and [ADR 0060](0060-rhythm-reading.md) gave it a second
reading. A document of any other format got none, because the reference that
the weights were fitted on holds Markdown files.

An MDX page is Markdown with components. The rules already read it: the
extractor takes the prose, and it leaves the tags, the attributes, the
expressions in braces, and the `import` and `export` lines out. So a
documentation site written in MDX got findings and no index
([issue 631](https://github.com/stokaro/unswell/issues/631)).

The pages of this site are MDX. Their index was read once from a copy of each
page without its `import` lines and its component tags, saved as Markdown. A
normal run showed nothing for them.

## Decision

1. An MDX page of 150 prose words or more gets the index. It is counted from
   the same blocks that the rules read. The text inside a component is prose
   and counts. A tag, an attribute, an expression, and an `import` or
   `export` line are code, and a dash or an arrow there is no mark.

2. A component ends a run of sentences, as a heading, a list, and a code
   block do. The sentences inside one component can form a run of their own.

3. The weights and the levels are those of Markdown pages. The two readings
   and the minimum of 150 words stay as they are.

4. `unswell fix` changes MDX pages too. A replacement stands where its mark
   is, and a mark is in prose, so the code of the page keeps its bytes. The
   command analyzes the page again after each pass and does not write a page
   whose analysis fails.

5. Other formats still get no index. The text report counts "pages" in its
   summary line, where it said "Markdown pages".

## Consequences

`unswell check` names an MDX page above `low` with its marks, the JSON report
holds `ai_sounding` for an MDX document, and the MCP tool returns it for an
MDX source. A project that reads the JSON report finds the field on more
documents than before.

## What was compared

The eight pages of this site that are long enough, at commit `843cd43c`,
were read both ways: as MDX, and as the copy without `import` lines and
component tags.

| Page | As MDX | As the copy |
| --- | ---: | ---: |
| `how-it-works` | 71 | 75 |
| `ci-action-mcp` | 63 | 66 |
| `review-and-trim` | 50 | 52 |
| `reports` | 48 | 52 |
| `getting-started` | 42 | 42 |
| `index` | 40 | 42 |
| `configuration` | 36 | 43 |
| `research-report` | 34 | 34 |

The MDX reading is lower by up to 7 points, for two reasons. In the copy the
text of a note joins the paragraphs around it, so a run of sentences goes on
through it; as MDX the run ends at the tag. And the copy loses the text of a
component that is indented by four spaces, which Markdown reads as code; as
MDX that text is prose, and the same marks stand in more words.

## The fix and a real build

The tests hold a page whose attribute, expression, and import keep their
bytes. One check went further, with the build of this site. Its ten pages got
marks in every kind of place: each comma followed by a space became a dash,
"then" became an arrow, and a list with a bold label and an emoji was added.
`unswell fix --write --include-low` removed 567 marks in the eight pages that
have an index. The site built before and after with Astro, and its markup and
link checks passed. The dashes that were left stood in code blocks, in
attributes, and in the strings of expressions.

That is one site and one MDX compiler.

## Not covered

No reference of MDX pages before and after a cleanup was measured, so the
reading of an MDX page rests on the measurement of Markdown pages. Text that
a component takes as an attribute, such as the lines of a terminal mock-up,
is outside the count.
