"""Negative controls for the published source-context admission screen."""

import copy
import json
import unittest
from pathlib import Path
from unittest import mock

import verify_result


HERE = Path(__file__).resolve().parent
RESULT = json.loads((HERE / "result.json").read_text())
FREEZE = json.loads((HERE / "freeze.json").read_text())
CONTROL = json.loads((verify_result.CONTROL / "result-v2.json").read_text())


class TestPublishedScreen(unittest.TestCase):
    def test_valid_result(self):
        self.assertEqual(verify_result.verify_public_files()["whole_full_fixed"], 44)

    def test_false_qualification_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["qualified"] = True
        with self.assertRaisesRegex(ValueError, "False qualification"):
            verify_result.verify(changed, FREEZE, CONTROL)

    def test_changed_input_identity_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["input_sha256"]["originals"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "Input identities"):
            verify_result.verify(changed, FREEZE, CONTROL)

    def test_missing_fold_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["folds"].pop("ptah")
        with self.assertRaisesRegex(ValueError, "Repository folds"):
            verify_result.verify(changed, FREEZE, CONTROL)

    def test_changed_denominator_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["fixed_0_5"]["whole"]["events"] = 56
        with self.assertRaisesRegex(ValueError, "Event denominator"):
            verify_result.verify(changed, FREEZE, CONTROL)

    def test_false_numeric_pass_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["optimistic_numeric_feasible"] = True
        with self.assertRaisesRegex(ValueError, "Failure boundary"):
            verify_result.verify(changed, FREEZE, CONTROL)

    def test_changed_fraction_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["fixed_0_5"]["whole"]["fraction"] = 0.9
        with self.assertRaisesRegex(ValueError, "Fraction mismatch"):
            verify_result.verify(changed, FREEZE, CONTROL)

    def test_changed_frozen_script_is_rejected(self):
        actual_digest = verify_result.digest

        def altered_digest(path):
            if path == HERE / "scan.py":
                return "0" * 64
            return actual_digest(path)

        with mock.patch.object(verify_result, "digest", side_effect=altered_digest):
            with self.assertRaisesRegex(ValueError, "Frozen source identity"):
                verify_result.verify_public_files()


if __name__ == "__main__":
    unittest.main()
