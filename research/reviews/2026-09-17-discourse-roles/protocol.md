# Discourse roles and scoped editorial judgments

Base: PR #323, commit 6b1fe0010b56b3b00e9b8dff10c0e9948a99bc77.
Its six-page result was 2/57 before and 6/57 after, with all four gains on one
partially exposed historical page. All 130 selected source identities are now
exposed. The goal remains 80% frozen-event recall and 85% soft diagnostic
precision; these targets are not measured product properties.

## Hypotheses fixed before fresh prose review

The accumulated ledger has 587 events without full coverage: 195 wordiness,
175 empty framing, 108 vague claims, 49 repetition, 48 intensifiers, ten
complexity and two transitions. Expand bounded discourse roles in the existing
rules, sharing their source-preserving clauses and POS tokens:

1. An information subject can have an evaluative selecting predicate: a row,
   column, detail, distinction, or ordinal reference is the part worth reading,
   the one that matters, or important to know. Recognize grammatical roles and
   local boundaries, including shared copular predicates. A nominal list item
   such as "the row counts" must not be read as a subject and importance verb.
2. Author-to-reader attention notices and document-outcome endorsements can
   frame technical information without adding it: the author wants to stress
   a fact, hopes the guide helped, or evaluates a reader's understanding. Preserve
   actual roadmap commitments and epistemic uncertainty about a system's state.
3. Unscoped qualitative judgments can use ordinary copular predicates rather
   than only extreme adjectives. Require a described subject/action and a quality
   or difficulty judgment, with bounded scope checks. A stated measurement,
   condition, comparison mechanism or operational requirement is a control.
   The diagnostic requests the missing scope; it does not assert the claim false.
4. Abstract functional clefts can endorse their preceding explanation: an
   anaphoric result is what an abstract discourse object is for, or a described
   action is what that surface is for. Concrete component functions, literal
   identity assignments, and conditions remain controls. Surface matching is
   not semantic entailment or dependency parsing.
5. Relative instruction support must distinguish an actor from a supplied
   operand. A checksum used to cache a result does not itself perform caching.
   Retain genuine named procedures and layered reader actions, and report losses.

Do not add a second engine, a blanket keyword ban, lower gates, arbitrary scores,
or online inference. Preserve all prior labels and all prior controls. These
families do not cover every semantic defect; all remaining misses stay visible.

## Source and label freeze

Select six complete pages with seed `discourse-roles-confirmation-v1`: one short,
medium and long per cohort, distinct historical repositories, excluding all 130
references and hashes and the original historical frame. Freeze protocol,
selector, source bytes, identities and notices before reading prose. Read every
extracted block with original context and annotate all seven rubric categories,
uncertainties and controls before runtime changes or diagnostics. Audit exact
whole-block overlap against all prior sources before interpreting confirmation;
retain overlapping pages with an explicit exposure limitation, never replace
pages after reading them. Absence of exact overlap is not proof of independence.

The implementing Codex assistant is the reviewer accepted under ADR 0041; it is
neither a human nor an independent reviewer. Six-page counts are not population
recall. Freeze semantic code and focused public tests before opening confirmation
output. No semantic tuning after those outputs in this iteration.

## Reporting and validation

Replay all 16 sets in technical and strict, preserving configurations, source
mapping, resource evidence and the known #312 abstention. Review every changed
finding and every fresh warning. Credit only source-bound diagnoses of the
frozen defect; incidental length and punctuation overlap do not count. Report
full, partial and additional post-diagnostic repairs separately, along with
losses, false positives, controls, misses and review burden.

Run public API and compiled CLI tests, ordinary module tests, strict lint,
policy/schema/catalog checks, resource checks and CLI/MCP dogfood before
publishing. Race, active fuzzing and coverage remain deferred to #123. This
experiment does not authorize merge, release or playground deployment.
