"""Check the fixed dual-gate calculation and its input binding."""

import importlib.util
from pathlib import Path
import unittest


PATH = Path(__file__).with_name("analyze.py")
SPEC = importlib.util.spec_from_file_location("dual_gate_analyze", PATH)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class DualGateAnalysisTest(unittest.TestCase):
    def test_whole_page_requires_precision_repair(self):
        result = MODULE.analyze(MODULE.SOURCE.read_bytes())
        rows = {(row["profile"], row["cohort"]): row for row in result["rows"]}
        self.assertEqual(rows["technical", "whole"]["full_gap"], 18)
        self.assertEqual(rows["technical", "whole"]["nonaccepted_suppressions_needed"], 27)
        self.assertEqual(rows["strict", "whole"]["nonaccepted_suppressions_needed"], 46)

    def test_rejects_changed_source(self):
        source = MODULE.SOURCE.read_bytes()
        with self.assertRaisesRegex(ValueError, "identity changed"):
            MODULE.analyze(source + b" ")

    def test_rejects_changed_result(self):
        source = MODULE.SOURCE.read_bytes()
        result = MODULE.render(MODULE.analyze(source))
        with self.assertRaisesRegex(ValueError, "differs"):
            MODULE.verify_result(result.replace(b'"full_gap": 18', b'"full_gap": 17', 1), source)


if __name__ == "__main__":
    unittest.main()
