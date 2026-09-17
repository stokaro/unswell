# Remaining confirmation defects

All 31 lack full coverage in both profiles; two have partial findings.

## c01-d02: wordiness

The claim/site/personification and present-catalog versus whatever-was-true layers obscure the actual ownership and freshness rule.

Proposed edit: Say the operator's compatibility matrix records the Ptah builds supported by each operator version and stays current outside Ptah's per-version archives; keep the link.

Bytes 2684–3095

```text
Which Ptah build an operator version runs is the operator's claim, not this
site's, and it changes when the operator measures something new. It is
published as a current table at
[the operator compatibility matrix](https://docs.ptah.run/compatibility/operator/),
outside this site's per-version archives so that it answers with the present
catalog rather than with whatever was true when a Ptah release shipped.
```

## c01-d03: empty_framing

The defensive explanation for avoiding duplicate documentation adds no destination or installation constraint.

Proposed edit: Keep the operator guide link, versioning and covered topics; remove the justification.

Bytes 3516–3641

```text
and this page deliberately does not restate any of them -- a second
copy of an install command is a second thing to get wrong
```

## c01-d04: empty_framing

The closing authenticity comparison repeats the already stated live recording and checked publication conditions.

Proposed edit: Keep the recorded operations, live checks and publication condition; remove the closing comparison.

Bytes 4069–4152

```text
so what a reader watches is a session that happened
rather than a screencast of one
```

## c02-d01: wordiness

The answer-to-what-happens wrapper delays the actual role of a consistency mode.

Proposed edit: Say a consistency mode determines how writes during backfill are handled; retain the concurrent-write context.

Bytes 596–662

```text
A consistency mode is
your answer to what happens to those writes.
```

## c02-d02: empty_framing

The guarantee/precision announcement adds no condition to the transaction explanation.

Proposed edit: Begin with the actual commit and rollback guarantee.

Bytes 1637–1700

```text
That is the whole guarantee, and it is worth stating precisely:
```

## c02-d03: wordiness

The repeated committed predicate restates the single-transaction relationship circularly.

Proposed edit: Say the change and its event commit together; keep the rollback case.

Bytes 1701–1790

```text
a change that
committed has an event, because a transaction that committed committed both
```

## c02-d04: empty_framing

The cost announcement adds no measured or operational detail beyond the following table and trigger counts.

Proposed edit: Begin with the two triggers and companion table and preserve their distinct lifetimes.

Bytes 2012–2029

```text
The cost is real:
```

## c02-d05: empty_framing

The reading advice does not define usable or add a retention condition.

Proposed edit: Begin with the shared table and slower-reader retention condition; keep the precise usable-feeder requirements.

Bytes 2365–2424

```text
"Every usable live feeder" is the part worth reading twice.
```

## c02-d06: unjustified_intensifiers

The emphasis adds no operational criterion to controlling writes; the examples already specify the intended circumstances.

Proposed edit: Use it when you control the writes; keep all examples.

Bytes 4113–4122

```text
genuinely
```

## c02-d07: wordiness

The missing-reporting-surface explanation repeats inability to report in three forms.

Proposed edit: Say the evidence assessment is implemented and tested, but writers have no reporting interface, so this build rejects the mode; keep the named alternatives.

Bytes 4359–4636

```text
What is missing is the reporting surface:
the assessment that would hold a writer's evidence to a policy is written and
tested, and nothing exists for a writer to report *through*, so a run selecting
the mode could only ever be told that its writer had never reported anything.
```

## c02-d08: empty_framing

The honesty self-evaluation adds no failure behavior beyond the explicit rejection.

Proposed edit: Keep the rejection and missing interface; remove the self-evaluation.

Bytes 4637–4685

```text
Refusing it is what this build can honestly say.
```

## c02-d09: wordiness

The design/mode-means/recorded-here/coming-back-versus-going-away chain lengthens one concrete future-support statement.

Proposed edit: Introduce the following section as the planned application-managed mode; preserve that it is intended to return.

Bytes 4687–4807

```text
The design below is what the mode means, and it is recorded here because the
mode is coming back rather than going away.
```

## c02-d10: vague_claims

The comparative prediction about reader comprehension has no evidence or technical criterion.

Proposed edit: Keep the weakest-guarantee description and the precise missing-report and freshness conditions.

Bytes 4962–5007

```text
and the one whose weakness is
easiest to miss
```

## c03-d01: empty_framing

The importance announcement adds no mechanism beyond the specific embedding consequences that follow.

Proposed edit: Begin with the listed embedding consequences.

Bytes 4368–4406

```text
That separation matters for embedders:
```

## c03-d02: empty_framing

The anthropomorphic closing repeats that all readers return the same schema representation.

Proposed edit: Keep the reader functions and common IR return type; remove the closing personification.

Bytes 9384–9429

```text
Nothing downstream
knows which one filled it.
```

## c04-d01: wordiness

The assurance wrapper adds no verification step to setting the environment variable.

Proposed edit: Set ETCDCTL_API=2 for the v2 API; keep the default and link.

Bytes 136–152

```text
make sure to set
```

## c04-d02: needless_repetition

The example preface repeats unsupported syntax already stated immediately before it.

Proposed edit: Use a short example label; retain the newline restriction and possible hang.

Bytes 1929–1974

```text
For example, following case is not supported:
```

## c04-d03: wordiness

Is used to specify wraps a direct command capability. The assignment clause is meaningful and must remain.

Proposed edit: Say ROLE specifies roles that can be assigned to etcd users; do not turn role assignment into a permission grant.

Bytes 30111–30189

```text
ROLE is used to specify different roles which can be assigned to etcd user(s).
```

## c05-d01: needless_repetition

The two predicates express the same skip behavior.

Proposed edit: Say the option skips the file; keep the generated-file example.

Bytes 251–300

```text
skip that file and excludes it from being checked
```

## c05-d02: wordiness

Check and verify plus script-is-an-effort lengthen the useful limit and purpose.

Proposed edit: Say checksrc checks common mistakes rather than the entire style guide; retain the contributor context.

Bytes 544–656

```text
does not check and verify the code against the entire style guide,
but the script is instead an effort to detect
```

## c05-d03: wordiness

Repeated lists and warnings-it-has/problems-it-detects wrap the usage and warning list.

Proposed edit: Say the help lists usage and supported warnings; retain the following warning catalog.

Bytes 887–981

```text
Lists how to use the script and it lists all existing warnings it has and
problems it detects.
```

## c05-d04: wordiness

The directory/activation/enabling wording repeats the same instruction before the next sentence supplies the per-line syntax.

Proposed edit: Add the selected warning names to .checksrc in the target directory, one per line; keep the default-cost explanation and syntax.

Bytes 4079–4210

```text
place a `.checksrc` file in
the directory where they should be activated with commands to enable the
warnings you are interested in
```

## c05-d05: wordiness

The vague nature/flaws/there-is-a-need wrapper delays suppression guidance.

Proposed edit: Say some code triggers false positives and checksrc supports suppressing those warnings; retain the following methods.

Bytes 4495–4622

```text
Due to the nature of the source code and the flaws of the checksrc tool, there
is sometimes a need to ignore specific warnings.
```

## c05-d06: wordiness

Within-a-source-file and in-the-source-code repeat the location while providing-instructions wraps the suppression action.

Proposed edit: Put checksrc suppression instructions in the source file; preserve the marker, count and region choices.

Bytes 4692–4817

```text
You can control what to ignore within a specific source file by providing
instructions to checksrc in the source code itself.
```

## c05-d07: wordiness

The nominal enabling/passive performed wrapper obscures the re-enable action.

Proposed edit: If the warning is not re-enabled before end of file, it is re-enabled automatically for the next file.

Bytes 5281–5339

```text
If the enabling isn't performed before the end of the file
```

## c05-d08: unjustified_intensifiers

The presumption about reader knowledge adds no range information.

Proposed edit: State that the count can be changed to any integer.

Bytes 5708–5717

```text
of course
```

## c05-d09: needless_repetition

The intended-only/nothing-extra purpose repeats the explicit counted-suppression behavior already described.

Proposed edit: Keep the counted example and configurable integer; remove this repeated assurance.

Bytes 5758–5850

```text
It can be used to make sure only the exact intended
instances are ignored and nothing extra.
```

## c06-d02: wordiness

Intended-to-make and easily-usable add layers around the purpose of the pipeline helper functions.

Proposed edit: Say these helpers let query results be passed through pipelines; keep the function names.

Bytes 1688–1779

```text
`first`, `label` and `value` are intended to make query results easily usable in pipelines.
```

## c06-d03: wordiness

Intended-to-produce/consumption-by-humans lengthens the display-format purpose.

Proposed edit: Say humanizing functions format values for display; keep that output may change across versions.

Bytes 2478–2566

```text
Humanizing functions are intended to produce reasonable output for consumption
by humans
```

## c06-d04: wordiness

Intended-to-allow/to-be-passed nests purpose and capability around the concrete argument mechanism.

Proposed edit: Say the generated argument map supports passing multiple arguments to templates; retain the key mapping.

Bytes 4114–4185

```text
This is intended to allow multiple arguments to be passed to templates.
```

## c06-d05: wordiness

Each-of-the-types, information-that-can-be-used-to, and unspecified-other-differences create a vague preview instead of naming the type-specific context.

Proposed edit: Say template types provide different parameter data, then introduce the listed alert and console contexts; preserve that the following sections also describe other differences.

Bytes 4548–4686

```text
Each of the types of templates provide different information that can be used to
parameterize templates, and have a few other differences.
```
