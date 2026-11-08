"""Negative controls for baseline artifact and semantic-review accounting."""

import copy
import json
import sys
import unittest
from pathlib import Path

DIRECTORY = Path(__file__).resolve().parent
sys.path.insert(0, str(DIRECTORY))
import verify_result  # noqa: E402


def load():
    summary = json.loads((DIRECTORY / "baseline.json").read_text())
    review = json.loads((DIRECTORY / "review.json").read_text())
    reports = verify_result.read_reports(summary)
    return summary, review, reports


class ResultContractTests(unittest.TestCase):
    def test_complete_result(self):
        self.assertEqual(verify_result.verify_files()["matched_fixed_diagnoses"], 0)

    def test_changed_source_receipt_rejected(self):
        summary, review, reports = load()
        summary["variants"][0]["source_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source or label receipt changed"):
            verify_result.validate_result(summary, review, reports)

    def test_missing_variant_rejected(self):
        summary, review, reports = load()
        summary["variants"].pop()
        with self.assertRaisesRegex(ValueError, "baseline variant count changed"):
            verify_result.validate_result(summary, review, reports)

    def test_changed_raw_report_rejected(self):
        summary, review, reports = load()
        key = summary["variants"][0]["case"], summary["variants"][0]["variant"]
        reports[key] += b" "
        with self.assertRaisesRegex(ValueError, "raw report missing or changed"):
            verify_result.validate_result(summary, review, reports)

    def test_missing_semantic_review_rejected(self):
        summary, review, reports = load()
        review["findings"].pop()
        with self.assertRaisesRegex(ValueError, "review omitted"):
            verify_result.validate_result(summary, review, reports)

    def test_foreign_semantic_review_rejected(self):
        summary, review, reports = load()
        decision = copy.deepcopy(review["findings"][0])
        decision["finding_id"] = "foreign"
        review["findings"].append(decision)
        with self.assertRaisesRegex(ValueError, "foreign"):
            verify_result.validate_result(summary, review, reports)

    def test_false_human_provenance_rejected(self):
        summary, review, reports = load()
        review["independent_human"] = True
        with self.assertRaisesRegex(ValueError, "review provenance"):
            verify_result.validate_result(summary, review, reports)


if __name__ == "__main__":
    unittest.main()
