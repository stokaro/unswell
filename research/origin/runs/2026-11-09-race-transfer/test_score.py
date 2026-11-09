"""Check the fixed RACE score and integrity-sensitive metric arithmetic."""

from __future__ import annotations

import importlib.util
import math
from pathlib import Path
import unittest


SCRIPT = Path(__file__).with_name("score.py")
SPEC = importlib.util.spec_from_file_location("race_score", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
SCORE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(SCORE)


class ScoreTests(unittest.TestCase):
    def test_rejects_nonfinite_or_wrong_class_logits(self) -> None:
        for values in ([0.0, 1.0], [0.0, math.nan, 2.0, 3.0]):
            with self.subTest(values=values):
                with self.assertRaisesRegex(ValueError, "Invalid four-class logits"):
                    SCORE.score_logits(values)

    def test_machine_score_is_one_minus_human_softmax(self) -> None:
        score, polished = SCORE.score_logits([0.0, 0.0, 0.0, 0.0])
        self.assertEqual(score, 0.75)
        self.assertEqual(polished, 0.25)

    def test_source_pair_metrics_keep_ties(self) -> None:
        pairs = [
            {"machine": {"race": 0.8}, "human": {"race": 0.2}},
            {"machine": {"race": 0.5}, "human": {"race": 0.5}},
        ]
        self.assertEqual(SCORE.paired_win(pairs, "race"), 0.75)
        self.assertEqual(SCORE.auc(pairs, "race"), 0.875)


if __name__ == "__main__":
    unittest.main()
