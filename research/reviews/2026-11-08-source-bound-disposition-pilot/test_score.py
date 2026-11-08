"""Reject incomplete or inconsistent post-run review evidence."""

import copy
import unittest
from unittest import mock

import packet
import score


class ScoreGateTest(unittest.TestCase):
    def setUp(self):
        self.original_read = packet.read
        self.review = self.original_read(score.ROOT / "review.json")

    def scored_with(self, modified_review):
        def read(path):
            if path == score.ROOT / "review.json":
                return modified_review
            return self.original_read(path)

        with mock.patch.object(packet, "read", side_effect=read):
            return score.score()

    def test_complete_review_matches_recorded_result(self):
        self.assertEqual(score.score(), self.original_read(score.ROOT / "result.json"))

    def test_missing_delivered_finding_review_rejected(self):
        review = copy.deepcopy(self.review)
        review["finding_reviews"].pop()
        with self.assertRaisesRegex(ValueError, "complete delivered stream"):
            self.scored_with(review)

    def test_untraceable_event_grade_rejected(self):
        review = copy.deepcopy(self.review)
        review["event_reviews"][0]["finding_indices"] = [999]
        with self.assertRaisesRegex(ValueError, "traceable delivered finding"):
            self.scored_with(review)

    def test_protected_flag_mismatch_rejected(self):
        review = copy.deepcopy(self.review)
        review["protected_reviews"][0]["flagged"] = True
        with self.assertRaisesRegex(ValueError, "protected flag disagrees"):
            self.scored_with(review)

    def test_frozen_packet_tampering_rejected(self):
        with mock.patch.object(score.run, "files", return_value={}):
            with self.assertRaisesRegex(ValueError, "frozen packet hash mismatch"):
                score.score()


if __name__ == "__main__":
    unittest.main()
