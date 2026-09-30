"""Reject corrupted, incomplete, and misleading saved reference evidence."""

import copy
import json
import unittest

import verify


class VerificationTest(unittest.TestCase):
    def setUp(self):
        self.controls = verify.read(verify.ROOT / "testdata/controls.json")
        self.protocol = verify.read(verify.ROOT / "protocol.json")
        self.records = verify.read(verify.ROOT / "records.json")

    def test_missing_lens_prediction_is_rejected(self):
        rows = self.records["lens"]["control_predictions"][:-1]
        with self.assertRaisesRegex(ValueError, "Missing or duplicate"):
            verify.lens_screen(self.controls, rows, self.protocol["lens"])

    def test_duplicate_lens_prediction_is_rejected(self):
        rows = copy.deepcopy(self.records["lens"]["control_predictions"])
        rows.append(copy.deepcopy(rows[0]))
        with self.assertRaisesRegex(ValueError, "Missing or duplicate"):
            verify.lens_screen(self.controls, rows, self.protocol["lens"])

    def test_unavailable_controls_cannot_reduce_the_denominator(self):
        rows = copy.deepcopy(self.records["lens"]["control_predictions"])
        rows[0]["result"] = {"available": False, "reason": "short_input"}
        with self.assertRaisesRegex(ValueError, "unavailable"):
            verify.lens_screen(self.controls, rows, self.protocol["lens"])

    def test_nan_is_rejected(self):
        rows = copy.deepcopy(self.records["lens"]["control_predictions"])
        rows[0]["result"]["score"] = float("nan")
        with self.assertRaisesRegex(ValueError, "Invalid"):
            verify.lens_screen(self.controls, rows, self.protocol["lens"])

    def test_one_grounding_direction_cannot_substitute_for_two(self):
        pair = {"forward": {"available": True, "all_input_tokens_retained": True, "support_score": 0.99, "published_label": 1}, "backward": {"available": True, "all_input_tokens_retained": True, "support_score": 0.01, "published_label": 0}, "pair_supported": True}
        with self.assertRaisesRegex(ValueError, "Pair flag"):
            verify.supported_pair(pair, 0.5)

    def test_incorrect_published_label_is_rejected(self):
        output = {"available": True, "all_input_tokens_retained": True, "support_score": 0.9, "published_label": 0}
        with self.assertRaisesRegex(ValueError, "Label differs"):
            verify.support(output, 0.5)

    def test_missing_grounding_group_is_rejected(self):
        mini = copy.deepcopy(self.records["minicheck"])
        mini["groups"].pop()
        with self.assertRaisesRegex(ValueError, "Missing MiniCheck"):
            verify.mini_screen(self.controls, mini, self.protocol["minicheck"])

    def test_forward_failure_still_validates_the_backward_input(self):
        pair = {"forward": {"available": True, "all_input_tokens_retained": True, "support_score": 0.01, "published_label": 0}, "backward": {"available": True, "all_input_tokens_retained": True, "support_score": 2.0, "published_label": 1}, "pair_supported": False}
        with self.assertRaisesRegex(ValueError, "Invalid support"):
            verify.supported_pair(pair, 0.5)


if __name__ == "__main__":
    unittest.main()
