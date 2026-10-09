"""Check that support reports preserve unobserved editorial decisions."""

import copy
import unittest

from annotation_contract import normalize_review
from label_support import summarize_label_support


class LabelSupportTest(unittest.TestCase):
    def test_retention_is_not_eight_intent_negatives(self):
        review = normalize_review({"necessity": {"status": "retain", "reason": "The condition is needed."}},
                                  source="The retry stops when the deadline expires.")
        result = summarize_label_support({"source/1": review})
        self.assertEqual(result["labels"]["necessity"]["negative"], 1)
        self.assertTrue(all(row["unobserved"] == 1 for name, row in result["labels"].items()
                            if name != "necessity"))
        self.assertFalse(result["class_support_is_model_qualification"])

    def test_nonexclusive_positive_and_explicit_negative(self):
        positive = normalize_review({
            "necessity": {"status": "necessary", "reason": "The opening adds no information."},
            "edits": [{"quote": "It is important to note that ", "replacement": "",
                       "intents": ["empty_framing", "wordiness"], "status": "necessary",
                       "reason": "The sentence states the behavior without the announcement."}],
        }, source="It is important to note that retries stop.")
        negative = normalize_review({
            "necessity": {"status": "retain", "reason": "The sentence states a useful condition."},
            "intents": {"empty_framing": {"status": "not_necessary",
                                         "reason": "There is no content-free announcement."}},
        }, source="Retries stop when the deadline expires.")
        result = summarize_label_support({"source/1": positive, "source/2": negative})
        self.assertTrue(result["labels"]["empty_framing"]["both_classes_observed"])
        self.assertEqual(result["labels"]["wordiness"]["positive"], 1)
        self.assertEqual(result["labels"]["wordiness"]["negative"], 0)
        self.assertFalse(result["labels"]["wordiness"]["both_classes_observed"])
        self.assertFalse(result["unmarked_tokens_are_background"])

    def test_optional_decisions_remain_unobserved(self):
        review = normalize_review({
            "necessity": {"status": "optional", "reason": "Both orders are clear."},
            "intents": {"clarity": {"status": "optional", "reason": "The current order is usable."}},
        }, source="The buffer is flushed after the write.")
        result = summarize_label_support({"source/1": review})
        self.assertEqual(result["labels"]["necessity"]["unobserved"], 1)
        self.assertEqual(result["labels"]["clarity"]["unobserved"], 1)

    def test_invalid_mask_and_label_fail(self):
        review = normalize_review({}, source="Retries stop.")
        for label, mask in [(None, True), (0, False), (True, True), (0.0, True), (2, True), (None, "false")]:
            changed = copy.deepcopy(review)
            changed["intent_loss_labels"]["clarity"] = label
            changed["intent_loss_mask"]["clarity"] = mask
            with self.subTest(label=label, mask=mask), self.assertRaises(ValueError):
                summarize_label_support({"source/1": changed})

    def test_missing_dimension_fails_instead_of_defaulting(self):
        review = normalize_review({}, source="Retries stop.")
        del review["intent_loss_labels"]["clarity"]
        with self.assertRaises(ValueError):
            summarize_label_support({"source/1": review})

    def test_empty_packet_is_unqualified(self):
        result = summarize_label_support({})
        self.assertEqual(result["source_cases"], 0)
        self.assertTrue(all(not row["both_classes_observed"] for row in result["labels"].values()))
        self.assertFalse(result["class_support_is_model_qualification"])


if __name__ == "__main__":
    unittest.main()
