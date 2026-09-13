"""Regression tests for the offline review comparator; no model or network."""

import copy
import hashlib
import importlib.util
from pathlib import Path
import tempfile
import unittest
import sys

sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("review_compare", Path(__file__).with_name("compare.py"))
compare = importlib.util.module_from_spec(spec)
spec.loader.exec_module(compare)


def finding(start=0, end=10, text="same text", rule="syntax.long-sentence"):
    return {"rule_id": rule, "rule_version": "1", "fingerprint": "old", "id": "old",
            "primary": {"path": "guide.md", "span": {"start": start, "end": end}, "snippet": text},
            "related": [], "evidence": {"kind": "heuristic", "activation": 1000,
                                      "metrics": [{"name": "words", "value": 40}],
                                      "occurrences": [{"block_id": 0, "spans": [{"start": start, "end": end}]}]}}


class AlignmentTest(unittest.TestCase):
    def test_fingerprint_version_and_local_ids_do_not_add_a_finding(self):
        old = finding()
        new = copy.deepcopy(old)
        new.update({"fingerprint": "new", "id": "new", "rule_version": "2"})
        new["evidence"]["occurrences"][0]["block_id"] = 3
        row, = compare.align([old], [new])
        self.assertEqual(row["status"], "retained")
        self.assertTrue(row["identity_changed"])
        self.assertTrue(row["mapping_changed"])

    def test_source_mapping_change_is_retained(self):
        row, = compare.align([finding()], [finding(2, 15)])
        self.assertEqual(row["status"], "retained")
        self.assertEqual(row["match"], "unique-excerpt")

    def test_metric_change_is_material(self):
        new = finding()
        new["evidence"]["metrics"][0]["value"] = 42
        row, = compare.align([finding()], [new])
        self.assertEqual(row["status"], "changed")

    def test_unique_overlap_is_one_change(self):
        new = finding(2, 15, "longer source")
        new["evidence"]["activation"] = 500
        row, = compare.align([finding()], [new])
        self.assertEqual((row["status"], row["match"]), ("changed", "unique-overlap"))

    def test_split_and_merge_stay_unresolved(self):
        old = [finding(0, 30, "whole")]
        new = [finding(0, 10, "first"), finding(15, 30, "second")]
        self.assertCountEqual([r["status"] for r in compare.align(old, new)],
                              ["unresolved-before", "unresolved-after", "unresolved-after"])
        self.assertCountEqual([r["status"] for r in compare.align(new, old)],
                              ["unresolved-before", "unresolved-before", "unresolved-after"])

    def test_same_words_in_another_file_or_rule_do_not_match(self):
        new = finding(rule="syntax.noun-stack")
        self.assertCountEqual([r["status"] for r in compare.align([finding()], [new])], ["removed", "added"])
        new = finding()
        new["primary"]["path"] = "other.md"
        self.assertCountEqual([r["status"] for r in compare.align([finding()], [new])], ["removed", "added"])

    def test_duplicates_are_not_arbitrarily_paired(self):
        rows = compare.align([finding(), finding()], [finding(), finding()])
        self.assertCountEqual([r["status"] for r in rows], ["unresolved-before"] * 2 + ["unresolved-after"] * 2)

    def test_nonfinite_json_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "bad.json"
            path.write_text('{"value": NaN}')
            with self.assertRaisesRegex(ValueError, "Nonfinite"):
                compare.read(path)

    def test_source_integrity_and_incomplete_runs(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = b"same text\n"
            sha = hashlib.sha256(source).hexdigest()
            manifest = {"sources": [{"path": "guide.md", "sha256": sha, "bytes": len(source)}]}
            report = {"schema_version": "1.0.0-alpha.1", "status": "incomplete",
                      "documents": [{"name": "guide.md", "source": source.decode(), "source_hash": sha,
                                     "bytes": len(source), "prose_words": 2, "excluded": []}],
                      "findings": [], "errors": [{"path": "other.go", "message": "parse failed"}],
                      "abstentions": [{"path": "guide.md", "reason": "budget_exhausted"}],
                      "gate": {"passed": False}, "manifest": {"rules": [], "config_sources": [{"sha256": "policy"}]}}
            compare.write(root / "result.json", report)
            compare.write(root / "invocation.json", {"exit_code": 2, "binary_sha256": "binary"})
            _, summary = compare.load_report(root / "result.json", manifest)
            self.assertEqual((summary["status"], summary["exit_code"]), ("incomplete", 2))
            self.assertEqual(summary["errors"], report["errors"])
            self.assertEqual(summary["abstentions"], report["abstentions"])
            manifest["sources"][0]["sha256"] = "wrong"
            with self.assertRaisesRegex(ValueError, "Source hash mismatch"):
                compare.load_report(root / "result.json", manifest)


if __name__ == "__main__":
    unittest.main()
