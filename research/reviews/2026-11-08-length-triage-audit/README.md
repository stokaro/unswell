# Length-only triage does not preserve useful findings

This [post hoc audit](PROTOCOL.md) joins the frozen 176-finding assistant review
to the later complete technical replay by exact finding ID and byte span. It
uses all 123 unchanged reference events and checks four input SHA-256 values.
The product runtime, policy, review labels, and source pages are unchanged.
The current frozen replay has 28/57 FULL whole-page events and 82/126 accepted
technical findings; both #349 gates fail.

Of 84 whole-page `syntax.long-sentence` findings, 46 are actionable, 37 are
uncertain, and one is nonactionable. Prose-word count has exposed pairwise
AUROC **0.707** for actionable versus nonaccepted findings. The same measure
is 1.000 on only nine contextual findings and 0.500 on only four historical
confirmation findings. These small groups cannot establish transfer.

| Simulated onset | Nonaccepted removed | Actionable removed | Optimistic accepted fraction after exactly 18 perfect new findings |
| ---: | ---: | ---: | ---: |
| 35 (current) | 0 | 0 | 69.4% |
| 37 | 17 | 7 | 77.5% |
| 41 | 27 | 21 | 82.3% |
| 42 | 31 | 24 | 85.4% |

The simulated onset is a **post-result** arithmetic parameter, never a
candidate policy. At onset 42, even the optimistic future-recall scenario
would reach 85% only by hiding **24 accepted** existing findings. With at most
five accepted findings removed, a length-only onset removes at most four
nonaccepted ones. An onset of 46 would also lose an existing FULL-event
witness. The [full threshold table](result.json) preserves every integer onset
and every cohort count.

This retires simple onset tuning as the next #349 step. The frozen rationales
show the actual distinction: many accepted sentences pack independent cases
or opposite conditions into one statement, while many uncertain sentences
form a coherent causal explanation. A useful triage mechanism must assess
that relationship in source context and be tested against close legitimate
controls. Recall still needs separate operators for empty framing, vague
claims, wordiness, and repetition. No origin or authorship conclusion follows.

The source-hashed quality-claim report was recorded before the recent
research-only merges. A tree comparison from its merged main commit
`0b5433f` to `d6a40a8` found no changes in the relevant runtime Go packages;
this audit nevertheless asserts only the pinned report, not an unrun current
36-page replay. No untouched prospective page was opened.

Run `make check-length-triage-audit` for the bounded verifier and mutation
checks. Full `make check` covers this target before publication.
