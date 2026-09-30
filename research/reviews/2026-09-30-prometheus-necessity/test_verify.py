"""Make incomplete outputs and changed literal decisions fail verification."""

import copy
import unittest

import verify


class JudgeVerificationTests(unittest.TestCase):
    def setUp(self):
        self.rows = verify.read(verify.ROOT / "predictions.json")["rows"]
        self.expected = verify.read(verify.ROOT / "control-expectations.json")

    def test_saved_screen_has_all_original_cases(self):
        result = verify.decision(self.rows, self.expected)
        self.assertFalse(result["passed"])
        self.assertEqual(result["groups"]["useful"]["retained"], 9)
        self.assertEqual(result["groups"]["unsafe"]["retained"], 4)

    def test_missing_or_duplicate_identity_is_rejected(self):
        for rows in [self.rows[:-1], self.rows[:-1] + [self.rows[0]]]:
            with self.assertRaises(ValueError):
                verify.decision(rows, self.expected)

    def test_truncation_or_unavailable_is_not_safe_rejection(self):
        for field, value in [("finish_reason", "length"), ("status", "unavailable")]:
            rows = copy.deepcopy(self.rows)
            rows[0][field] = value
            with self.assertRaises(ValueError):
                verify.decision(rows, self.expected)

    def test_changed_grade_or_retention_is_rejected(self):
        for field, value in [("score", 5), ("retained", True), ("score", float("nan"))]:
            rows = copy.deepcopy(self.rows)
            rows[0][field] = value
            with self.assertRaises(ValueError):
                verify.decision(rows, self.expected)

    def test_ambiguous_or_extra_result_is_rejected(self):
        for text in ["[RESULT] 4", "An adequately long assessment. [RESULT] 4.0", "An adequately long assessment. [RESULT] 4 [RESULT] 5"]:
            rows = copy.deepcopy(self.rows)
            rows[0]["raw"] = text
            with self.assertRaises(ValueError):
                verify.decision(rows, self.expected)

    def test_changed_group_denominator_is_rejected(self):
        expected = copy.deepcopy(self.expected)
        expected["cases"][self.rows[0]["id"]]["group"] = "optional"
        with self.assertRaises(ValueError):
            verify.decision(self.rows, expected)


if __name__ == "__main__":
    unittest.main()
