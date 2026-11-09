"""Negative checks for the published semantic admission result."""

from __future__ import annotations

import copy
import json
from pathlib import Path
import unittest

import verify_result


HERE = Path(__file__).resolve().parent


class PublishedResultTest(unittest.TestCase):
    def setUp(self) -> None:
        self.result = json.loads((HERE / "result.json").read_text())
        self.freeze = json.loads((HERE / "freeze.json").read_text())
        self.control = json.loads((HERE.parent / "2026-11-09-structured-admission" /
                                   "result-v2.json").read_text())
        self.oracle = json.loads((HERE / "oracle.json").read_text())

    def test_public_result(self) -> None:
        self.assertFalse(verify_result.verify_public_files()["qualified"])

    def test_rejects_changed_input(self) -> None:
        modified = copy.deepcopy(self.result)
        modified["input_sha256"]["cores"] = "0" * 64
        with self.assertRaises(ValueError):
            verify_result.verify(modified, self.freeze, self.control)

    def test_rejects_false_gate_pass(self) -> None:
        modified = copy.deepcopy(self.result)
        modified["qualified"] = True
        with self.assertRaises(ValueError):
            verify_result.verify(modified, self.freeze, self.control)

    def test_rejects_changed_denominator(self) -> None:
        modified = copy.deepcopy(self.result)
        modified["fixed_0_5"]["whole"]["events"] = 56
        with self.assertRaises(ValueError):
            verify_result.verify(modified, self.freeze, self.control)

    def test_rejects_invented_acceptance(self) -> None:
        modified = copy.deepcopy(self.result)
        modified["fixed_0_5"]["whole"]["fraction"] = 0.86
        with self.assertRaises(ValueError):
            verify_result.verify(modified, self.freeze, self.control)

    def test_rejects_missing_fold(self) -> None:
        modified = copy.deepcopy(self.result)
        modified["folds"].pop(next(iter(modified["folds"])))
        with self.assertRaises(ValueError):
            verify_result.verify(modified, self.freeze, self.control)

    def test_rejects_invented_oracle_coverage(self) -> None:
        modified = copy.deepcopy(self.oracle)
        modified["accepted_only"]["whole"]["full"] = 49
        with self.assertRaises(ValueError):
            verify_result.verify_oracle(modified, self.control)

    def test_rejects_oracle_as_selector(self) -> None:
        modified = copy.deepcopy(self.oracle)
        modified["scope"] = "deployable_selector"
        with self.assertRaises(ValueError):
            verify_result.verify_oracle(modified, self.control)


if __name__ == "__main__":
    unittest.main()
