"""Negative controls for the published #349 triage accounting."""

import copy
import json
import unittest
from pathlib import Path

from verify_result import verify


RESULT = json.loads(Path(__file__).with_name("result.json").read_text())


class TestPublishedAudit(unittest.TestCase):
    def test_valid_result(self):
        self.assertEqual(verify(RESULT)["whole_full"], 45)

    def test_false_pass_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["qualified"] = True
        with self.assertRaisesRegex(ValueError, "False qualification"):
            verify(changed)

    def test_changed_source_identity_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["input_sha256"]["reviews"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "Input identities"):
            verify(changed)

    def test_omitted_fold_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["leave_one_repository_out"]["per_group"].pop("ptah")
        with self.assertRaisesRegex(ValueError, "Repository folds"):
            verify(changed)

    def test_fraction_tamper_is_rejected(self):
        changed = copy.deepcopy(RESULT)
        changed["leave_one_repository_out"]["heldout"]["whole"]["fraction"] = 0.9
        with self.assertRaisesRegex(ValueError, "Fraction mismatch"):
            verify(changed)


if __name__ == "__main__":
    unittest.main()
