"""Negative tests for the saved control-decision verifier."""

from __future__ import annotations

import json
from pathlib import Path
import shutil
import tempfile
import unittest

from verify_score import verify


HERE = Path(__file__).resolve().parent
OUTPUTS = ("calibration.json", "documents.json", "window_scores.json", "result.json")


class VerifyScoreTest(unittest.TestCase):
    def test_committed_result(self) -> None:
        result = verify(HERE)
        self.assertTrue(result["control_gate_pass"])
        self.assertEqual(result["test_flagged"], 1)

    def test_changed_window_score_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory)
            for name in OUTPUTS:
                shutil.copy2(HERE / name, target / name)
            path = target / "window_scores.json"
            windows = json.loads(path.read_text())
            windows[0]["score"] = 0.123456789
            path.write_text(json.dumps(windows))
            with self.assertRaisesRegex(ValueError, "Document mean"):
                verify(target)

    def test_changed_gate_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory)
            for name in OUTPUTS:
                shutil.copy2(HERE / name, target / name)
            path = target / "result.json"
            result = json.loads(path.read_text())
            result["control_gate_pass"] = False
            path.write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "Control gate"):
                verify(target)


if __name__ == "__main__":
    unittest.main()
