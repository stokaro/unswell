# Remaining confirmation defects

All five defects below remain missed. Adjacent-word repetitions, repeated
clauses and paraphrases are distinct from complete list-item duplicates.
The labels were frozen before diagnostic outputs were opened.

## c04-d01: adjacent_word

[nats-io/nats-server/ROADMAP.md](https://github.com/nats-io/nats-server/blob/7c7662603115c1f53a9bacfa0115893c8ebe6fd4/ROADMAP.md)

```text
See
```

```text
See
```

The adjacent imperative is copied twice without a second instruction. This is a local editing defect regardless of the issue target.

Proposed edit: Delete one occurrence of See; preserve the issue link.

## c04-d02: adjacent_word

[nats-io/nats-server/ROADMAP.md](https://github.com/nats-io/nats-server/blob/7c7662603115c1f53a9bacfa0115893c8ebe6fd4/ROADMAP.md)

```text
See
```

```text
See
```

The adjacent imperative is copied twice without a second instruction. This is a local editing defect regardless of the issue target.

Proposed edit: Delete one occurrence of See; preserve the issue link.

## c06-d01: adjacent_word

[BurntSushi/ripgrep/FAQ.md](https://github.com/BurntSushi/ripgrep/blob/7cb211378a2ac6d421c5f6f3f71411937af23136/FAQ.md)

```text
requires
```

```text
requires
```

The same verb is accidentally adjacent; it contributes no additional requirement.

Proposed edit: Delete the second requires and retain the list of binaries.

## c06-d02: within_sentence

[BurntSushi/ripgrep/FAQ.md](https://github.com/BurntSushi/ripgrep/blob/7cb211378a2ac6d421c5f6f3f71411937af23136/FAQ.md)

```text
The reasons for why this is are quite complex,
```

```text
and they are complex because the optimizations that ripgrep uses to implement
fast search are complex.
```

The explanation repeats complex three times instead of naming a source of cost. The next paragraphs actually identify line-by-line searching and Unicode validation.

Proposed edit: Replace the two clauses with a direct preview of line-by-line search and Unicode validation, preserving the subsequent explanation.

## c06-d03: within_block

[BurntSushi/ripgrep/FAQ.md](https://github.com/BurntSushi/ripgrep/blob/7cb211378a2ac6d421c5f6f3f71411937af23136/FAQ.md)

```text
In particular, ripgrep's dependencies (direct and transitive)
will always be limited to permissive licenses.
```

```text
That is, ripgrep will never
depend on code that is not permissively licensed.
```

The second sentence states the same permissive-only constraint by negating its complement. The following copyleft examples and weak-copyleft boundary already explain the scope.

Proposed edit: Delete the That is sentence; retain the direct/transitive scope and the copyleft examples.
