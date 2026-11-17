# Habits of a page

Two rules of `unswell check` look at a page as a whole. Each names a habit
that shows only when a page keeps to it from the first paragraph to the last,
and each reports the page once. They're warnings, with no gate and no weight
in the score.

A model that writes under a style guide leaves few of the marks that the
other rules count, and its page can still read as a model's. The
[second reading of the AI-sounding index](ai-sounding-index.md) counts one
reason for that, runs of short plain sentences. These two rules name two
more.

## Paragraphs that end on a denial

`syntax.closing-denial-density` counts the paragraphs of two sentences or
more whose last sentence says what something is not, or what it does not do:

```text
The check compares two schemas. The comparison runs on a copy; it does not call the service.

We ran the plan on a copy. This is a rehearsal, not a migration.

The report lists every statement. It does not approve the plan.
```

One such sentence is an honest limit, and a reader needs it. A page that ends
paragraph after paragraph that way uses the denial as a device (a reader
hears it by the third one).

A last sentence denies in one of three forms:

- Its last clause, after a semicolon or a dash, opens with a negation:
  "; it does not call the service."
- It ends on ", not X": "This is a rehearsal, not a migration."
- Its main clause has a negation among the first six words: "It does not
  approve the plan."

A negation is "not", "cannot", "never", "neither", "nor", or a word that ends
in "n't". Some sentences with one are left alone:

- a question, and a sentence that ends with a colon;
- a clause that opens with "if", "when", "unless", "because", or another word
  of a condition, a time, or a reason;
- an instruction, such as "Do not edit the file by hand.";
- a sentence that speaks to the reader in its first words, such as "You
  cannot resume it.";
- "not only", "not just", "not all", and the like, which limit a statement;
- a sentence with "yet" or "anymore", which says where things stand now.

The rule reports the page when more than `allowed_occurrences` paragraphs end
on a denial (2 by default) and they are at least `onset` percent of the
paragraphs with two sentences or more (8 by default). The finding lists each
closing sentence.

**What to do.** Keep the denials that a reader would miss. End the other
paragraphs on what the thing is or does, or move the limit to where the
reader needs it: beside the claim that it limits, or into one section of
limits.

## One formal register

`syntax.uniform-register` counts the sentences of the running paragraphs and
those of them with a casual mark:

- a question or an exclamation;
- a first word "And", "But", "So", "Or", or "Yet";
- a contraction: "n't", "'re", "'ve", "'ll", "'d", "'m", and "'s" after a
  pronoun, as in "it's" (a possessive isn't one);
- parentheses that hold two words or more (one word in parentheses is an
  abbreviation or a unit).

The rule reports the page, at its first sentence, when it has at least
`min_sentences` sentences (25 by default) and at most `allowed_occurrences`
of them have a mark (0 by default).

People who write articles and posts use these marks all the time. In the
measurement below no blog post of a person went without one, and 2% of the
articles did. Nine in ten of the texts that two models wrote under a style
guide have none.

**What to do.** Say one or two things the way you'd say them to a colleague.
Don't sprinkle contractions over the page to quiet the rule: a reader hears
that too.

Reference pages and specifications are often written in one register on
purpose. Turn the rule off for them:

```yaml
version: 1
overrides:
  - files: ["docs/reference/**"]
    rules:
      syntax.uniform-register: {enabled: false}
```

## Measured behavior

The forms and the limits of both rules were chosen on the fitting halves of
the sets below. The bars were written on issue 603 before the rules ran on
the measured parts, and all 10 hold.

| Pages of the measured parts | Pages | Paragraphs end on a denial | One formal register | Either |
| --- | ---: | ---: | ---: | ---: |
| Posts of a product blog, the style-guide set of the earlier records | 30 | 13 (43.3%) | 18 (60.0%) | 23 (76.7%) |
| Blog posts that people wrote before November 2022 | 77 | 1 (1.3%) | 0 (0.0%) | 1 (1.3%) |
| Articles that people wrote under a style guide | 205 | 1 (0.5%) | 4 (2.0%) | 5 (2.4%) |
| Documentation pages that people wrote before November 2022 | 61 | 2 (3.3%) | 2 (3.3%) | 4 (6.6%) |
| Texts that two models wrote under a style guide | 120 | 1 (0.8%) | 110 (91.7%) | 110 (91.7%) |

Each rule stays within the limit of
[ADR 0058](adr/0058-default-rules-quiet-on-human-pages.md) for a rule that
ships on, one in twenty pages that people wrote. On all 121 documentation
pages of people the first rule reports 2 and the second 3.

Two sets have no parts and were read whole. Of 3,302 pages of public
repositories today, the first rule reports 8.8% and the second 6.1%. Of 455
pages that three models wrote or polished for the panel corpus, 4.4% and
34.9%.

The [record](../research/reviews/2026-11-17-page-habits/README.md) holds
these counts, and `make check-page-habits` recounts them.

## Limits

- The denials are a habit of that product blog and not of every model. The
  two writer models of the measurement almost never end a paragraph that way:
  1 of their 120 texts is reported.
- Neither rule says who wrote a page. A person who writes formally gets the
  second finding. A model that was told to use contractions doesn't.
- People who clean AI-sounding wording out of documentation don't take these
  habits out. On the test part of the reference of
  [ADR 0056](adr/0056-rules-tested-on-de-slop-commits.md), files before and
  after a commit that says it removes such wording, the first rule reports
  54 of 985 files before the commit and 55 after it, and the second 41
  and 42.
- Whether a page that is revised until both rules are quiet reads as a
  person's has not been measured. A revision by the report of the index did
  not get there; see
  [a revised post against human text](ai-sounding-index.md#a-revised-post-against-human-text).
- The forms are read from the words of a sentence and not from a parse, and
  the words are English. A denial whose negation stands after the sixth word
  of its clause is not counted.
- Reference pages and specifications are outside the measurement.
