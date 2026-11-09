"""Reject corrupted accounting and false promotion of exploratory evidence."""

import copy
import json
import unittest

import verify_result


class ResultVerifierTests(unittest.TestCase):
    def setUp(self):
        self.freeze = json.loads((verify_result.HERE / "freeze.json").read_text())
        self.result = json.loads((verify_result.HERE / "result.json").read_text())

    def test_frozen_result(self):
        self.assertEqual(verify_result.verify_file(), verify_result.RESULT_SHA256)

    def test_rejects_full_qualification_claim(self):
        altered = copy.deepcopy(self.result)
        altered["status"] = "qualified"
        with self.assertRaisesRegex(ValueError, "promoted"):
            verify_result.verify_data(altered, self.freeze)

    def test_rejects_protected_count_loss(self):
        altered = copy.deepcopy(self.result)
        altered["matches"]["negation-frame"]["excluded_protected"] -= 1
        with self.assertRaisesRegex(ValueError, "protected match"):
            verify_result.verify_data(altered, self.freeze)

    def test_rejects_invented_event(self):
        altered = copy.deepcopy(self.result)
        altered["previously_non_full_event_ids"].append("whole/p999-d99")
        altered["previously_non_full_event_ids"].sort()
        with self.assertRaisesRegex(ValueError, "non-FULL cohort"):
            verify_result.verify_data(altered, self.freeze)

    def test_rejects_changed_input_hash(self):
        altered = copy.deepcopy(self.result)
        altered["input_sha256"]["archive_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "input identity"):
            verify_result.verify_data(altered, self.freeze)


if __name__ == "__main__":
    unittest.main()
