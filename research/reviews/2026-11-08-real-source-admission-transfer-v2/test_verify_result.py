"""Negative tests for independent real-source transfer accounting."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import verify_result


class ResultVerifierTests(unittest.TestCase):
    def copied_result(self):
        temporary = tempfile.TemporaryDirectory()
        self.addCleanup(temporary.cleanup)
        parent = Path(temporary.name)
        current = parent / "v2"
        previous = parent / "v1"
        shutil.copytree(verify_result.ROOT, current)
        shutil.copytree(verify_result.V1, previous)
        return current, previous

    def test_original_result(self):
        summary = verify_result.verify()
        self.assertEqual(summary["calls"], 24)
        self.assertFalse(summary["advance"])

    def test_changed_request_rejected(self):
        current, previous = self.copied_result()
        request = next((current / "packet/requests").glob("*.txt"))
        request.write_bytes(request.read_bytes() + b"changed\n")
        with self.assertRaisesRegex(ValueError, "packet file set or freeze"):
            verify_result.verify(current, previous)

    def test_changed_response_rejected(self):
        current, previous = self.copied_result()
        response = next((current / "run").glob("*/response.json"))
        data = json.loads(response.read_text())
        data["source_sha256"] = "0" * 64
        response.write_text(json.dumps(data))
        with self.assertRaisesRegex(ValueError, "response or trace"):
            verify_result.verify(current, previous)

    def test_changed_edit_review_rejected(self):
        current, previous = self.copied_result()
        path = current / "review.json"
        review = json.loads(path.read_text())
        review["edits"][0]["safety"] = "unresolved"
        path.write_text(json.dumps(review))
        with self.assertRaisesRegex(ValueError, "scored result changed"):
            verify_result.verify(current, previous)

    def test_changed_result_gate_rejected(self):
        current, previous = self.copied_result()
        path = current / "result.json"
        result = json.loads(path.read_text())
        result["advance"] = True
        path.write_text(json.dumps(result))
        with self.assertRaisesRegex(ValueError, "scored result changed"):
            verify_result.verify(current, previous)

    def test_v1_stop_erasure_rejected(self):
        current, previous = self.copied_result()
        stderr = next((previous / "run").glob("*/stderr.txt"))
        stderr.write_text("")
        with self.assertRaisesRegex(ValueError, "v1 terminal stop"):
            verify_result.verify(current, previous)


if __name__ == "__main__":
    unittest.main()
