"""Negative mutation checks for the recorded paired pilot result."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import run
import verify_result


class RecordedResultTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        shutil.copytree(run.RUN, self.root / "run")
        shutil.copy2(verify_result.ROOT / "result.json", self.root / "result.json")
        shutil.copy2(verify_result.ROOT / "verification.json", self.root / "verification.json")
        self.previous_run = run.RUN
        self.previous_root = verify_result.ROOT
        run.RUN = self.root / "run"
        verify_result.ROOT = self.root

    def tearDown(self):
        run.RUN = self.previous_run
        verify_result.ROOT = self.previous_root
        self.temp.cleanup()

    def test_recorded_result_verifies(self):
        record = verify_result.verify()
        self.assertEqual(len(record["run_files_sha256"]), 33)

    def test_false_advance_is_rejected(self):
        path = self.root / "result.json"
        data = run.read(path)
        data["advance"] = True
        path.write_bytes(run.encoded(data))
        with self.assertRaisesRegex(ValueError, "scored result"):
            verify_result.verify()

    def test_changed_response_is_rejected(self):
        path = self.root / "run" / "bat" / "response.json"
        data = run.read(path)
        data["A_start"] += 1
        path.write_bytes(run.encoded(data))
        with self.assertRaisesRegex(ValueError, "trace|binding"):
            verify_result.verify()

    def test_tool_event_is_rejected(self):
        path = self.root / "run" / "curl" / "events.jsonl"
        with path.open("a", encoding="utf-8") as output:
            output.write(json.dumps({"type": "item.completed",
                                     "item": {"type": "command_execution"}}) + "\n")
        with self.assertRaisesRegex(ValueError, "tool use"):
            verify_result.verify()

    def test_changed_verification_record_is_rejected(self):
        path = self.root / "verification.json"
        data = run.read(path)
        data["result_sha256"] = "0" * 64
        path.write_bytes(run.encoded(data))
        with self.assertRaisesRegex(ValueError, "verification record"):
            verify_result.verify()


if __name__ == "__main__":
    unittest.main()
