"""Guard the frozen GEN screen's calibration and source-pair accounting."""

import importlib.util
import unittest
from pathlib import Path


spec = importlib.util.spec_from_file_location("gen_run", Path(__file__).with_name("run.py"))
assert spec is not None and spec.loader is not None
gen_run = importlib.util.module_from_spec(spec)
spec.loader.exec_module(gen_run)


class ScreenTests(unittest.TestCase):
    def test_threshold_limits_calibration_false_positives_with_ties(self):
        controls = [0.0] * 98 + [1.0, 1.0]
        cutoff = gen_run.threshold(controls)
        self.assertEqual(sum(value >= cutoff for value in controls), 0)
        self.assertGreater(cutoff, 1.0)

    def test_threshold_rejects_empty_controls(self):
        with self.assertRaisesRegex(ValueError, "Invalid calibration"):
            gen_run.threshold([])

    def test_auc_counts_ties_as_half(self):
        self.assertEqual(gen_run.auc([1.0] * 20, [1.0] * 20), 0.5)
        self.assertEqual(gen_run.auc([2.0] * 20, [1.0] * 20), 1.0)
        self.assertEqual(gen_run.auc([0.0] * 20, [1.0] * 20), 0.0)

    def test_ineligible_pairs_stay_in_coverage_denominator(self):
        group = ("arxiv", 1)
        humans = {group: {"prefix100": {"closed_fraction": 0.2}}}
        rows = [
            {"group": group, "scores": {"prefix100": {"closed_fraction": 0.4}},
             "edit_ratio": 30.0},
            {"group": group, "scores": {"prefix100": None}, "edit_ratio": 30.0},
        ]
        result = gen_run.summarize(rows, humans, 0.3, "prefix100", "closed_fraction")
        self.assertEqual(result["rows"], 2)
        self.assertEqual(result["eligible"], 1)
        self.assertEqual(result["coverage"], 0.5)
        self.assertEqual(result["groups"], 1)
        self.assertEqual(result["tpr"], 1.0)


if __name__ == "__main__":
    unittest.main()
