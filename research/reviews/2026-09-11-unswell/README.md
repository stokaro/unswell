# Review on real text: this repository, September 11, 2026

This directory records the third review of the kind that
[ADR 0037](../../../docs/adr/0037-diagnostics-not-authorship.md) asks
for. Every catalog rule ran on this repository at commit `a1c6d7a`, over
the tracked files that its own policy selects. A reader judged a sample
of the findings under each rule's own statement. Nothing here is an
authorship claim, and the review is not a quality verdict on the
repository's documentation.

One reader judged the sample: the maintainer's agent session, on the day
of the run. The maintainer has not confirmed the judgments.

## What ran

The tool was the merged commit `a1c6d7a`, the same commit as the
subject. [config-all.yaml](config-all.yaml) is the all-rules
configuration of the [Ptah review](../2026-09-11-ptah/README.md) with
the file selection and the extraction exceptions of `.unswell.yaml`. The
selection keeps Markdown, Go, shell, and YAML files and leaves out test
data, licenses, and build artifacts. The exceptions keep the deliberate
violations of the rule catalogs and the test fixtures out of the count.
The run used `--no-gate --include-source --timeout 20m --jobs 4` and
completed: 785 documents, 149,157 prose words, 488 findings.
[summary.json](summary.json) holds the counts by rule and by file kind.

| File kind | Documents | Prose words |
| --- | --- | --- |
| Markdown | 113 | 97,772 |
| Go | 329 | 38,274 |
| Go tests | 298 | 5,885 |
| Shell and YAML | 45 | 7,226 |

## Findings by rule

| Rule | Total | Markdown | Go | Go tests | Shell and YAML |
| --- | --- | --- | --- | --- | --- |
| `readability.grade-metric` | 365 | 348 | 11 | 1 | 5 |
| `syntax.passive-candidate-density` | 73 | 67 | 5 | 1 | 0 |
| `syntax.long-sentence` | 19 | 10 | 1 | 1 | 7 |
| `syntax.noun-stack` | 16 | 9 | 3 | 2 | 2 |
| `syntax.parenthetical-load` | 10 | 3 | 0 | 1 | 6 |
| `repetition.sentence-openers` | 4 | 4 | 0 | 0 | 0 |
| `readability.long-paragraph` | 1 | 1 | 0 | 0 | 0 |

The other 33 rules fired nowhere. Sixteen of them sit behind the gate of
the repository's own policy, which holds them at zero on every merge.
Readability grade and passive candidates carry 438 of the 488 findings,
and 415 of those sit in the Markdown documentation.

## The judged sample

[sample.json](sample.json) holds a seeded sample of ten findings per
rule, or every finding where a rule has fewer, with the judgment and its
reason for each. The table condenses it.

| Rule | Justified | Not | Why not |
| --- | --- | --- | --- |
| `readability.grade-metric` | 0 | 10 | technical terms, hyphenated compounds, and identifiers raise the formula in short sentences; one jq program in a shell string |
| `syntax.passive-candidate-density` | 9 | 1 | `are fixed stability runs` is a copula with an adjective |
| `syntax.long-sentence` | 7 | 3 | Python and jq programs inside shell strings |
| `syntax.noun-stack` | 4 | 6 | the tagger reads `applies`, `reconstruct`, `remain`, and `check` as nouns; Python import lines in a shell string |
| `syntax.parenthetical-load` | 1 | 9 | code inside shell strings and a workflow file; issue references such as `(#56)`; license labels after dependency names |
| `repetition.sentence-openers` | 3 | 1 | each entry of the API log opens with its symbol |
| `readability.long-paragraph` | 1 | 0 | |

Twenty-five of the 55 sampled findings name a construction in prose a
reader could act on. The passives are real and mostly agentless, and a
named actor would carry each of them. The long sentences are
enumerations of 38 to 45 words. The four justified noun stacks sit in
error messages and comments. Thirty do not. The readability grade
accounts for ten of them: on this documentation the formula measures the
vocabulary, not the sentences. Code in shell strings accounts for
eleven, across four rules. The part-of-speech tagger accounts for four
noun stacks, where a verb read as a noun completes the sequence.

## What it shows

The repository's gated policy already holds the forbid rules at zero, so
the review reads what the remaining rules add. The passive candidates
add findings a reader could act on. The readability grade adds a number
that tracks the vocabulary of the subject, and the rule's own statement
says to keep necessary technical terms. The string context adds the same
noise here as in Ptah: programs inside shell strings read as prose.

Two limits of the tool appear that the earlier reviews did not name.
Issue references and license labels in parentheses count as
parenthetical load. And the tagger's noun reading of a verb after a noun
sequence completes a stack that the sentence does not hold. Both are
questions for the rules, recorded here and not acted on.

## What it is not

The judgments are one reader's, and the reader is an agent session. The
counts are not a precision figure, not a claim about the documentation's
quality, and not a change to the repository's policy. The limits are
observations, not defects filed.
