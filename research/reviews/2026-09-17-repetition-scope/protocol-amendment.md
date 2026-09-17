# Scope clarification from existing regression tests

Before opening confirmation diagnostics, the existing Go fixtures exposed an
unintended side effect of treating every named source declaration as a different
comparison owner: duplicate strings in different constants stopped matching.
This issue concerns document sections, task variants and table roles. Retain the
existing scope for source declarations without heading structure. In the protocol,
structural owner means a heading/container section; named variables and functions
do not independently split comparison groups.

This clarification came from existing blackbox fixtures, not confirmation outputs.
The six selected inputs, their labels and the before/after policy remain fixed.
The confirmation files are Markdown and MDX, so this clarification does not change
their comparison scope. No threshold changes or additional exclusions were made.
