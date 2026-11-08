"""Negative checks for the frozen paired editorial request packet."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import run


class PacketTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.original_packet = run.PACKET
        self.addCleanup(setattr, run, "PACKET", self.original_packet)
        run.PACKET = Path(self.temp.name) / "packet"
        shutil.copytree(self.original_packet, run.PACKET)

    def test_frozen_packet_verifies(self):
        self.assertEqual(run.verify()["calls_max"], 24)

    def test_changed_request_is_rejected(self):
        request = next((run.PACKET / "requests").glob("*.txt"))
        request.write_bytes(request.read_bytes() + b"\n")
        with self.assertRaisesRegex(ValueError, "frozen packet"):
            run.verify()

    def test_changed_old_grade_is_rejected(self):
        path = run.PACKET / "key.json"
        grades = json.loads(path.read_text())
        grades[0]["old_status"] = "uncertain" if grades[0]["old_status"] == "accepted" else "accepted"
        path.write_bytes(run.encoded(grades))
        with self.assertRaisesRegex(ValueError, "balanced case inventory"):
            run.verify()

    def test_broken_source_offset_is_rejected(self):
        path = run.PACKET / "sample.json"
        sample = json.loads(path.read_text())
        sample["items"][0]["target_start_byte"] += 1
        path.write_bytes(run.encoded(sample))
        with self.assertRaisesRegex(ValueError, "source binding"):
            run.verify()

    def test_current_grade_does_not_enter_requests(self):
        for payload, _, arm, grade, _ in run.entries():
            request = run.request(payload, arm).decode("utf-8")
            self.assertNotIn(grade["candidate_id"], request)
            self.assertNotIn(grade["old_rationale"], request)
            if arm == "control":
                self.assertEqual(run.request(payload, arm), run.base().request(payload))


if __name__ == "__main__":
    unittest.main()
