# Origin packs

A pack is a fitted model in the shape the scanner loads. This directory holds
the ones that beat their baseline. Nothing here decides a gate.

## unswell-origin-lexical-v1

Fitted in [the lexical run](../runs/2026-09-13-lexical/README.md) and packed
from its artifact without change. It estimates similarity to a training class
of generated documentation. It is not a quality judgment, not a verdict that a
text was written by a tool, and not a share of any text.

| Property | Value |
| --- | --- |
| Task | `origin_endpoint` |
| Declared status | experimental, corpus `not_qualified` |
| Unit kind | paragraph |
| Columns | 128 n-gram counts from its own frozen vocabulary |
| Word band | 25 to 89 |
| Development false-flag rate | 0.040 on real code comments |
| Development recall | 0.744 |

### What it abstains on

Everything outside the band. The fit admitted units of 25 to 89 words and says
nothing above or below, so the pack reports `insufficient_evidence` there
rather than extrapolating. On a documentation tree that is most paragraphs.

It also abstains when the scan's extraction and preparation policy differs from
the one the corpus used, because the same text prepared differently yields
different columns. The policy below is the one it was fitted under.

### Running it

```yaml
# unswell.yaml
version: 1
extends: [builtin:strict-v1]
language: en
extraction:
  contexts: [comment, heading, list-item, paragraph, string, table-cell]
origin:
  model: pack
  accept_experimental: true
  on_incompatible: unavailable
```

```sh
unswell check --config unswell.yaml \
  --origin-model research/origin/packs/unswell-origin-lexical-v1.json .
```

Set `on_incompatible: fail` while setting this up: an incompatible policy is
otherwise silent, and every unit abstains without saying why.

### What it cannot tell you

Its negative class is human comments in code, which are terser than prose
documentation. A prose page scoring above a comment corpus is partly a
statement about genre. Reading a rate on prose as a false-flag rate overstates
what was measured.

The reported numbers are from the development partition. The confirmation
partition has a controlled arm and a measured corpus but is not yet scored, so
no confirmation number exists. The run record says why.
