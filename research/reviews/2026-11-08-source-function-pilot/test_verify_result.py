"""Negative controls for the #349 source-function result verifier."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import verify_result

ROOT = Path(__file__).resolve().parent


class ResultVerifierTest(unittest.TestCase):
    def test_frozen_result(self):
        self.assertFalse(verify_result.verify(ROOT)["advance"])

    def test_packet_tampering(self):
        with tempfile.TemporaryDirectory() as folder:
            target = Path(folder) / "pilot"
            shutil.copytree(ROOT, target, ignore=shutil.ignore_patterns("__pycache__", "baseline", "empty_root"))
            request = target / "packet" / "requests" / "caddy.txt"
            request.write_bytes(request.read_bytes() + b"tamper")
            with self.assertRaises(ValueError):
                verify_result.verify(target)

    def test_review_omission(self):
        with tempfile.TemporaryDirectory() as folder:
            target = Path(folder) / "pilot"
            shutil.copytree(ROOT, target, ignore=shutil.ignore_patterns("__pycache__", "baseline", "empty_root"))
            path = target / "review.json"
            review = json.loads(path.read_text())
            review["findings"].pop()
            path.write_text(json.dumps(review))
            with self.assertRaises(ValueError):
                verify_result.verify(target)

    def test_false_success_result(self):
        with tempfile.TemporaryDirectory() as folder:
            target = Path(folder) / "pilot"
            shutil.copytree(ROOT, target, ignore=shutil.ignore_patterns("__pycache__", "baseline", "empty_root"))
            path = target / "result.json"
            result = json.loads(path.read_text())
            result["advance"] = True
            path.write_text(json.dumps(result))
            with self.assertRaises(ValueError):
                verify_result.verify(target)

    def test_tool_event(self):
        with tempfile.TemporaryDirectory() as folder:
            target = Path(folder) / "pilot"
            shutil.copytree(ROOT, target, ignore=shutil.ignore_patterns("__pycache__", "baseline", "empty_root"))
            path = target / "run" / "caddy" / "events.jsonl"
            path.write_text(path.read_text() +
                            json.dumps({"type": "item.completed", "item": {"type": "command_execution"}}) + "\n")
            with self.assertRaises(ValueError):
                verify_result.verify(target)


if __name__ == "__main__":
    unittest.main()
