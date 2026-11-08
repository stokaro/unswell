"""Positive and negative checks for the frozen real-source request packet."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest.mock import patch

import run


class FrozenPacketTests(unittest.TestCase):
    def test_original_packet_and_blinding(self):
        protocol = run.verify()
        self.assertEqual(protocol["calls_max"], 24)
        self.assertEqual(len(run.entries()), 24)
        for payload, _, _ in run.entries():
            request = run.request(payload)
            self.assertNotIn(b'"old_status"', request)
            self.assertNotIn(b'"candidate_id"', request)
            self.assertNotIn(b'"group"', request)

    def copied_packet(self):
        temporary = tempfile.TemporaryDirectory()
        target = Path(temporary.name) / "packet"
        shutil.copytree(run.PACKET, target)
        return temporary, target

    def test_changed_request_rejected(self):
        temporary, target = self.copied_packet()
        self.addCleanup(temporary.cleanup)
        request = next((target / "requests").glob("*.txt"))
        request.write_bytes(request.read_bytes() + b"extra\n")
        with patch.object(run, "PACKET", target):
            with self.assertRaisesRegex(ValueError, "request set"):
                run.verify()

    def test_changed_old_grade_rejected(self):
        temporary, target = self.copied_packet()
        self.addCleanup(temporary.cleanup)
        key_path = target / "key.json"
        key = json.loads(key_path.read_text())
        key[0]["old_status"] = ("uncertain" if key[0]["old_status"] == "accepted"
                                else "accepted")
        key_path.write_text(json.dumps(key))
        with patch.object(run, "PACKET", target):
            with self.assertRaisesRegex(ValueError, "stratified sample"):
                run.verify()

    def test_changed_source_offset_rejected(self):
        temporary, target = self.copied_packet()
        self.addCleanup(temporary.cleanup)
        sample_path = target / "sample.json"
        sample = json.loads(sample_path.read_text())
        sample["items"][0]["target_start_byte"] += 1
        sample_path.write_text(json.dumps(sample))
        with patch.object(run, "PACKET", target):
            with self.assertRaisesRegex(ValueError, "source binding"):
                run.verify()

    def test_changed_schema_rejected(self):
        temporary, target = self.copied_packet()
        self.addCleanup(temporary.cleanup)
        schema_path = target / "schema.json"
        schema = json.loads(schema_path.read_text())
        schema["properties"]["decision"]["enum"].append("maybe")
        schema_path.write_text(json.dumps(schema))
        with patch.object(run, "PACKET", target):
            with self.assertRaisesRegex(ValueError, "response schema"):
                run.verify()


if __name__ == "__main__":
    unittest.main()
