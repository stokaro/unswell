"""Negative controls for the post-result RACE verifier."""

from __future__ import annotations

import importlib.util
from pathlib import Path
import unittest


SCRIPT = Path(__file__).with_name("verify_result.py")
SPEC = importlib.util.spec_from_file_location("race_verify_result", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
VERIFY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VERIFY)


class VerifyResultTests(unittest.TestCase):
    def test_rejects_fabricated_success(self) -> None:
        with self.assertRaisesRegex(ValueError, "Saved result differs"):
            VERIFY.compare_result({"status": "advance"}, {"status": "retire"})

    def test_brute_auc_counts_ties(self) -> None:
        self.assertEqual(VERIFY.brute_auc([0.8, 0.5], [0.2, 0.5]), 0.875)


if __name__ == "__main__":
    unittest.main()
