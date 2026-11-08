"""Negative integrity checks for the frozen source-only audit."""

from __future__ import annotations

import json
import shutil
import tempfile
import unittest
from pathlib import Path

import evaluate
import prepare


class EvaluateTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.packet = self.root / "packet"
        prepare.build(self.packet)
        self.study = self.root / "study"
        self.study.mkdir()
        for name in ("freeze.json", "key.json", "review.json", "review-seal.json"):
            shutil.copy2(evaluate.HERE / name, self.study / name)

    def test_frozen_result(self) -> None:
        result = evaluate.evaluate(self.packet, self.study)
        self.assertEqual(result["separated_pairs"], 2)
        self.assertTrue(result["witness_split_needed"])
        self.assertFalse(result["screen_passed"])

    def test_changed_review_is_refused(self) -> None:
        path = self.study / "review.json"
        review = json.loads(path.read_text())
        review["decisions"][0]["decision"] = "split_needed"
        path.write_text(json.dumps(review))
        with self.assertRaisesRegex(ValueError, "Review changed after sealing"):
            evaluate.evaluate(self.packet, self.study)

    def test_changed_source_is_refused(self) -> None:
        path = self.packet / "sources/page-01.source"
        path.write_bytes(path.read_bytes() + b"\n")
        with self.assertRaisesRegex(ValueError, "Original source hash mismatch"):
            evaluate.evaluate(self.packet, self.study)


if __name__ == "__main__":
    unittest.main()
