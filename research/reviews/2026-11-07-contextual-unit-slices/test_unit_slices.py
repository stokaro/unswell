"""Negative checks for the frozen source-local raw-unit contract."""

import copy
import json
import sys
import tempfile
import unittest
from pathlib import Path

import unit_slices


sys.path.insert(0, str(unit_slices.ROOT.parent / "2026-09-26-contextual-evidence"))
import study  # noqa: E402 - import the original frozen target binder


class UnitSliceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.manifest, cls.packets = unit_slices.build()

    def test_all_frozen_pages_match_manifest(self):
        unit_slices.verify_manifest(self.manifest)
        self.assertEqual(len(self.packets), 36)
        self.assertEqual(sum(row["units"] for row in self.manifest["pages"]), 4549)

    def test_historical_markdown_boundary_is_visible(self):
        packet = json.loads(self.packets["page-019.json"])
        unit = next(row for row in packet["units"] if row["id"] == "u0038")
        self.assertTrue(unit["raw"].startswith("Deployment reports**"))
        self.assertNotIn("**Deployment reports**", unit["raw"])
        with self.assertRaisesRegex(ValueError, "not in the selected source unit"):
            study.bind_target(packet, {"unit": "u0038", "quote": "**Deployment reports**",
                                       "occurrence": 1})
        bound = study.bind_target(packet, {"unit": "u0038", "quote": "Deployment reports**",
                                           "occurrence": 1})
        source = packet["source"].encode("utf-8")
        self.assertEqual(source[bound["span"]["start"]:bound["span"]["end"]],
                         b"Deployment reports**")

    def test_non_ascii_repeated_quote_uses_original_bytes(self):
        source = "café café"
        packet = {"page": "synthetic", "source": source,
                  "source_sha256": unit_slices.sha256(source.encode("utf-8")),
                  "protected": [], "units": [{"id": "u0001", "span":
                                             {"start": 0, "end": len(source.encode("utf-8"))}}]}
        augmented = unit_slices.augment(packet)
        unit_slices.verify_augmented(augmented, packet)
        target = study.bind_target(augmented, {"unit": "u0001", "quote": "café",
                                               "occurrence": 2})
        self.assertEqual(target["span"], {"start": 6, "end": 11})

    def test_changed_raw_slice_is_rejected(self):
        original = json.loads(self.packets["page-019.json"])
        original_units = [{"id": row["id"], "span": row["span"]}
                          for row in original["units"]]
        prior = {**original, "units": original_units}
        changed = copy.deepcopy(original)
        changed["units"][0]["raw"] += "x"
        with self.assertRaisesRegex(ValueError, "Raw unit slice changed"):
            unit_slices.verify_augmented(changed, prior)

    def test_shifted_unit_span_is_rejected(self):
        original = json.loads(self.packets["page-019.json"])
        prior = {**original, "units": [{"id": row["id"], "span": row["span"]}
                                        for row in original["units"]]}
        changed = copy.deepcopy(original)
        changed["units"][0]["span"]["start"] += 1
        with self.assertRaisesRegex(ValueError, "Original unit identity changed"):
            unit_slices.verify_augmented(changed, prior)

    def test_protected_region_drift_is_rejected(self):
        original = json.loads(self.packets["page-019.json"])
        prior = {**original, "units": [{"id": row["id"], "span": row["span"]}
                                        for row in original["units"]]}
        changed = copy.deepcopy(original)
        self.assertTrue(changed["protected"])
        changed["protected"][0]["span"]["start"] += 1
        with self.assertRaisesRegex(ValueError, "Original packet field changed: protected"):
            unit_slices.verify_augmented(changed, prior)

    def test_split_utf8_character_is_rejected(self):
        source = "éa"
        packet = {"source": source,
                  "source_sha256": unit_slices.sha256(source.encode("utf-8")),
                  "units": [{"id": "u0001", "span": {"start": 1, "end": 3}}]}
        with self.assertRaisesRegex(ValueError, "splits a UTF-8 character"):
            unit_slices.augment(packet)

    def test_changed_manifest_is_rejected(self):
        changed = copy.deepcopy(self.manifest)
        changed["pages"][0]["packet_sha256"] = "0" * 64
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "manifest.json"
            path.write_bytes(unit_slices.encoded(changed))
            with self.assertRaisesRegex(ValueError, "manifest changed"):
                unit_slices.verify_manifest(self.manifest, path)

    def test_changed_archive_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "evidence.tar.gz"
            path.write_bytes(unit_slices.ARCHIVE.read_bytes() + b"x")
            with self.assertRaisesRegex(ValueError, "archive changed"):
                unit_slices.build(path)


if __name__ == "__main__":
    unittest.main()
