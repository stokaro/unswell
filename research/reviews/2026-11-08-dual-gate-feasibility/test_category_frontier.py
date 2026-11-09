"""Negative integrity checks for the exposed #349 category ceiling."""

import copy
import importlib.util
import json
import tempfile
import unittest
from pathlib import Path
from unittest import mock


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("category_frontier", HERE / "category_frontier.py")
FRONTIER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(FRONTIER)


class CategoryFrontierTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.inputs = [FRONTIER.load_verified(name)
                      for name in ("events", "scoped", "quality", "dual_gate")]

    def test_saved_ceiling_and_recovery_requirements(self):
        result = FRONTIER.analyze(*copy.deepcopy(self.inputs))
        self.assertEqual(result, json.loads((HERE / "category-frontier.json").read_text()))
        self.assertEqual([result["cohorts"][name]["full_now"]
                          for name in ("whole", "context", "confirmation")],
                         [28, 16, 4])
        self.assertEqual([result["cohorts"][name]["strategies"]["empty_framing"]
                          ["recall_gate_reachable"]
                          for name in ("whole", "context", "confirmation")],
                         [False, False, False])
        self.assertEqual(result["cohorts"]["confirmation"]["strategies"]
                         ["empty_framing_and_wordiness"]["minimum_selected_recovery_fraction"],
                         1.0)

    def test_duplicate_or_removed_event_rejected(self):
        for mutation in ("duplicate", "removed"):
            with self.subTest(mutation=mutation):
                inputs = copy.deepcopy(self.inputs)
                if mutation == "duplicate":
                    inputs[0][0]["event"] = inputs[0][1]["event"]
                else:
                    inputs[0].pop()
                with self.assertRaises(ValueError):
                    FRONTIER.analyze(*inputs)

    def test_reviewed_addition_must_have_been_missed(self):
        inputs = copy.deepcopy(self.inputs)
        target = next(row for row in inputs[0]
                      if row["event"] == "whole/p023-d01")
        target["status"] = "full"
        with self.assertRaisesRegex(ValueError, "already credited"):
            FRONTIER.analyze(*inputs)

    def test_reviewed_cohort_count_must_agree(self):
        inputs = copy.deepcopy(self.inputs)
        inputs[2]["profiles"]["technical"]["cohorts"]["whole"]["full"] += 1
        with self.assertRaisesRegex(ValueError, "reviewed count mismatch"):
            FRONTIER.analyze(*inputs)

    def test_changed_source_hash_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "events.json"
            path.write_text("[]")
            with mock.patch.dict(FRONTIER.INPUTS,
                                 {"events": (path, FRONTIER.INPUTS["events"][1])}):
                with self.assertRaisesRegex(ValueError, "source hash changed"):
                    FRONTIER.load_verified("events")


if __name__ == "__main__":
    unittest.main()
