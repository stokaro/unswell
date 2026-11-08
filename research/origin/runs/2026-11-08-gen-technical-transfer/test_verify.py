"""Exercise transfer verification failure paths and fixed threshold semantics."""

from __future__ import annotations

import importlib.util
import json
import shutil
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent


def load(name: str):
    spec = importlib.util.spec_from_file_location(name, HERE / f"{name}.py")
    assert spec is not None and spec.loader is not None
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


runner = load("run")
verifier = load("verify")


class TransferTest(unittest.TestCase):
    def test_fixed_threshold_is_inclusive(self):
        thresholds = {"richness": 0.9, "entropy": 0.96}
        self.assertEqual(runner.flags({"richness": 0.9, "entropy": 0.95}, thresholds),
                         {"richness": True, "entropy": False})
        self.assertIsNone(runner.flags(None, thresholds))

    def test_advance_requires_every_gate(self):
        result = json.loads((HERE / "result.json").read_text())
        self.assertFalse(result["decision"]["advance_screen"])
        self.assertFalse(result["decision"]["gates"][
            "technical_generate_flag_at_least_10_percent"])
        self.assertTrue(result["decision"]["gates"][
            "positive_repository_means_at_least_15"])

    def test_mutated_result_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            temporary = Path(directory)
            for name in ("freeze.json", "decision.json", "result.json",
                         "PROTOCOL.md", "run.py"):
                shutil.copy2(HERE / name, temporary / name)
            result_path = temporary / "result.json"
            result = json.loads(result_path.read_text())
            result["technical"]["operations"]["generate"]["overall"][
                "richness"]["flagged"] += 1
            result_path.write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "Saved result changed"):
                verifier.verify(temporary)

    def test_mutated_protocol_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            temporary = Path(directory)
            for name in ("freeze.json", "decision.json", "result.json",
                         "PROTOCOL.md", "run.py"):
                shutil.copy2(HERE / name, temporary / name)
            (temporary / "PROTOCOL.md").write_text("changed")
            with self.assertRaisesRegex(ValueError, "protocol changed"):
                verifier.verify(temporary)


if __name__ == "__main__":
    unittest.main()
