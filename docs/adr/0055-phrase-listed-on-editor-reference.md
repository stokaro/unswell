# ADR 0055: A phrase can be listed on what editors removed

Status: proposed on November 12, 2026. It takes effect when the maintainer
merges it. It adds a second route to the listing of
[ADR 0047](0047-qualified-defaults-and-assistant-review.md). Decision 3 was
replaced on the same day, after the first release with the opener.

## Context

ADR 0047 lists a phrase in a default rule when two assistant reviewers
accepted its findings. [ADR 0052](0052-claims-follow-human-references.md) then
found that the fixed rules do not fall on what people edit, and the word-level
comparison of [ADR 0054](0054-trim-measured-against-editors.md) repeats that
for the default rules as a group.

The same comparison can be read for one phrase at one position: did the editor
remove it? On the development part of the expert-edited answers,
27 of 150 answers open with `Sure,` or `Sure!`, most often `Sure, here is` and
`Sure! Here's`. The experts removed or replaced that opening word in 27 of
27. `scaffold.chat-preamble` listed three openers and reported one finding
on those 150 answers.

## Decision

1. **Listing on a human reference.** A phrase may enter a default rule when,
   on a test part that was not used to choose it, editors removed or replaced
   every flagged word in at least 85% of its findings, the lower end of the
   95% interval of that share is above 70%, and the phrase produces at most
   one finding on the documentation pages and fiction paragraphs of the
   record. The rule was registered on issue 525 before the outcome on the
   test part was read.

2. **The opener.** `scaffold.chat-preamble` lists `Sure,` and `Sure!` at the
   start of a document. On the test part the two phrases produce 39 findings
   in 300 answers, and the expert removed or replaced the flagged word in
   39 of 39 (91.0% to 100.0%). They produce no finding on 121 documentation
   pages and 282 fiction paragraphs.

3. **Under a title too.** The answers of the reference have no headings, and
   the rule looked only at the very first sentence of a document. A README or
   a documentation page opens with a title, so a pasted answer under it went
   unseen. The rule now also looks at the first sentence after the leading
   headings, through a new phrase position, `prose-start`. `document-start`
   keeps its meaning. The findings on the reference are the same. On 3,302
   pages of public repositories and 455 generated pages of the
   assistant-panel corpus the rule reports nothing.

   The rule does not look after every heading. Real pages have FAQ sections
   in which a person answers a question heading with `Sure,` or `Sure!`, and
   that is ordinary writing. The gate, the severity, and the place outside the
   business profile are unchanged.

4. **Not listed.** A wide list of 160 words and phrases that models favor
   was part of the run with every rule. On the test part, which had been read,
   the words under its findings were words the expert changed in 75.8% of
   cases. The list was then tested under a registered rule on the
   1,733 expert-edited answers that no result had used: pruned on 850
   of them, where 10 phrases left, and confirmed on the other 883. There
   the words under its findings are words the expert changed in 69.2% of
   cases, against 56.3% for random words of the same answers
   (+12.9 points, +8.7 to +17.4). The expert removed or replaced the whole
   flagged phrase in 227 of 343 (66.2%) findings. The rule asked for
   70%, so the list does not ship, not even off by default. It beats chance
   and reports one phrase in three that an expert kept.

## Consequences

One part of `unswell check` now rests on what people did: the opener is a
phrase that experts removed every time they met it.

The reach is narrow. Chat models open an answer this way; a model that writes
a documentation page into a file does not. None of the 455 generated pages of
the assistant-panel corpus opens with it. The phrases catch chat output that
was pasted into a document.

The default rules as a group stay where ADR 0052 left them: on documentation
that maintainers reworded they flag edited and unedited sentences at the same
rate.
