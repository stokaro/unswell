"""Tests for exact baseline evidence and span accounting."""

import json
import tempfile
import unittest
from pathlib import Path

from audit_baseline import audit, overlap


ROOT = Path(__file__).resolve().parent
BASELINE = ROOT.parents[2] / "artifacts/research/2026-11-10-atomic-claim-pilot/baseline"


class AuditBaselineTests(unittest.TestCase):
    def test_half_open_spans(self):
        self.assertFalse(overlap({"start": 1, "end": 3}, {"start": 3, "end": 5}))
        self.assertTrue(overlap({"start": 1, "end": 4}, {"start": 3, "end": 5}))

    def test_frozen_result(self):
        if not BASELINE.exists():
            self.skipTest("local raw baseline reports are absent")
        result = audit(ROOT, BASELINE)
        self.assertEqual((result["event_primary_or_related_overlap"], result["event_count"]), (1, 15))
        self.assertEqual((result["control_primary_or_related_overlap"], result["control_count"]), (5, 21))
        self.assertEqual(result, json.loads((ROOT / "baseline-audit.json").read_text()))

    def test_changed_report_rejected(self):
        if not BASELINE.exists():
            self.skipTest("local raw baseline reports are absent")
        with tempfile.TemporaryDirectory() as directory:
            altered = Path(directory)
            for original in BASELINE.iterdir():
                (altered / original.name).symlink_to(original)
            victim = altered / "page-001.json"
            victim.unlink()
            victim.write_bytes((BASELINE / victim.name).read_bytes() + b" ")
            with self.assertRaisesRegex(ValueError, "baseline report changed"):
                audit(ROOT, altered)


if __name__ == "__main__":
    unittest.main()
