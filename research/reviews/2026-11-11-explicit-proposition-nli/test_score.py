"""Negative controls for frozen result scoring."""

import copy
import json
from pathlib import Path
import sys
import unittest

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
import score


class ScoreTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.packet = json.loads((HERE / "packet.json").read_text())
        cls.freeze = json.loads((HERE / "freeze.json").read_text())
        cls.freeze_sha = score.digest(HERE / "freeze.json")
        cls.result = {
            "format": "unswell-349-explicit-proposition-nli-result-v1",
            "freeze_sha256": cls.freeze_sha,
            "packet_sha256": cls.freeze["packet_sha256"],
            "model_sha256": cls.freeze["assets"]["onnx/model_quantized.onnx"]["sha256"],
            "rows": [dict(id=case["id"], token_ids=12,
                          logits=[3.0 if label == case["expected_relation"] else 0.0
                                  for label in score.LABELS],
                          relation=case["expected_relation"])
                     for case in cls.packet["cases"]]}

    def check(self, result):
        return score.score(self.packet, result, self.freeze, self.freeze_sha)

    def test_perfect_result_passes(self):
        result = self.check(self.result)
        self.assertTrue(result["directional_pass"])
        self.assertEqual(result["correct"], 16)

    def test_missing_case_rejected(self):
        result = copy.deepcopy(self.result)
        result["rows"].pop()
        with self.assertRaisesRegex(ValueError, "missing or extra cases"):
            self.check(result)

    def test_relation_must_match_logits(self):
        result = copy.deepcopy(self.result)
        result["rows"][0]["relation"] = "neutral"
        with self.assertRaisesRegex(ValueError, "relation disagrees with logits"):
            self.check(result)

    def test_neutral_false_entailment_fails_gate(self):
        result = copy.deepcopy(self.result)
        changed = 0
        for case, row in zip(self.packet["cases"], result["rows"]):
            if case["expected_relation"] == "neutral" and changed < 2:
                row["logits"] = [3.0, 0.0, 0.0]
                row["relation"] = "entailment"
                changed += 1
        summary = self.check(result)
        self.assertEqual(summary["correct"], 14)
        self.assertEqual(summary["neutral_false_entailment"], 2)
        self.assertFalse(summary["directional_pass"])


if __name__ == "__main__":
    unittest.main()
