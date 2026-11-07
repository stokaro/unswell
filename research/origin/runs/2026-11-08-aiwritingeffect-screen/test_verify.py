"""Negative controls for frozen research artifact verification."""

from __future__ import annotations

import contextlib
import importlib.util
import io
import json
import shutil
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent


def load(name: str):
    spec = importlib.util.spec_from_file_location(name, HERE / name)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"Cannot load {name}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class VerifyTest(unittest.TestCase):
    def test_saved_artifacts_pass(self):
        with contextlib.redirect_stdout(io.StringIO()):
            load("verify.py").verify()

    def test_changed_protocol_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            copy = Path(temporary) / "screen"
            shutil.copytree(HERE, copy, ignore=shutil.ignore_patterns("__pycache__"))
            (copy / "PROTOCOL.md").write_text(
                (copy / "PROTOCOL.md").read_text() + "\nChanged after freeze.\n"
            )
            with self.assertRaisesRegex(ValueError, "PROTOCOL.md hash mismatch"):
                load("verify.py").verify(copy)

    def test_changed_result_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            copy = Path(temporary) / "screen"
            shutil.copytree(HERE, copy, ignore=shutil.ignore_patterns("__pycache__"))
            path = copy / "result.json"
            result = json.loads(path.read_text())
            result["richness_feasibility_pass"] = True
            path.write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "result.json hash mismatch"):
                load("verify.py").verify(copy)

    def test_auc_endpoints_and_ties(self):
        for name in ("run.py", "run_docs.py"):
            module = load(name)
            for positive, negative, expected in (
                (1.0, 0.0, 1.0),
                (0.0, 1.0, 0.0),
                (0.5, 0.5, 0.5),
            ):
                rows = [{"ai": {"richness": positive},
                         "human": {"richness": negative}}] * 20
                self.assertEqual(module.auc(rows, "richness"), expected)


if __name__ == "__main__":
    unittest.main()
