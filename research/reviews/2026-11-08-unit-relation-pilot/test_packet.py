"""Reject source, inventory, and verdict violations before any model call."""

import copy
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest import mock

import packet


class PacketTest(unittest.TestCase):
    def test_source_byte_change_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory)
            shutil.copytree(packet.ROOT / "sources", target / "sources")
            for name in ("sources.json", "inventory.json", "input.json", "labels.json"):
                shutil.copyfile(packet.ROOT / name, target / name)
            path = target / "sources" / "mise" / "README.md"
            path.write_bytes(path.read_bytes() + b"\nchanged\n")
            with mock.patch.object(packet, "ROOT", target):
                with self.assertRaisesRegex(ValueError, "source identity changed"):
                    packet.validate_sources_and_labels()

    def test_missing_inventory_unit_is_rejected(self):
        bounds = packet.chunks("zellij")[0]
        page_id, raw, _ = packet.page("zellij")
        response = {"page_id": page_id, "source_sha256": packet.sha(raw),
                    "chunk_id": packet.chunk_id("zellij", bounds),
                    "units": [{"ordinal": ordinal, "function": "other",
                               "proposition": "", "support": "No claim.",
                               "relation": "not_claim"}
                              for ordinal in range(*bounds)]}
        packet.validate_inventory_chunk("zellij", bounds, response)
        response["units"].pop()
        with self.assertRaisesRegex(ValueError, "chunk identity"):
            packet.validate_inventory_chunk("zellij", bounds, response)

    def test_wrong_relation_is_rejected(self):
        bounds = packet.chunks("zellij")[0]
        page_id, raw, _ = packet.page("zellij")
        response = {"page_id": page_id, "source_sha256": packet.sha(raw),
                    "chunk_id": packet.chunk_id("zellij", bounds),
                    "units": [{"ordinal": ordinal, "function": "other",
                               "proposition": "", "support": "No claim.",
                               "relation": "not_claim"}
                              for ordinal in range(*bounds)]}
        response["units"][4]["relation"] = "maybe"
        with self.assertRaisesRegex(ValueError, "invalid inventory unit"):
            packet.validate_inventory_chunk("zellij", bounds, response)

    def test_verdict_requires_original_eligible_span_and_every_candidate(self):
        page_id, raw, record = packet.page("zellij")
        label = next(e for e in packet.read(packet.ROOT / "labels.json")["events"]
                     if e["id"] == "zellij-unique-features")
        rows = [{"ordinal": ordinal, "function": "other", "proposition": "",
                 "support": "No claim.", "relation": "not_claim"}
                for ordinal in range(len(record["units"]))]
        ordinal = label["overlapping_units"][0]["ordinal"]
        rows[ordinal]["relation"] = "narrower"
        merged_raw = packet.encoded({"page_id": page_id, "source_sha256": packet.sha(raw),
                                     "units": rows})
        finding = {"ordinal": ordinal, "decision": "diagnose",
                   "start": label["span"]["start"], "end": label["span"]["end"],
                   "quote": label["quote"], "diagnosis": "Unbounded comparison.",
                   "evidence": "Only two pane features are named.",
                   "source_function": "Promotion.", "deletion_test": "Keep the named features.",
                   "reason": ""}
        response = {"page_id": page_id, "source_sha256": packet.sha(raw),
                    "inventory_sha256": packet.sha(merged_raw), "verdicts": [finding]}
        bound = packet.validate_verdict("zellij", response, merged_raw)
        self.assertEqual(bound["findings"][0]["quote"], label["quote"])
        bad = copy.deepcopy(response)
        bad["verdicts"] = []
        with self.assertRaisesRegex(ValueError, "candidate count"):
            packet.validate_verdict("zellij", bad, merged_raw)
        bad = copy.deepcopy(response)
        bad["verdicts"][0]["start"] -= 1
        with self.assertRaisesRegex(ValueError, "unbound diagnosis"):
            packet.validate_verdict("zellij", bad, merged_raw)
        bad = copy.deepcopy(response)
        bad["verdicts"][0]["start"] = 0
        bad["verdicts"][0]["end"] = len(b"<h1")
        bad["verdicts"][0]["quote"] = raw[:len(b"<h1")].decode("utf-8")
        with self.assertRaisesRegex(ValueError, "protected or other"):
            packet.validate_verdict("zellij", bad, merged_raw)


if __name__ == "__main__":
    unittest.main()
