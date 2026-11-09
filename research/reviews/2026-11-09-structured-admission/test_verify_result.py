"""Negative controls for the frozen #349 structured-admission accounting."""

import copy
import json
import unittest
from pathlib import Path

from verify_result import verify, verify_freezes


RESULT = json.loads(Path(__file__).with_name("result-v2.json").read_text())


class TestPublishedScreen(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.freeze = verify_freezes()

    def test_valid_result(self):
        self.assertEqual(verify(RESULT, self.freeze)["whole_full"], 45)

    def test_false_qualification_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["qualified"] = True
        with self.assertRaisesRegex(ValueError, "False qualification"):
            verify(changed, self.freeze)

    def test_changed_source_identity_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["input_sha256"]["reviews"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "Input identities"):
            verify(changed, self.freeze)

    def test_omitted_fold_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["folds"].pop("ptah")
        with self.assertRaisesRegex(ValueError, "Repository folds"):
            verify(changed, self.freeze)

    def test_changed_denominator_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["fixed_0_5"]["whole"]["events"] = 56
        with self.assertRaisesRegex(ValueError, "Event denominator"):
            verify(changed, self.freeze)

    def test_fraction_tamper_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["optimistic_best_at_85_percent"]["whole"]["fraction"] = 0.9
        with self.assertRaisesRegex(ValueError, "Best threshold accounting"):
            verify(changed, self.freeze)

    def test_false_feasible_claim_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["optimistic_numeric_feasible"] = True
        with self.assertRaisesRegex(ValueError, "Failure boundary"):
            verify(changed, self.freeze)


if __name__ == "__main__":
    unittest.main()
