# Confirmation misses

These ten judgments were frozen before diagnostic output review. They did not
become tuning inputs in this iteration. A later iteration may use them as
development material only with a new confirmation selection.

## c01-d01 — vague_claims

[docs/site/src/content/docs/inference/guides/migrate-a-paused-source.md](https://github.com/stokaro/ptah/blob/654eae5591392278e6c8bce8e54737f780766f19/docs/site/src/content/docs/inference/guides/migrate-a-paused-source.md)

```text
The second is what most people want after discovering the source was not as
paused as they thought.
```

Claims what most operators want without a stated observation; the operational choice is whether later writes need replay.

Proposed edit: Use the outbox when the source may receive writes that must be replayed.

## c02-d01 — empty_framing

[docs/site/src/content/docs/inference/guides/rollback-and-retire.md](https://github.com/stokaro/ptah/blob/654eae5591392278e6c8bce8e54737f780766f19/docs/site/src/content/docs/inference/guides/rollback-and-retire.md)

```text
This is the part people miss.
```

The paragraph announces reader error without explaining the procedure; the following paragraph already names the rollback precondition.

Proposed edit: Delete this paragraph.

## c03-d01 — empty_framing

[docs/site/src/content/docs/databases/sqlite.md](https://github.com/stokaro/ptah/blob/654eae5591392278e6c8bce8e54737f780766f19/docs/site/src/content/docs/databases/sqlite.md)

```text
matching the community binary is what the surface
is for
```

The compatibility rationale repeats the named compatibility contract rather than specifying how loss is handled.

Proposed edit: Remove the parenthesis and retain that the document is unchanged and the loss is reported.

## c05-d01 — empty_framing

[docs/site/src/content/docs/versioned/generate.mdx](https://github.com/stokaro/ptah/blob/654eae5591392278e6c8bce8e54737f780766f19/docs/site/src/content/docs/versioned/generate.mdx)

```text
which is the question a reviewer opens a
safety report to answer
```

Claims the report answers the reader's purpose after already specifying its severity counts; removing it preserves the report behavior.

Proposed edit: Delete this relative clause; retain the counts and offline behavior.

## c05-d02 — empty_framing

[docs/site/src/content/docs/versioned/generate.mdx](https://github.com/stokaro/ptah/blob/654eae5591392278e6c8bce8e54737f780766f19/docs/site/src/content/docs/versioned/generate.mdx)

```text
This is a first-class
origin rather than a fallback:
```

Assigns status to the documented workflow before the concrete cases that already justify using it.

Proposed edit: Start the next sentence with A project that describes no desired schema; preserve all use cases.

## c06-d01 — empty_framing

[docs/site/src/content/docs/atlas/schema-commands.mdx](https://github.com/stokaro/ptah/blob/654eae5591392278e6c8bce8e54737f780766f19/docs/site/src/content/docs/atlas/schema-commands.mdx)

```text
and refusing would copy a limitation this
implementation does not have
```

Defends a rejected design after SQLite support and the reader behavior are already explicit.

Proposed edit: Delete this final coordinated clause.

## c06-d02 — empty_framing

[docs/site/src/content/docs/atlas/schema-commands.mdx](https://github.com/stokaro/ptah/blob/654eae5591392278e6c8bce8e54737f780766f19/docs/site/src/content/docs/atlas/schema-commands.mdx)

```text
matching is the floor
rather than the ceiling, and a silent scope failure is the one answer Ptah
declines to reproduce
```

Metaphor and self-praise repeat the already stated divergence and concrete diagnostic behavior.

Proposed edit: Keep the measured Atlas behavior and the preceding Ptah refusal; remove the metaphor and moralized refusal.

## c06-d03 — vague_claims

[docs/site/src/content/docs/atlas/schema-commands.mdx](https://github.com/stokaro/ptah/blob/654eae5591392278e6c8bce8e54737f780766f19/docs/site/src/content/docs/atlas/schema-commands.mdx)

```text
A pattern that
deep is almost always an attempt to name a table
```

Makes an unmeasured prevalence claim about intent. The potential destructive interpretation is sufficient to explain refusal.

Proposed edit: A pattern that deep may be intended to name a table.

## c08-d01 — vague_claims

[docs/README.md](https://github.com/ocornut/imgui/blob/e5cb04b132cba94f902beb6186cb58b864777012/docs/README.md)

```text
bloat-free
```

An undefined efficiency claim; no size, cost or comparison defines bloat. The same paragraph separately specifies the useful dependency property.

Proposed edit: Delete bloat-free and retain the description and no-external-dependencies fact.

## c08-d02 — vague_claims

[docs/README.md](https://github.com/ocornut/imgui/blob/e5cb04b132cba94f902beb6186cb58b864777012/docs/README.md)

```text
It is less error prone (less code and less bugs) than traditional retained-mode interfaces
```

Broad comparative claims about code and bugs lack a task, measurement or condition; architectural state reduction does not establish the general comparison.

Proposed edit: Describe reduced state synchronization without claiming fewer bugs across retained-mode interfaces.

