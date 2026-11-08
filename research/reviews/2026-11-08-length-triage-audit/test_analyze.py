"""Negative checks for the frozen length-triage accounting."""

import copy
import importlib.util
import json
from pathlib import Path
import unittest


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("length_triage", HERE / "analyze.py")
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class LengthTriageAuditTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.freeze = json.loads((HERE / "freeze.json").read_text())
        cls.inputs = MODULE.read_inputs(cls.freeze)

    def test_frozen_source_digest_rejects_changed_input(self):
        changed = copy.deepcopy(self.freeze)
        name = next(iter(changed["inputs"]))
        changed["inputs"][name] = "0" * 64
        with self.assertRaisesRegex(ValueError, "Input changed"):
            MODULE.read_inputs(changed)

    def test_missing_decision_does_not_become_nonaccepted(self):
        changed = copy.deepcopy(self.inputs)
        changed["findings.json"] = [row for row in changed["findings.json"]
                                     if row["id"] != "1a61a1b770b72c44"]
        with self.assertRaisesRegex(ValueError, "review denominators changed"):
            MODULE.analyze(changed)

    def test_moved_finding_span_is_rejected(self):
        changed = copy.deepcopy(self.inputs)
        row = next(row for row in changed["technical.json.gz"]["findings"]
                   if row["id"] == "1a61a1b770b72c44")
        row["primary"]["span"]["start"] += 1
        with self.assertRaisesRegex(ValueError, "source span changed"):
            MODULE.analyze(changed)


if __name__ == "__main__":
    unittest.main()
