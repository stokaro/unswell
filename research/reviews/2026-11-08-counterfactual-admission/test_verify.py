"""Black-box contract tests for the frozen counterfactual controls."""

import copy
import json
import sys
import unittest
from pathlib import Path

DIRECTORY = Path(__file__).resolve().parent
sys.path.insert(0, str(DIRECTORY))
import verify  # noqa: E402


class PacketContractTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.packet = json.loads((DIRECTORY / "cases.json").read_text(encoding="utf-8"))

    def test_packet_has_sixteen_source_bound_variants(self):
        receipts = verify.validate(self.packet)
        self.assertEqual(len(receipts), 16)
        self.assertEqual(sum(row["applies"] for row in receipts), 8)
        self.assertEqual(len({row["source_sha256"] for row in receipts}), 16)

    def test_missing_target_rejected(self):
        packet = copy.deepcopy(self.packet)
        packet["cases"][0]["variants"][0]["source"] = "# Empty\n"
        with self.assertRaisesRegex(ValueError, "target must occur exactly once"):
            verify.validate(packet)

    def test_repeated_target_rejected(self):
        packet = copy.deepcopy(self.packet)
        variant = packet["cases"][0]["variants"][0]
        variant["source"] += variant["source"].split("\n")[2] + "\n"
        with self.assertRaisesRegex(ValueError, "target must occur exactly once"):
            verify.validate(packet)

    def test_label_change_rejected(self):
        packet = copy.deepcopy(self.packet)
        packet["cases"][0]["variants"][0]["applies"] = False
        with self.assertRaisesRegex(ValueError, "expected decision changed"):
            verify.validate(packet)

    def test_role_change_rejected(self):
        packet = copy.deepcopy(self.packet)
        packet["cases"][0]["variants"][0]["role"] = "support-added"
        with self.assertRaisesRegex(ValueError, "role or rationale changed"):
            verify.validate(packet)

    def test_duplicate_source_rejected(self):
        packet = copy.deepcopy(self.packet)
        packet["cases"][0]["variants"][1]["source"] = packet["cases"][0]["variants"][0]["source"]
        with self.assertRaisesRegex(ValueError, "duplicate source"):
            verify.validate(packet)

    def test_review_provenance_rejected(self):
        packet = copy.deepcopy(self.packet)
        packet["independent_human"] = True
        with self.assertRaisesRegex(ValueError, "review provenance is missing"):
            verify.validate(packet)

    def test_original_event_label_use_rejected(self):
        packet = copy.deepcopy(self.packet)
        packet["original_349_events_used"] = True
        with self.assertRaisesRegex(ValueError, "original event labels"):
            verify.validate(packet)


if __name__ == "__main__":
    unittest.main()
