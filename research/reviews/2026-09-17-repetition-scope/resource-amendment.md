# Context serialization budget

Code review after the initial scans found that repeated heading labels could
amplify context serialization work: a long heading can be copied for every short
block below it. This is an implementation resource issue, not a confirmation
finding used to tune a matcher.

Scope construction now charges each block and each structural label's bytes
against `analysis.max_candidates` before serialization. Exhaustion returns the
existing typed rule abstention and discards partial findings. The matching
predicates, thresholds, frozen sources and annotations are unchanged.

Blackbox tests exercise all five affected rules with a long heading and a small
budget. The published after reports are regenerated from the bounded code;
compare every diagnostic and abstention against the initial implementation at
`2ded5ff7e3a15b68d722a6d2b8e1599eca9351e7`. Any difference must be reported rather
than tuning the budget from confirmation outcomes.
