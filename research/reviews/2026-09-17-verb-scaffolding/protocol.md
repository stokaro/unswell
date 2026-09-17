# Verb scaffolding and indirect action descriptions

Continue #299 from PR #315, commit
`ffa89b573d5b4d75c80c8639a3be5dd876b70d43`. All 88 previously reviewed
source identities are exposed. Their measured recall remains below the goal.

## Hypotheses fixed before selection

Test grammatical scaffolding that separates a technical subject from its action:

- An enabling verb followed by a nominal action or gerund and a passive
  infinitive: a clause allows specifying labels to be attached, for example.
  Require the nested action structure. A component allowing an actor to perform
  an action, or an authorization system granting permission, is a control.
- An action nominalization followed by a light passive method predicate and a
  gerund complement: configuration can be performed by using a setting.
  Preserve modality, actor and method in the advice. A passive identifying who
  performed an action, a legal obligation, or a named process is a control.
- Multiple capability or use predicates around one action: the ability to use
  something to perform an action, or a relative clause saying a setting can be
  used to perform its action. Ordinary can, used to, allows and ability are not
  violations by themselves. A diagnostic needs a removable layer of syntax.
- Nominal role definitions and relative-clause chains may hide a direct action.
  Only implement an explainable bounded construction that preserves its meaning;
  a density count, passive-voice ban or inferred semantic redundancy is insufficient.

Use the existing source-preserving NLP and rule engine. Bound candidate length
and work; respect protected tokens, quotations, negation, conditions, permissions,
comparison, risk and actor distinctions. Preserve every operational qualifier.
Do not lower thresholds, raise weights, invent authorship evidence, or claim a
whole paragraph is removable. Version changed rules. Unsupported hypotheses stay
missed, not retroactively removed from the denominator.

## Source-only confirmation

Select one page per Ptah/historical and short/medium/long cell from the pinned
metadata frame used in the prior review. Exclude all 88 exposed references and
hashes and the original historical study selection. Rank by SHA-256 of
`verb-scaffolding-confirmation-v1\n` plus reference. Historical allocation is
long, medium, short with one page per repository; Ptah is short, medium, long.
Record shortages rather than replacing a page after viewing its prose.

Freeze the protocol, selector, source bytes, manifest and notices before reading
selected prose. Review each complete page across all seven rubric categories,
including exact targets, context, proposed edits, uncertain events and controls.
Freeze annotations before runtime edits or diagnostic inspection. The implementing
Codex assistant is the maintainer-accepted reviewer under ADR 0041; this is not
human annotation or independent qualification. One page per cell supports counts,
not population estimates.

## Evaluation and delivery

Replay the exposed and new complete pages under unchanged technical and strict
policies before and after. Review every changed and confirmation finding. Preserve
full and partial credits, false positives, misses, applicability, gates, source
spans and resource costs. Keep the existing #312 abstention explicit. No semantic
tuning after opening confirmation output. Retain the working 80% recall / 85%
soft-precision goals as unfulfilled unless the declared evidence supports them.

Public-API and CLI tests cover positive and close negative cases, mapping,
configuration and limits. Complete required checks and CLI/MCP dogfood before
publication. Race, active fuzzing and coverage remain deferred to #123.
