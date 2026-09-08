# Experimental repetition signals

Unswell uses these signals to find repeated wording and form that may need editing.
They do not identify an author or establish that two passages have the same meaning.
All five rules are disabled in builtin profiles until explicitly enabled. Their
thresholds are experimental policy proposals; corpus precision has not been measured.

```yaml
version: 1
extends: [builtin:technical]
rules:
  repetition.ngram-density:
    enabled: true
    parameters:
      min_ngram_words: 3
      max_ngram_words: 8
      window_sentences: 8
      allowed_occurrences: 2
      saturation_occurrences: 5
  repetition.paragraph-overlap:
    enabled: true
    parameters:
      min_words: 12
      window_blocks: 8
      similarity: 0.85
  repetition.summary-echo:
    enabled: true
    parameters:
      phrases: [summary, conclusion, recap]
vocabulary:
  terms: [optimistic concurrency control]
  term_exemptions: [repetition.ngram-density, repetition.paragraph-overlap]
```

The existing [configuration](configuration.md), file overrides, terminology,
[suppressions](suppressions.md), and [baseline](baseline.md) contracts apply.
Changing severity does not change scoring. Choose `gate: forbid` explicitly when
the team wants any active finding from a rule to fail its policy.

## Measured behavior

| Rule | Default candidate and window | Activation and evidence |
| --- | --- | --- |
| `repetition.ngram-density` | 3–8 consecutive words; at least 12 words per sentence; 8 prose sentences | Allow 2 nonoverlapping occurrences, saturate at 5; report counts, n-gram length, prose denominator, and coverage |
| `repetition.syntax-template` | Complete normalized POS shape; at least 12 words; 8 prose sentences | Allow 2 occurrences, saturate at 5; require at least 2 different lexical realizations |
| `repetition.paragraph-overlap` | At least 12 nonexempt words and 2 distinct content words; pairs within 8 prose blocks | Set Jaccard at least 0.85; allow 1 block, saturate at 4; report connected clusters and qualifying pair metrics |
| `repetition.heading-echo` | A selected heading and its immediately following paragraph; at least 6 nonexempt words in each | Set Jaccard at least 0.85; advisory match with full activation |
| `repetition.summary-echo` | Paragraphs under configured summary headings and earlier nonsummary paragraphs; 8 prose blocks | Same overlap formula and count activation as paragraph overlap; summary locations appear first |

Count activation is `clamp((count - allowed) / (saturation - allowed), 0, 1)`,
stored in integer thousandths. The n-gram coverage denominator includes prose words
from the first through the last participating sentence, including intervening prose.
It excludes protected sentences and does not grow when unrelated clean paragraphs
are appended. Coverage is an explanatory measurement; it is not a probability or
the activation formula.

Set Jaccard is the number of shared distinct words divided by the number of distinct
words in either block. Approved terms do not supply overlap evidence. Candidate
postings use content words, excluding a fixed function-word list; exact similarity
uses all nonexempt prose words. Each accepted pair meets the configured threshold.
Connected clusters may span more than one window and do not claim that every possible
pair meets it. Metrics report the number of accepted pairs, their minimum Jaccard,
and the sums of their distinct-word intersections and unions.

N-grams never cross punctuation, sentence boundaries, or protected code. They require
two distinct content words and use a protected signature for their surrounding
sentence. Longer phrases suppress a shorter cluster when all its occurrences are
already covered. Partially overlapping phrases can remain separate findings; the
existing repetition-group cap limits their combined contribution.

POS templates collapse noun, verb, adjective, and adverb tag variants into families.
They retain token order, punctuation, approved term wording, and protected signatures.
Exact repeated sentences alone do not activate the template rule. Imperative openings,
questions, and list items are excluded. The POS model can assign different tags to
similar constructions; this exact tag-shape method does not recognize every repeated
grammatical structure. It makes no dependency-parser claim.

## Technical and structural boundaries

By default, negation, numbers, versions, identifier candidates, numeric units, modal and condition
cues, and approved term wording form a protected signature. Different signatures
cannot form a template or overlap cluster. Thus changing a permission to an obligation,
a timeout, a unit of measurement, or an identifier does not become duplicated
information merely because the other words overlap. These checks are lexical
heuristics; equal signatures do not prove semantic equivalence.
The `protect_negation`, `protect_numbers`, and `protect_identifiers` parameters
control their respective checks. Modal and condition cues are always retained.

N-grams and templates operate on selected paragraph, comment, and string prose.
Phrase occurrences are separate source ranges; no text is joined across comments or
strings. Paragraph overlap compares separate blocks of the same kind. Lists and
tables are excluded from these experimental signals. A paragraph containing protected
code is ineligible for overlap, rather than being compared after removing its code.

Summary titles are whole normalized heading matches, not substrings or words found
in paragraph prose. Defaults recognize `summary`, `conclusion`, and `conclusions`.
Nested headings remain in the section until grammar-derived ancestry leaves it.
An actual selected heading is required; an excluded heading's structural context
cannot activate the rule. Heading echoes stop at intervening blocks or excluded code.

## Limits and qualification

Window sizes are bounded: 1–1000 sentences and 1–128 prose blocks. Short prose blocks
still consume a position in the block window. An expiring inverted index bounds
active paragraph postings, and feature extraction, index work, and comparisons
consume each rule's `max_candidates` budget. Exhaustion returns an operational error. There is no
sampling, silent truncation, or successful incomplete gate.

The existing engine owns evidence validation, source spans, local scores, correlation
caps, permissions, and report generation. MCP returns that same completed result.
Saved reports retain measurements without the source text by default. The new typed
parameters and structural requirement are documented in the [public API](public_api.md).

Root `e2e/testdata/repetition_patterns` and `repetition_code` fixtures contain readable
expected detections and exact CLI/JSON goldens, checked against SARIF and the original
source. Library tests cover structural selection, technical differences, terminology,
count thresholds, cancellation, limits, concurrent reuse, and nondilution. Catalog
examples use their declared input format, including grammar-parsed Markdown.

`BenchmarkRepetitionScaling` measures the full public-engine scan for 128, 512, and
2048 synthetic technical blocks with different numeric facts. Run it with:

```sh
go test . -run '^$' -bench BenchmarkRepetitionScaling -benchtime=3x -count=1
```

Its declared candidate budget is 8,000,000. Reported allocation totals are not peak
resident memory. This scaling check does not establish corpus accuracy or the
roadmap's Linux 2-vCPU/512-MiB performance target. Qualification and full resource
acceptance remain in #26 and #28. [ADR 0010](adr/0010-repetition-signals.md) records
the design and compatibility boundaries.
