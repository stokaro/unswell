# Added and removed diagnostics

All added findings are listed below. There are no removals. Partial credit
does not enter event recall. Indices are zero-based within the saved report.
Identical evidence across profiles shares one entry and one review note.

## exposed_local / filler.document-justification / 0

technical: 0; strict: 0

Review: [n1](#n1)

- sources/c02.md [656, 760): `This page is what each of them means, and how to find out whether they are right before anything is sent`

Full events: ['c02-d01']. Partial events: [].

## exposed_local / filler.unscoped-assurance / 20

technical: 20; strict: 25

Review: [n2](#n2)

- sources/c05.md [13108, 13125): `nobody can act on`

Full events: ['c05-d09']. Partial events: [].

## exposed_local / filler.unscoped-assurance / 25

technical: 25; strict: 30

Review: [n3](#n3)

- sources/c05.md [18893, 18918): `nobody could have checked`

Full events: ['c05-d12']. Partial events: [].

## exposed_local / filler.evaluative-closure / 32

technical: 32; strict: 38

Review: [n4](#n4)

- sources/c06.md [6218, 6291): `Several ClickHouse-specific points are worth knowing before adopting them`

Full events: ['c06-d03']. Partial events: [].

## exposed_local / filler.unscoped-assurance / 40

technical: 40; strict: 46

Review: [n5](#n5)

- sources/c08.md [5936, 6023): `Caddy exposes an unprecedented level of control compared to any web server in existence`
- sources/c08.md [6201, 6321): `Caddy is also ridiculously extensible, with a powerful plugin system that makes vast improvements over other web servers`

Full events: ['c08-d08', 'c08-d09']. Partial events: [].

## Review notes

### n1

The page announces what its own explanation means; the finding identifies the removable self-description.

### n2

The unqualified nobody-can-act assertion overstates the consequence; the warning locates that reader claim.

### n3

The warning identifies the unrestricted nobody-could-have-checked assertion; the edit retains the actual verification limit.

### n4

The warning identifies the cognitive-worth announcement before the list; removing the preface preserves the technical points.

### n5

The grouped warning locates both the all-competitor control claim and the exaggerated extensibility claim. Both frozen defects receive credit, once each.
