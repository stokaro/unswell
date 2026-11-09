"""Negative controls for the frozen whole-stream scorer."""

import unittest

import run
import score


class ScoreTest(unittest.TestCase):
    def packet(self):
        bound = {}
        for number in run.PAGES:
            inp = run.payload(number)
            bound[inp["page_id"]] = {
                "page_id": inp["page_id"],
                "source_sha256": inp["source_sha256"],
                "reviews": [{"id": finding["id"], "decision": "keep",
                             "reason": "The source supports this diagnosis."}
                            for finding in inp["baseline_findings"]],
                "new_findings": [],
            }
        review = {
            "format": "unswell-349-joint-pilot-source-review-v1",
            "reviewer": "root Codex assistant; not a human annotator",
            "bound_sha256": {name: run.digest(run.encoded(page))
                             for name, page in bound.items()},
            "new_dispositions": [],
            "event_grades": [{"id": event_id, "grade": "MISSED",
                              "witnesses": [], "reason": "No delivered diagnosis covers it."}
                             for event_id in score.POSITIVES],
        }
        return bound, review

    def test_baseline_only_denominator_is_complete(self):
        bound, review = self.packet()
        result = score.compute(bound, review)
        self.assertEqual(result["pages"], 4)
        self.assertEqual(result["full"], 0)
        self.assertEqual(result["positive_events"], 8)
        self.assertEqual(result["accepted_delivered"], 9)
        self.assertEqual(result["delivered"], 11)
        self.assertFalse(result["directional_pass"])
        self.assertFalse(result["qualified"])

    def test_missing_page_is_invalid(self):
        bound, review = self.packet()
        del bound["page-010"]
        with self.assertRaisesRegex(ValueError, "incomplete bound page set"):
            score.compute(bound, review)

    def test_changed_response_hash_is_invalid(self):
        bound, review = self.packet()
        review["bound_sha256"]["page-002"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "bound response hash changed"):
            score.compute(bound, review)

    def test_duplicate_event_grade_is_invalid(self):
        bound, review = self.packet()
        review["event_grades"][1]["id"] = review["event_grades"][0]["id"]
        with self.assertRaisesRegex(ValueError, "missing or duplicate event grade"):
            score.compute(bound, review)

    def test_nonexistent_event_witness_is_invalid(self):
        bound, review = self.packet()
        review["event_grades"][0]["grade"] = "FULL"
        review["event_grades"][0]["witnesses"] = ["not-delivered"]
        with self.assertRaisesRegex(ValueError, "not delivered"):
            score.compute(bound, review)

    def test_lost_accepted_baseline_is_reported(self):
        bound, review = self.packet()
        accepted = next(finding_id for finding_id, decision in score.DISPOSITIONS.items()
                        if decision == "accepted")
        for page in bound.values():
            for item in page["reviews"]:
                if item["id"] == accepted:
                    item["decision"] = "withhold"
        review["bound_sha256"] = {name: run.digest(run.encoded(page))
                                   for name, page in bound.items()}
        result = score.compute(bound, review)
        self.assertIn(accepted, result["withheld_accepted_baseline"])
        self.assertFalse(result["directional_pass"])

    def test_retained_control_overlap_is_reported(self):
        bound, review = self.packet()
        control = score.CONTROLS[0]
        name = run.page_name(control["page"])
        candidate = {"id": name + "/new-01", "start_byte": control["start_byte"],
                     "end_byte": control["end_byte"]}
        bound[name]["new_findings"].append(candidate)
        review["bound_sha256"] = {key: run.digest(run.encoded(page))
                                   for key, page in bound.items()}
        review["new_dispositions"] = [{"id": candidate["id"], "decision": "accepted",
                                       "reason": "Synthetic test disposition."}]
        result = score.compute(bound, review)
        self.assertTrue(result["retain_control_overlaps"])
        self.assertFalse(result["directional_pass"])

    def test_missing_new_disposition_is_invalid(self):
        bound, review = self.packet()
        page = bound["page-008"]
        page["new_findings"].append({"id": "page-008/new-01",
                                     "start_byte": 1051, "end_byte": 1067})
        review["bound_sha256"] = {name: run.digest(run.encoded(item))
                                   for name, item in bound.items()}
        with self.assertRaisesRegex(ValueError, "new disposition coverage"):
            score.compute(bound, review)


if __name__ == "__main__":
    unittest.main()
