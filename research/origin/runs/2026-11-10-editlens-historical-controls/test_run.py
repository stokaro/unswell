"""Check historical control identities and fixed threshold accounting."""

from __future__ import annotations

import hashlib
import importlib.util
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("historical_run", HERE / "run.py")
assert SPEC is not None and SPEC.loader is not None
run = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(run)


def fixture() -> tuple[list[dict], list[dict]]:
    docs = []
    windows = []
    for index in range(462):
        source_id = f"source-{index:03d}"
        count = 9 if index < 39 else 8
        docs.append({"source_id": source_id, "windows_100": count,
                     "repository": f"repo-{index % 153}",
                     "language": ("C#", "C++", "Java", "Python")[index % 4]})
        for window_index in range(count):
            windows.append({"source_id": source_id, "index": window_index,
                            "score": 0.1})
    return docs, windows


class HistoricalRunTests(unittest.TestCase):
    def test_blob_identity(self) -> None:
        raw = b"source text\n"
        self.assertEqual(run.blob_sha(raw),
                         hashlib.sha1(b"blob 12\0" + raw).hexdigest())

    def test_source_identity(self) -> None:
        row = {"repository": "Owner/repo", "commit": "abc", "path": "README.md"}
        self.assertEqual(run.source_id(row),
                         hashlib.sha256(b"Owner/repo\0abc\0README.md").hexdigest())

    def test_zero_flag_gate_and_changed_score(self) -> None:
        docs, windows = fixture()
        result = run.summarize(docs, windows, "selection")
        self.assertEqual((result["documents"], result["repositories"],
                          result["windows"], result["flagged"]), (462, 153, 3735, 0))
        self.assertTrue(result["control_transfer_pass"])
        docs, windows = fixture()
        windows[0]["score"] = 0.9
        result = run.summarize(docs, windows, "selection")
        self.assertEqual(result["flagged"], 0)
        for row in windows[:9]:
            row["score"] = 0.9
        docs, _ = fixture()
        result = run.summarize(docs, windows, "selection")
        self.assertEqual(result["flagged"], 1)
        self.assertFalse(result["control_transfer_pass"])

    def test_missing_window_rejected(self) -> None:
        docs, windows = fixture()
        windows.pop()
        with self.assertRaisesRegex(ValueError, "Missing or repeated source window"):
            run.summarize(docs, windows, "selection")


if __name__ == "__main__":
    unittest.main()
