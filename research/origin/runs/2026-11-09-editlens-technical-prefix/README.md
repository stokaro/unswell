# EditLens technical-document prefix screen

See [the frozen protocol](PROTOCOL.md), [the complete result](RESULT.md), and
[`freeze.json`](freeze.json). The independent aggregate check is:

```sh
python3 -B research/origin/runs/2026-11-09-editlens-technical-prefix/verify.py
```

This screen uses the saved long-prose-v3 archives and the same pinned model
as the [ARB screen](../2026-11-09-editlens-transfer/README.md). The scores
describe prefixes of technical documents; they do not calibrate a detector.
