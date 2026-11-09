"""Reject false reachability claims in the public #349 audit."""

from __future__ import annotations

import copy
import json
from pathlib import Path
import unittest

import verify_result


HERE = Path(__file__).resolve().parent


class VerifyResultTest(unittest.TestCase):
    def setUp(self) -> None:
        self.result = json.loads((HERE / "result.json").read_text())
        self.freeze = json.loads((HERE / "freeze.json").read_text())

    def test_public_result(self) -> None:
        self.assertEqual(verify_result.verify_public_files()["zero_target_events"], 0)

    def test_rejects_changed_archive(self) -> None:
        value = copy.deepcopy(self.result)
        value["archive_sha256"] = "0" * 64
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)

    def test_rejects_missing_event(self) -> None:
        value = copy.deepcopy(self.result)
        value["cohorts"]["whole"]["events"] = 56
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)

    def test_rejects_false_unavailability(self) -> None:
        value = copy.deepcopy(self.result)
        value["available_pages"] = 35
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)

    def test_rejects_fabricated_full_coverage(self) -> None:
        value = copy.deepcopy(self.result)
        value["cohorts"]["whole"]["all_targets_full"] = 57
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)

    def test_rejects_false_qualification(self) -> None:
        value = copy.deepcopy(self.result)
        value["qualified"] = True
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)


if __name__ == "__main__":
    unittest.main()
