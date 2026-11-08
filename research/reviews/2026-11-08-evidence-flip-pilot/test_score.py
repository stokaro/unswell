"""Gate tests using synthetic traces, never a model response."""

import json
from pathlib import Path
import tempfile
import unittest
from unittest import mock

import run
import score


class ScoreTests(unittest.TestCase):
    def make_run(self, directory, flip=None, unsafe=False):
        destination = Path(directory) / "run"
        destination.mkdir()
        calls, admitted = [], []
        for payload, key in run.entries():
            decision = "admit" if key["applies"] else "reject"
            if flip == key["case_id"] + "/" + key["variant_id"]:
                decision = "reject" if decision == "admit" else "admit"
            response = {"item_id": payload["item_id"],
                        "source_sha256": payload["source_sha256"],
                        "target_start_byte": payload["target_start_byte"],
                        "target_end_byte": payload["target_end_byte"],
                        "decision": decision,
                        "replacement": "Revised." if decision == "admit" else "",
                        "basis": "The source context determines this diagnosis."}
            target = destination / payload["item_id"]
            target.mkdir()
            (target / "response.json").write_bytes(run.encoded(response))
            (target / "bound.json").write_bytes(run.encoded(response))
            events = [{"type": "thread.started"}, {"type": "turn.started"},
                      {"type": "item.completed", "item": {"type": "agent_message",
                         "text": json.dumps(response, sort_keys=True)}},
                      {"type": "turn.completed"}]
            (target / "events.jsonl").write_text(
                "".join(json.dumps(event) + "\n" for event in events))
            calls.append({"item_id": payload["item_id"], "status": "valid",
                          "request_sha256": run.file_sha(
                              run.PACKET / "requests" / (payload["item_id"] + ".txt")),
                          "bound_sha256": run.file_sha(target / "bound.json")})
            if decision == "admit":
                admitted.append({"item_id": payload["item_id"],
                                 "safety": "safe", "rationale": "Synthetic gate fixture."})
        journal = {"status": "complete", "complete": True,
                   "freeze_sha256": run.file_sha(run.PACKET / "freeze.json"),
                   "model": run.MODEL, "effort": run.EFFORT,
                   "calls": calls, "unattempted": []}
        (destination / "run.json").write_bytes(run.encoded(journal))
        if unsafe:
            admitted[0]["safety"] = "unsafe"
        review = Path(directory) / "review.json"
        review.write_bytes(run.encoded({"format": "unswell-349-evidence-flip-safety-review-v1",
                                        "edits": admitted}))
        return destination, review

    def test_full_success_advances(self):
        with tempfile.TemporaryDirectory() as temp:
            destination, review = self.make_run(temp)
            with mock.patch.object(run, "RUN", destination):
                result = score.evaluate(review)
            self.assertTrue(result["advance"])
            self.assertEqual(result["correct"], 16)

    def test_support_flip_error_rejects(self):
        with tempfile.TemporaryDirectory() as temp:
            destination, review = self.make_run(temp, flip="table-announcement/heading-removed")
            with mock.patch.object(run, "RUN", destination):
                result = score.evaluate(review)
            self.assertFalse(result["advance"])
            self.assertEqual(result["correct"], 15)
            self.assertFalse(result["pair_results"][0]["both_correct"])

    def test_unsafe_advice_rejects(self):
        with tempfile.TemporaryDirectory() as temp:
            destination, review = self.make_run(temp, unsafe=True)
            with mock.patch.object(run, "RUN", destination):
                result = score.evaluate(review)
            self.assertFalse(result["advance"])
            self.assertFalse(result["safety_clear"])

    def test_missing_safety_review_rejects(self):
        with tempfile.TemporaryDirectory() as temp:
            destination, review = self.make_run(temp)
            review.unlink()
            with mock.patch.object(run, "RUN", destination):
                result = score.evaluate(review)
            self.assertFalse(result["advance"])
            self.assertFalse(result["safety_clear"])


if __name__ == "__main__":
    unittest.main()
