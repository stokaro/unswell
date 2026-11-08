"""Negative controls for the frozen mixed-authorship result verifier."""

from __future__ import annotations

import json
import tempfile
import unittest
from pathlib import Path

from verify import HERE, verify


class VerifyTest(unittest.TestCase):
    def test_committed_result(self) -> None:
        verify(HERE / "result.json")

    def test_rejects_changed_metric(self) -> None:
        result = json.loads((HERE / "result.json").read_text())
        result["span"]["richness"]["overall"]["auroc"] = 0.99
        with tempfile.TemporaryDirectory() as temp:
            path = Path(temp) / "changed.json"
            path.write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "Result hash changed"):
                verify(path)

    def test_rejects_truncated_result(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            path = Path(temp) / "truncated.json"
            path.write_text("{}")
            with self.assertRaisesRegex(ValueError, "Result hash changed"):
                verify(path)


if __name__ == "__main__":
    unittest.main()
