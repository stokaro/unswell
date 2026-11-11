"""Negative accounting checks for the exposed historical utility audit."""

import copy
import unittest

from score import load_inputs, score


class ScoreTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.inputs = load_inputs()

    def scored(self, *, ledger=None, labels=None, sources=None, review=None):
        original_ledger, original_labels, original_sources, original_review, digest = self.inputs
        return score(
            copy.deepcopy(ledger if ledger is not None else original_ledger),
            copy.deepcopy(labels if labels is not None else original_labels),
            copy.deepcopy(sources if sources is not None else original_sources),
            copy.deepcopy(review if review is not None else original_review),
            digest,
        )

    def test_complete_review(self):
        result = self.scored()
        self.assertEqual(result["dispositions"]["accepted"], 4)
        self.assertEqual(result["delivered_findings"], 12)
        self.assertEqual(result["positive_pages_with_span_overlap"], 2)
        self.assertEqual(result["positive_pages_with_full_event_match"], 0)

    def test_missing_disposition_fails(self):
        review = copy.deepcopy(self.inputs[3])
        review["findings"].pop()
        with self.assertRaisesRegex(ValueError, "every delivered finding"):
            self.scored(review=review)

    def test_unbound_event_credit_fails(self):
        review = copy.deepcopy(self.inputs[3])
        review["full_event_matches"] = [{"pr": "fullstack-hy2020/fullstack-hy2020.github.io#2168", "start": 1709}]
        with self.assertRaisesRegex(ValueError, "unbound event match"):
            self.scored(review=review)

    def test_incomplete_report_fails(self):
        ledger = copy.deepcopy(self.inputs[0])
        ledger["rows"][0]["status"] = "incomplete"
        with self.assertRaisesRegex(ValueError, "incomplete report"):
            self.scored(ledger=ledger)

    def test_changed_source_identity_fails(self):
        sources = copy.deepcopy(self.inputs[2])
        sources[0]["sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source mismatch"):
            self.scored(sources=sources)

    def test_invalid_disposition_fails(self):
        review = copy.deepcopy(self.inputs[3])
        review["findings"][0]["status"] = "maybe"
        with self.assertRaisesRegex(ValueError, "invalid disposition"):
            self.scored(review=review)


if __name__ == "__main__":
    unittest.main()
