"""Reject corrupted saved EditLens full-window evidence."""

from __future__ import annotations

import copy
import importlib.util
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("fullwindow_verify", HERE / "verify.py")
assert SPEC is not None and SPEC.loader is not None
verify = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(verify)


class SavedEvidenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.inventory = verify.load(verify.INVENTORY)
        cls.freeze = verify.load(HERE / "freeze.json")
        cls.documents = verify.load(HERE / "documents.json")
        cls.windows = verify.load(HERE / "window_scores.json")
        cls.pairs = verify.load(HERE / "pairs.json")
        cls.result = verify.load(HERE / "result.json")

    def check(self, *, documents=None, windows=None, pairs=None, result=None) -> dict:
        return verify.verify_payload(self.inventory, self.freeze,
                                     self.documents if documents is None else documents,
                                     self.windows if windows is None else windows,
                                     self.pairs if pairs is None else pairs,
                                     self.result if result is None else result)

    def test_complete_evidence(self) -> None:
        self.assertEqual(self.check(), {"sources": 38, "responses": 456,
                                        "windows": 6484, "generate_pass": True,
                                        "polish_pass": True})

    def test_changed_window_score_rejected(self) -> None:
        changed = copy.deepcopy(self.windows)
        changed[0]["score"] += 0.1
        with self.assertRaisesRegex(ValueError, "document mean"):
            self.check(windows=changed)

    def test_missing_pair_rejected(self) -> None:
        with self.assertRaisesRegex(ValueError, "Output coverage"):
            self.check(pairs=self.pairs[:-1])

    def test_changed_source_identity_rejected(self) -> None:
        changed = copy.deepcopy(self.documents)
        changed[0]["text_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "Document differs from inventory"):
            self.check(documents=changed)

    def test_false_gate_rejected(self) -> None:
        changed = copy.deepcopy(self.result)
        changed["operations"]["polish"]["direction_pass"] = False
        with self.assertRaisesRegex(ValueError, "result.operations.polish.direction_pass"):
            self.check(result=changed)


if __name__ == "__main__":
    unittest.main()
