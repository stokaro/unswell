"""Prove the result verifier rejects changed inputs and unsupported claims."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import verify_result


class ResultVerifierTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name) / "study"
        shutil.copytree(verify_result.ROOT, self.base,
                        ignore=shutil.ignore_patterns("__pycache__"))

    def test_complete_negative_result_verifies(self):
        report = verify_result.verify(self.base)
        self.assertEqual(report["calls"], 24)
        self.assertFalse(report["advance"])

    def test_request_mutation_is_rejected(self):
        request = next((self.base / "packet/requests").glob("*.txt"))
        request.write_bytes(request.read_bytes() + b"\n")
        with self.assertRaisesRegex(ValueError, "frozen packet file set"):
            verify_result.verify(self.base)

    def test_response_mutation_is_rejected(self):
        item_id = json.loads((self.base / "run/run.json").read_text())["calls"][0]["item_id"]
        path = self.base / "run" / item_id / "response.json"
        response = json.loads(path.read_text())
        response["basis"] += " altered"
        path.write_text(json.dumps(response))
        with self.assertRaisesRegex(ValueError, "trace, response, or bound file"):
            verify_result.verify(self.base)

    def test_tool_event_is_rejected(self):
        item_id = json.loads((self.base / "run/run.json").read_text())["calls"][0]["item_id"]
        path = self.base / "run" / item_id / "events.jsonl"
        with path.open("a") as stream:
            stream.write(json.dumps({"type": "item.completed", "item": {"type": "tool_call"}}) + "\n")
        with self.assertRaisesRegex(ValueError, "tool or error item"):
            verify_result.verify(self.base)

    def test_changed_edit_review_is_rejected(self):
        path = self.base / "review.json"
        review = json.loads(path.read_text())
        review["edits"][0]["rationale"] += " altered"
        path.write_text(json.dumps(review))
        with self.assertRaisesRegex(ValueError, "scored result changed"):
            verify_result.verify(self.base)

    def test_claimed_advance_is_rejected(self):
        path = self.base / "result.json"
        result = json.loads(path.read_text())
        result["advance"] = True
        path.write_text(json.dumps(result))
        with self.assertRaisesRegex(ValueError, "scored result changed"):
            verify_result.verify(self.base)


if __name__ == "__main__":
    unittest.main()
