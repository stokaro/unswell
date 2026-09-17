# Added and removed diagnostics

All added findings are listed below. There are no removals. Partial credit
does not enter event recall. Indices are zero-based within the saved report.
Identical evidence across profiles shares one entry and one review note.

## development / repetition.repeated-claim / 41

technical: 41; strict: 46

Review: [n1](#n1)

- sources/p004.md [3877, 3922): `The command exits '1' because drift was found`
- sources/p004.md [4158, 4203): `The command exits '1' because drift was found`

Full events: ['p004-d01']. Partial events: [].

## exposed_repetition / repetition.adjacent-word / 25

technical: 25; strict: 31

Review: [n2](#n2)

- sources/c04.md [3587, 3590): `See`
- sources/c04.md [3591, 3594): `See`

Full events: ['c04-d01']. Partial events: [].

## exposed_repetition / repetition.adjacent-word / 26

technical: 26; strict: 32

Review: [n2](#n2)

- sources/c04.md [3944, 3947): `See`
- sources/c04.md [3948, 3951): `See`

Full events: ['c04-d02']. Partial events: [].

## exposed_repetition / repetition.adjacent-word / 27

technical: 27; strict: 33

Review: [n2](#n2)

- sources/c06.md [5281, 5289): `requires`
- sources/c06.md [5290, 5298): `requires`

Full events: ['c06-d01']. Partial events: [].

## exposed_repetition / repetition.explanatory-restart / 30

technical: 30; strict: 36

Review: [n3](#n3)

- sources/c06.md [15011, 15056): `The reasons for why this is are quite complex`
- sources/c06.md [15062, 15078): `they are complex`

Full events: []. Partial events: ['c06-d02'].

## Review notes

### n1

Both source ranges repeat the exit-1 drift explanation within one section. The exit operand and causal condition are identical; the distinct exit-0 case is not grouped.

### n2

Both adjacent copies of the same word are located; removing the second preserves the surrounding instruction or requirement.

### n3

The finding locates an unnecessary restart of the same complexity judgment. It does not identify the final complex-optimizations rationale or replace it with the missing concrete costs; c06-d02 remains a partial match and receives no full event credit.
