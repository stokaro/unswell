"""Negative checks for the research packet's identity and span verifier."""

import copy
import json
import unittest
from pathlib import Path

from verify import verify


ROOT = Path(__file__).resolve().parent
REVIEW = json.loads((ROOT / "review.json").read_text(encoding="utf-8"))


class VerifyPacketTest(unittest.TestCase):
    def test_exact_packet(self):
        self.assertEqual(
            verify(ROOT, copy.deepcopy(REVIEW)),
            {"positive": 3, "control": 7, "unresolved": 1},
        )

    def test_rejects_changed_source_digest(self):
        changed = copy.deepcopy(REVIEW)
        changed["sources"]["caddy"]["sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source digest: caddy"):
            verify(ROOT, changed)

    def test_rejects_changed_source_bytes_at_span(self):
        changed = copy.deepcopy(REVIEW)
        changed["cases"][0]["spans"][0][0] += 1
        with self.assertRaisesRegex(ValueError, "changed quote: animal-intro-kickoff"):
            verify(ROOT, changed)

    def test_rejects_case_reclassification(self):
        changed = copy.deepcopy(REVIEW)
        changed["cases"][0]["role"] = "control"
        with self.assertRaisesRegex(ValueError, "changed case counts"):
            verify(ROOT, changed)

    def test_rejects_path_escape(self):
        changed = copy.deepcopy(REVIEW)
        changed["sources"]["caddy"]["path"] = "sources/caddy/../README.md"
        with self.assertRaisesRegex(ValueError, "unsafe source path"):
            verify(ROOT, changed)


if __name__ == "__main__":
    unittest.main()
