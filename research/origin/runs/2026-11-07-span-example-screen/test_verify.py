"""Negative checks for the frozen span-neighbor research ledger."""

import gzip
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
import unittest

from verify import HERE, verify


FEATURE = HERE.parent / "2026-11-07-richness-screen" / "run.py"


class ScreenVerifierTest(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        for name in ("PROTOCOL.md", "run.py", "freeze.json", "scores.json.gz", "result.json"):
            shutil.copy2(HERE / name, self.root / name)

    def check(self):
        return verify(self.root, feature_source=FEATURE)

    def change_scores(self, update):
        path = self.root / "scores.json.gz"
        rows = json.loads(gzip.decompress(path.read_bytes()))
        update(rows)
        content = (json.dumps(rows, indent=2, sort_keys=True) + "\n").encode()
        path.write_bytes(gzip.compress(content, compresslevel=9, mtime=0))
        result_path = self.root / "result.json"
        result = json.loads(result_path.read_text())
        result["scores_sha256"] = hashlib.sha256(content).hexdigest()
        result_path.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")

    def test_saved_screen(self):
        result = self.check()
        self.assertEqual(result["rows"], 362)
        self.assertFalse(result["feasibility_pass"])

    def test_rejects_same_author_neighbor_after_hash_refresh(self):
        def change(rows):
            row = next(r for r in rows if r.get("assisted") and "neighbors" in r)
            row["neighbors"][0]["assisted"][0][0] = row["row_idx"]
        self.change_scores(change)
        with self.assertRaisesRegex(ValueError, "author isolation"):
            self.check()

    def test_rejects_changed_span_margin_after_hash_refresh(self):
        def change(rows):
            row = next(r for r in rows if "score" in r)
            row["span_scores"][0] += 0.01
            row["score"] = sum(row["span_scores"]) / 4
        self.change_scores(change)
        with self.assertRaisesRegex(ValueError, "Span score disagrees"):
            self.check()

    def test_rejects_dropped_row_after_hash_refresh(self):
        self.change_scores(lambda rows: rows.pop())
        with self.assertRaisesRegex(ValueError, "Incomplete or reordered"):
            self.check()

    def test_rejects_false_pass_claim(self):
        path = self.root / "result.json"
        result = json.loads(path.read_text())
        result["feasibility_pass"] = True
        path.write_text(json.dumps(result, indent=2, sort_keys=True) + "\n")
        with self.assertRaisesRegex(ValueError, "feasibility decision"):
            self.check()


if __name__ == "__main__":
    unittest.main()
