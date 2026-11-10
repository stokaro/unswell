"""Negative accounting tests for the repair-first pilot scorer."""

import copy
import unittest

import score


class RepairScoringTests(unittest.TestCase):
    def setUp(self):
        self.reference = {
            "events": [{"id": f"event-{n}", "page": "page.md",
                        "span": {"start": n * 20, "end": n * 20 + 10}}
                       for n in range(4)],
            "retain_controls": [{"id": f"control-{n}", "page": "page.md",
                                 "span": {"start": 200 + n * 20,
                                          "end": 210 + n * 20}}
                                for n in range(4)],
        }
        self.bound = [{"page_id": "page.md",
                       "edits": [{"edit_id": "page.md/edit-01",
                                  "span": {"start": 0, "end": 10}}]}]
        self.review = {
            "edits": [{"edit_id": "page.md/edit-01", "status": "accepted",
                       "safety": "yes", "rationale": "A necessary safe edit."}],
            "events": [{"event_id": f"event-{n}",
                        "status": "FULL" if n == 0 else "MISSED",
                        "credited_edit_ids": ["page.md/edit-01"] if n == 0 else [],
                        "rationale": "Source-specific review."}
                       for n in range(4)],
        }

    def test_complete_review_counts(self):
        result = score.evaluate(self.bound, self.review, self.reference)
        self.assertEqual(result["full_events"], 1)
        self.assertEqual(result["accepted_safe_edits"], 1)
        self.assertFalse(result["advance"])

    def test_unsafe_edit_cannot_receive_full(self):
        review = copy.deepcopy(self.review)
        review["edits"][0]["safety"] = "no"
        with self.assertRaisesRegex(ValueError, "FULL lacks accepted safe edit"):
            score.evaluate(self.bound, review, self.reference)

    def test_full_requires_target_overlap(self):
        bound = copy.deepcopy(self.bound)
        bound[0]["edits"][0]["span"] = {"start": 100, "end": 110}
        with self.assertRaisesRegex(ValueError, "does not overlap"):
            score.evaluate(bound, self.review, self.reference)

    def test_missing_delivery_review_rejected(self):
        review = copy.deepcopy(self.review)
        review["edits"] = []
        with self.assertRaisesRegex(ValueError, "not every edit reviewed"):
            score.evaluate(self.bound, review, self.reference)

    def test_control_overlap_is_reported(self):
        bound = copy.deepcopy(self.bound)
        bound[0]["edits"].append({"edit_id": "page.md/edit-02",
                                  "span": {"start": 200, "end": 210}})
        review = copy.deepcopy(self.review)
        review["edits"].append({"edit_id": "page.md/edit-02",
                                "status": "rejected", "safety": "no",
                                "rationale": "Control must be retained."})
        result = score.evaluate(bound, review, self.reference)
        self.assertEqual(result["retain_control_hits"]["control-0"],
                         ["page.md/edit-02"])
        self.assertFalse(result["advance"])


if __name__ == "__main__":
    unittest.main()
