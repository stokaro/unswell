"""Negative tests for the independent raw-result verifier."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import verify_result


class ResultVerifierTests(unittest.TestCase):
    def copied(self, destination):
        base = Path(destination) / "copy"
        base.mkdir()
        shutil.copytree(verify_result.ROOT / "packet", base / "packet")
        shutil.copytree(verify_result.ROOT / "run", base / "run")
        for name in ("review.json", "result.json"):
            shutil.copy2(verify_result.ROOT / name, base / name)
        return base

    def test_real_result_verifies(self):
        self.assertEqual(verify_result.verify()["correct"], 15)

    def test_tool_event_rejected(self):
        with tempfile.TemporaryDirectory() as temp:
            base = self.copied(temp)
            first = sorted((base / "run").glob("*/events.jsonl"))[0]
            events = first.read_text().splitlines()
            events.insert(2, json.dumps({"type": "item.completed",
                                          "item": {"type": "command_execution", "command": "true"}}))
            first.write_text("\n".join(events) + "\n")
            with self.assertRaisesRegex(ValueError, "tool or error item"):
                verify_result.verify(base)

    def test_wrong_result_count_rejected(self):
        with tempfile.TemporaryDirectory() as temp:
            base = self.copied(temp)
            path = base / "result.json"
            result = json.loads(path.read_text())
            result["correct"] = 16
            path.write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "result content changed"):
                verify_result.verify(base)

    def test_unsafe_edit_cannot_keep_advance(self):
        with tempfile.TemporaryDirectory() as temp:
            base = self.copied(temp)
            path = base / "review.json"
            review = json.loads(path.read_text())
            review["edits"][0]["safety"] = "unsafe"
            path.write_text(json.dumps(review))
            with self.assertRaisesRegex(ValueError, "result content changed|result gate changed"):
                verify_result.verify(base)

    def test_packet_tamper_rejected(self):
        with tempfile.TemporaryDirectory() as temp:
            base = self.copied(temp)
            first = sorted((base / "packet" / "requests").glob("*.txt"))[0]
            first.write_bytes(first.read_bytes() + b"\n")
            with self.assertRaisesRegex(ValueError, "packet freeze changed"):
                verify_result.verify(base)


if __name__ == "__main__":
    unittest.main()
