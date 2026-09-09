# Bash empty assignments

Issue [#62](https://github.com/stokaro/unswell/issues/62) was reproduced with
gotreesitter v0.52.0 on `! A=x B= command` and Ptah's installation-check script at
commit `8873cca616bcd021e5d6588037076fc9d7001c1a`. The scanner treated the command
word as the value of `B` and could insert a missing command-name node. Some related
inputs produced an incorrect tree without an error flag.

## Cause and repair

The Go Bash scanner probes for an opening parenthesis before handling
`_empty_value`. That probe skips horizontal whitespace when `_concat` is invalid,
even when it does not find a parenthesis. The later empty-value check sees the
next word instead of the whitespace after `=`. A trace of the minimal input shows
`_empty_value` allowed at byte 8 and the scanner returning no token there.

The original C scanner checks the whitespace boundary without that preliminary
probe. A local two-condition patch to the Go scanner, preserving whitespace when
an empty value is allowed, repaired all six initial controls and the Ptah script.
The current upstream revision `851945aa6fec19e1808acaf181334084127ec6be` still failed
the minimal case and the complete script when checked on September 9, 2026.

Unswell's adapter recognizes the existing zero-width `_empty_value` token when
the parser allows it, `_concat` is invalid, and the current character is ASCII
whitespace. All other scanning uses the original scanner. Token IDs are resolved
from the loaded grammar instead of hard-coded. The adapter operates on an owned
instance decoded from the same embedded grammar, preserving the global cached
grammar for other consumers. It changes no input bytes or grammar productions.

## Reference comparison

The [saved comparison](bash-empty-values.json) records 40 structural probes and
the unchanged Ptah script. They include negation, multiple assignments, empty and
nonempty values, tabs, CRLF, Unicode, substitutions, redirects, pipelines, case
statements, heredocs, and incomplete syntax. These are parser tests, not editorial
or authorship labels.

The reference was compiled from tree-sitter-bash
`a06c2e4415e9bc0346c6b86d401879ffb44058f7`, the revision locked by gotreesitter
v0.52.0, and loaded with Python tree-sitter 0.25.2 in an isolated research
environment. The delivered runtime and normal CI require neither C nor Python.
GNU Bash 3.2.57 syntax-only checks also accepted the 38 complete structural
probes and rejected the two incomplete ones. No shell input or Ptah program was
executed.

| Comparison with C | Pinned Go scanner | Adapter |
| --- | ---: | ---: |
| Matching error status | 32 / 41 | 41 / 41 |
| Matching named-node trees and byte ranges | 19 / 41 | 39 / 41 |

The tree projection recursively records named node kinds, byte spans, missing
flags, and ordered named children. It does not compare anonymous tokens or field
labels. Hashes use SHA-256 of JSON with sorted keys and compact separators.
The entire named-node projection of the Ptah script matches the C result after
the repair. The two remaining differences involve `${x:-some text}` and
`${x:=some text}`; their adapter output is identical to the original Go output.
This repair does not claim to resolve those separate parser differences.

An additional probe with an escaped newline inside a nonempty assignment and
surrounding CRLF passes Bash syntax checking but retains the same missing command
node in all three parser paths. Its source and result are recorded separately
from the 41-case comparison. That grammar limitation remains an explicit error;
it is outside the empty-value repair.

The C source, compiled reference, raw projections, trace, and upstream probe are
retained in the local `parser-issue62` research directory. The JSON file records
their relevant results; the research environment is not part of the product.

## Product regression tests

Blackbox extraction tests exercise Bash, POSIX shell, and the supported Zsh
subset, checking comment and string preservation, Unicode source spans,
structural context, substitution boundaries, and invalid-source errors. A test
parses the same input through the original gotreesitter grammar before and after
Unswell to verify that the shared grammar was not changed.

Root CLI fixtures require six exact detections across three shell formats, with
CRLF and protected substitutions. An incomplete `if` statement remains an
operational failure with exit code 2. All expectations run through the normal
engine and reporting path.
