"""Reject fabricated FULL credit and incomplete source reviews."""

import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import run
import score


class PilotScoreTest(unittest.TestCase):
    def reference_event(self):
        return run.read(run.ROOT / "reference.json")["events"][0]

    def candidate(self):
        event = self.reference_event()
        return {"candidate_id": "candidate-a", "page_id": event["page"],
                "category": event["category"], "span": event["span"]}

    def review_file(self, directory, findings):
        path = Path(directory) / "review.json"
        path.write_text(json.dumps({"format": "unswell-349-additive-admission-review-v1",
                                    "reviewer": "Codex assistant, complete-source review",
                                    "independent_human": False,
                                    "findings": findings}) + "\n")
        return path

    def test_wrong_category_cannot_receive_full_credit(self):
        candidate = self.candidate()
        candidate["category"] = "vague_claims"
        with tempfile.TemporaryDirectory() as directory:
            path = self.review_file(directory, [{"candidate_id": candidate["candidate_id"],
                "status": "accepted", "rationale": "Some source-specific reason",
                "credited_event_ids": [self.reference_event()["id"]],
                "retain_hit_ids": []}])
            with mock.patch.object(score, "collect", return_value=(
                {}, [candidate], [{"candidate_id": candidate["candidate_id"],
                                   "decision": "admit"}], True)):
                with self.assertRaisesRegex(ValueError, "unsupported FULL credit"):
                    score.score(path)

    def test_missing_admitted_review_rejected(self):
        candidate = self.candidate()
        with tempfile.TemporaryDirectory() as directory:
            path = self.review_file(directory, [])
            with mock.patch.object(score, "collect", return_value=(
                {}, [candidate], [{"candidate_id": candidate["candidate_id"],
                                   "decision": "admit"}], True)):
                with self.assertRaisesRegex(ValueError, "complete admitted review"):
                    score.score(path)

    def test_retain_control_hit_fails_pilot(self):
        control = run.read(run.ROOT / "reference.json")["retain_controls"][0]
        candidate = {"candidate_id": "candidate-b", "page_id": control["page"],
                     "category": "empty_framing", "span": control["span"]}
        with tempfile.TemporaryDirectory() as directory:
            path = self.review_file(directory, [{"candidate_id": candidate["candidate_id"],
                "status": "rejected", "rationale": "The tagline is useful here.",
                "credited_event_ids": [], "retain_hit_ids": [control["id"]]}])
            with mock.patch.object(score, "collect", return_value=(
                {}, [candidate], [{"candidate_id": candidate["candidate_id"],
                                   "decision": "admit"}], True)), \
                 mock.patch.object(score.run, "file_sha", return_value="hash"):
                result = score.score(path)
            self.assertFalse(result["pilot_passed"])
            self.assertEqual(result["retain_control_hits"], [control["id"]])


if __name__ == "__main__":
    unittest.main()
