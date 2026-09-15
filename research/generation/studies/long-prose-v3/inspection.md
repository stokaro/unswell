# Construction review

This is agent inspection of source-bound cases, not a human annotation round.
`analysis/cases.json` retains the first two findings by source path and byte offset
for each rule and category: historical originals, each generate/neutral family,
and other controlled arms. Each case includes the source hash, exact diagnostic,
related spans, and the containing paragraph range. The category is recorded, so
an editing example cannot silently become evidence about generation from facts.

The table records concrete cases inspected in all four data layers.
The complete per-rule counts and statistical dispositions are in the results
report. A recognized construction is not automatically an editorial defect.

| Construction | Case | Inspection and limitation |
| --- | --- | --- |
| Announced importance | Qwen's ripgrep benchmark and release-timing paragraphs | The prefaces can be removed while retaining their following assertions. The fixtures preserve 3.423 and 13.031 seconds, unchanged ugrep behavior, both documentation filenames, volunteer delivery, and no release-date promise. |
| Connective density | Qwen's curl vulnerability-process paragraph | The paragraph adds `Furthermore` and `Additionally`. Its distinct restrictions on public trackers, mailing lists, and commit messages must survive any edit. Counting connectives does not make those restrictions redundant. |
| Weak intensifiers | Diesel's contribution welcome | `very much look forward` carries emphasis in a welcome. This is an optional tone choice, not evidence of an AI defect in a technical contract. |
| Wordy phrases | ripgrep's explanation of finite-state matching | `in order to` can be shortened to `to`. The limits on backreferences and the linear worst-case guarantee must remain. The source predates the experiment. |
| Em-dash density | Qwen's Libevent compatibility description | The inserted definition of correct code states which interfaces and behaviors qualify for compatibility. Splitting the sentence may improve structure, but deleting the definition weakens the contract. |
| Grade metric | ripgrep's explanation of nondeterministic result order | The paragraph repeats the consequences of thread interleaving. The grade formula identifies complexity; it cannot distinguish necessary explanation from repetition or establish authorship. |
| Long paragraphs and sentences | ripgrep's single-line matching explanation | The text explains why scanning whole buffers is faster and how matching must still respect line boundaries. Length is measurable. A rewrite must retain both the performance rationale and the matching restriction. |
| Exact sentence repetition | curl's Visual C++ build instructions | Read the repeated sentence with its surrounding build steps. Different sections may need the same prerequisite; exact equality alone does not establish that one occurrence can be removed. |
| Near-sentence repetition | Diesel's `setup` and `database setup` descriptions | The commands share database creation and migration-table behavior, while their surrounding setup responsibilities differ. Similarity does not establish interchangeable commands. |
| N-gram repetition | Libevent's event-loop wakeup explanation | Repeating `main event loop` keeps the subject clear while comparing socketpair, eventfd, and EVFILT_USER behavior. Replacing technical terms with synonyms merely to reduce a count would be inappropriate. |
| Paragraph openings | ripgrep's platform-specific installation instructions | Parallel `If you're a ... user` introductions precede different package-manager commands. The repetition supports lookup by platform. |
| Sentence openings | ripgrep's default-engine and PCRE2 explanations | Repeated entity names can connect related API explanations. Inspect the actual sentences and their sections rather than treating an opener count as a defect label. |
| Paragraph overlap | Pydantic's repeated sponsor acknowledgments | Similar attribution appears in separate release entries. Its repetition is documentary context, not proof that either acknowledgment is unwanted. |
| Noun stacks | ripgrep's default regex engine; Qwen's version-control wording | The historical phrase is established terminology. The generated phrase `version control ignore rules` is also technical. A noun/POS sequence alone does not justify replacing the terms. |
| Paired contrasts | Moment's ordinal return value and `Date` copying changes | The two alternatives report distinct behavior changes. The rule recognizes repeated contrast wording correctly; removing either distinction would discard information. |
| Parenthetical load | ripgrep's `--dfa-size-limit 1G` caveat | The parenthesis distinguishes a permitted cache maximum from immediate memory allocation. That distinction must survive a shorter rendering. |
| Passive candidates | ripgrep's shell-completion distribution instructions | Availability, separate maintenance, and inclusion in releases naturally focus on the files. Repeated passive forms alone do not establish avoidable vagueness. |
| Paired contrasts in Luna | ripgrep's Unicode-performance explanation | Two adjacent contrasts distinguish an observed experiment from a universal claim about PCRE2. Their form repeats, but the qualification is necessary. A rewrite must preserve that evidential limit. |
| Repeated openings in Luna | curl's private-fix and disclosure sequence | Repeated `The fix` openings track one fix across private preparation and public release. The sequence must retain the CVE instruction, distribution sharing, the 48-hour limit, and the requirement to include the fix in the release. |
| Release-note similarity in Luna | Moment's versioned change history | Similar `Version ... fixed ...` sentences describe different releases and different defects. Their shared template supports navigation; deleting one because of lexical similarity loses a change record. |
| Noun-stack tagging boundary in Luna | fzf's `default preview command exists` and `fix duplicate preview lines` | The marked spans include verbs in context. These are counterexamples to interpreting the surface POS sequence as a verified grammatical noun phrase. Keep the uncertain construction claim separate from editorial advice. |
| Nominalization in Haiku | curl's `conducts an investigation of the submission` | `investigates the submission` preserves the action. The executable edit leaves acknowledgment, rejection reasons, the policy condition, acceptance, and the notification of fix work unchanged. |
| Long paragraphs in Haiku | containerd's release-support matrix rendered as prose | The response turns version dates and lifecycle states into one paragraph. Restoring a table can aid lookup, but the dates and conditional end-of-life rules must remain. Length alone does not establish an AI-origin defect. |
| Noun-stack tagging boundary in Haiku | ripgrep's `language documentation translation exists` and `build output path resembles` | Both marked spans include a finite verb. These extend the tagging counterexamples observed in Luna and limit interpretation of a noun-stack count. |
| Paired contrasts in Haiku | bat's syntax-map and highlight-range changes | Extensions versus globs and a single line versus ranges are separate API changes. Their shared comparison form does not make either change expendable. |

## Executable rewrite cases

`e2e/testdata/long_prose_research` contains separate original and edited files for
the two Qwen prefaces and the Haiku nominalization, plus the Moment contrast counterexample. The custom fixture
policy uses nonblocking notes and zero weight for the preface rule. It isolates
recognition and source coordinates without changing product defaults.

The first rewrite removes only the preface and capitalizes the new sentence
opening. The second does the same inside a paragraph. Both retain every other
word, including the numerical results and the absence of a timing promise. The
expected detections disappear in the edited files. The technical contrast still
produces a note because its repeated form remains; the fixture expects a passing
gate.

The Haiku edit changes only `conducts an investigation of` to `investigates`.
Its note disappears while every other word of the security-report process stays
unchanged. The nominalization rule also has zero weight in this fixture.

`e2e/testdata/contextual_ptah` supplies three additional source-bound contrast
cases and alternatives. Its review preserves failure behavior, URI schemes,
directory overrides, missing-directory errors, and registry/service boundaries.
Removing a restriction is not a successful rewrite merely because the score falls.

These tests establish deterministic construction changes and retained named
obligations. Reader preference, editorial precision, and recall remain unmeasured.
No origin or quality label is inferred from the fact that a model wrote a passage.
