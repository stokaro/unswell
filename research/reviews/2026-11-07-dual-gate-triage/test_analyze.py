"""Reject incomplete or changed inputs to the exposed #349 gate analysis."""

import copy
import unittest

import analyze


class TriageTest(unittest.TestCase):
    def test_frozen_review_matches_record(self):
        self.assertEqual(analyze.analyze(analyze.load_inputs())["profiles"]["technical"]["whole"]
                         ["perfect_new_findings_needed_without_removal"], 172)

    def test_changed_finding_decision_is_rejected(self):
        inputs = copy.deepcopy(analyze.load_inputs())
        inputs["findings.json"][0]["status"] = "actionable"
        with self.assertRaisesRegex(ValueError, "finding summary mismatch"):
            analyze.analyze(inputs)

    def test_missing_event_is_rejected(self):
        inputs = copy.deepcopy(analyze.load_inputs())
        inputs["events.json"].pop()
        with self.assertRaisesRegex(ValueError, "all 123 events"):
            analyze.analyze(inputs)


if __name__ == "__main__":
    unittest.main()
