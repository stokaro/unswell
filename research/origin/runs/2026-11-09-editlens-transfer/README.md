# EditLens source-matched ARB screen

See [the frozen protocol](PROTOCOL.md), [the complete result](RESULT.md), and
[`freeze.json`](freeze.json). The independent aggregate check is:

```sh
python3 -B research/origin/runs/2026-11-09-editlens-transfer/verify.py
```

Replaying inference requires the pinned public ARB Parquet, the attributed
EditLens FP32 ONNX checkpoint, and the Python dependencies named in the
protocol. Raw corpus text and model weights are not stored here.
