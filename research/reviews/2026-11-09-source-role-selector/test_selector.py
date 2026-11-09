"""Source-only development controls and malformed-span rejection."""

import gzip
import json
from pathlib import Path
import tempfile
import unittest

from admit import load_pinned, target_spans
from score import measures
from selector import classify_spans, classify_target


ROOT = Path(__file__).resolve().parent
PRIOR = ROOT.parent / "2026-11-08-source-role-rescue"
REVIEW = json.loads((PRIOR / "review.json").read_text(encoding="utf-8"))


def source_bytes(source_id):
    path = PRIOR / REVIEW["sources"][source_id]["path"]
    raw = path.read_bytes()
    return gzip.decompress(raw) if path.suffix == ".gz" else raw


class SourceRoleTest(unittest.TestCase):
    def test_source_only_packet(self):
        expected = {"positive": "admit", "control": "withhold", "unresolved": "abstain"}
        for case in REVIEW["cases"]:
            with self.subTest(case=case["id"]):
                spans = [(start, end) for start, end, _ in case["spans"]]
                actual = classify_spans(source_bytes(case["source"]), spans)
                self.assertEqual(actual.action, expected[case["role"]], actual.reason)

    def test_rejects_out_of_source_span(self):
        with self.assertRaisesRegex(ValueError, "outside"):
            classify_target(b"source", 0, 999)

    def test_rejects_empty_proposal(self):
        with self.assertRaisesRegex(ValueError, "no targets"):
            classify_spans(b"source", [])

    def test_rejects_changed_pinned_input(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "input.json"
            path.write_text("{}", encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "changed input"):
                load_pinned(path, "0" * 64)

    def test_rejects_changed_target_quote(self):
        with self.assertRaisesRegex(ValueError, "target quote mismatch"):
            target_spans(b"source", {"targets": [{"span": {"start": 0, "end": 6}, "quote": "other"}]})

    def test_complete_witness_accounting(self):
        evaluation = {"complete_witnesses": {"whole/e1": {"full_groups": [["a", "b"]]}}}
        metadata = {alias: {"page": "p", "stream": "s", "finding_id": alias} for alias in ("a", "b")}
        originals = {"paths": {"p": "whole/p.md"}}
        reviews = {"s": {"judgments": {
            "a": {"status": "accepted", "suggestion_safety": {"status": "uncertain"}},
            "b": {"status": "uncertain"},
        }}}
        partial = measures({"a"}, evaluation, metadata, originals, reviews)["whole"]
        self.assertEqual((partial["full"], partial["accepted"], partial["delivered"]), (0, 1, 1))
        full = measures({"a", "b"}, evaluation, metadata, originals, reviews)["whole"]
        self.assertEqual((full["full"], full["accepted"], full["delivered"]), (1, 1, 2))
        self.assertEqual(full["unresolved_archived_edits"], 1)


if __name__ == "__main__":
    unittest.main()
