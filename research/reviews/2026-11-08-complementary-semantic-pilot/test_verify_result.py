"""Negative checks for complete-stream evidence verification."""

import copy
import unittest
from unittest import mock

import score
import verify_result


class ResultTests(unittest.TestCase):
    def test_saved_result_verifies(self):
        self.assertFalse(verify_result.verify()["advance"])

    def test_missing_finding_review_is_rejected(self):
        original = score.read

        def altered(path):
            value = original(path)
            if path == score.ROOT / "review.json":
                value = copy.deepcopy(value)
                value["findings"].pop()
            return value

        with mock.patch.object(score, "read", side_effect=altered):
            with self.assertRaisesRegex(ValueError, "complete stream"):
                score.score()

    def test_fabricated_positive_result_is_rejected(self):
        original = verify_result.read

        def altered(path):
            value = original(path)
            if path == verify_result.ROOT / "result.json":
                value = dict(value, advance=True)
            return value

        with mock.patch.object(verify_result, "read", side_effect=altered):
            with self.assertRaisesRegex(ValueError, "result arithmetic"):
                verify_result.verify()

    def test_tool_trace_failure_is_propagated(self):
        with mock.patch.object(verify_result.run, "trace", side_effect=ValueError("tool use")):
            with self.assertRaisesRegex(ValueError, "tool use"):
                verify_result.verify()


if __name__ == "__main__":
    unittest.main()
