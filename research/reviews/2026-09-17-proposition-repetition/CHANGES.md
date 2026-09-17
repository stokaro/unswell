# Reviewed changes

Indices refer to the retained strict reports.

## exposed_repetition / added

### 37: repetition.explanatory-restart

The source-bound warning now includes all three complex judgments, including the circular because clause. Retain the named optimizations but replace the repeated adjective with a concrete cause.

Status: actionable. Full: c06-d02. Partial: none.

`sources/c06.md` bytes 15011–15056

```text
The reasons for why this is are quite complex
```

`sources/c06.md` bytes 15062–15078

```text
they are complex
```

`sources/c06.md` bytes 15087–15159

```text
the optimizations that ripgrep uses to implement
fast search are complex
```

### 47: repetition.repeated-claim

The dependency owner, full direct-and-transitive scope and permissive licensing property match across the positive restriction and explicit negative reformulation. State the license restriction once; do not remove its scope.

Status: actionable. Full: c06-d03. Partial: none.

`sources/c06.md` bytes 34691–34798

```text
In particular, ripgrep's dependencies (direct and transitive)
will always be limited to permissive licenses
```

`sources/c06.md` bytes 34800–34876

```text
That is, ripgrep will never
depend on code that is not permissively licensed
```

## exposed_repetition / removed

### 37: repetition.explanatory-restart

The finding locates an unnecessary restart of the same complexity judgment. It does not identify the final complex-optimizations rationale or replace it with the missing concrete costs; c06-d02 remains a partial match and receives no full event credit.

Status: actionable. Full: none. Partial: c06-d02.

`sources/c06.md` bytes 15011–15056

```text
The reasons for why this is are quite complex
```

`sources/c06.md` bytes 15062–15078

```text
they are complex
```

## exposed_scope / added

### 36: repetition.definition-echo

The identical typing action and both-flags object recur on each side of the copula. The bounded shorthand resolves to that same object; remove the repeated judgment and retain the actual flag interaction.

Status: actionable. Full: c03-d12. Partial: none.

`sources/c03.md` bytes 64997–65035

```text
Typing both flags is
still typing both
```
