"""Reject false success or source identity claims in the public screen."""

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
        self.assertFalse(verify_result.verify_public_files()["screen_passed"])

    def test_rejects_changed_archive(self) -> None:
        value = copy.deepcopy(self.result)
        value["archive_sha256"] = "0" * 64
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)

    def test_rejects_false_screen_pass(self) -> None:
        value = copy.deepcopy(self.result)
        value["screen_passed"] = True
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)

    def test_rejects_missing_historical_event(self) -> None:
        value = copy.deepcopy(self.result)
        value["cohorts"]["confirmation"]["events"] = 19
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)

    def test_rejects_fabricated_location_gain(self) -> None:
        value = copy.deepcopy(self.result)
        value["cohorts"]["whole"]["previously_nonfull_location_covered"] = 8
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)

    def test_rejects_qualification_claim(self) -> None:
        value = copy.deepcopy(self.result)
        value["qualified"] = True
        with self.assertRaises(ValueError):
            verify_result.verify(value, self.freeze)


if __name__ == "__main__":
    unittest.main()
