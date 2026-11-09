"""Check the frozen mask and advancement gate against false-positive passes."""

import importlib.util
import math
from pathlib import Path
import unittest


SCRIPT = Path(__file__).with_name("screen.py")
SPEC = importlib.util.spec_from_file_location("stacked_screen", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
screen = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(screen)


class ScreenTest(unittest.TestCase):
    def test_strict_confidence_and_retention(self):
        cutoff = math.log(0.01 / 0.99)
        self.assertEqual(screen.mask_count([cutoff]), (0, 0))
        self.assertEqual(screen.mask_count([cutoff - 1]), (1, 0))
        self.assertEqual(screen.mask_count([cutoff - 1] * 4), (4, 1))

    def test_both_cohorts_must_activate(self):
        sufficient = {"eligible": 100, "filtering_documents": 5}
        insufficient = {"eligible": 100, "filtering_documents": 4}
        self.assertTrue(screen.advance_decision(sufficient, sufficient))
        self.assertFalse(screen.advance_decision(sufficient, insufficient))
        self.assertFalse(screen.advance_decision(insufficient, sufficient))
        self.assertFalse(screen.advance_decision(sufficient, {"eligible": 0,
                                                          "filtering_documents": 0}))


if __name__ == "__main__":
    unittest.main()
