"""Negative checks for the exposed current-engine editorial review ledger."""

import copy
import importlib.util
from pathlib import Path
import unittest


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("current_editorial_verify", HERE / "verify.py")
VERIFY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VERIFY)


class ReviewEvidenceTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.findings = VERIFY.read(HERE / "findings.json")
        cls.events = VERIFY.read(HERE / "events.json")
        cls.reports = {profile: VERIFY.read(VERIFY.BASE / (profile + ".json.gz"))
                       for profile in ("technical", "strict")}
        cls.baseline = VERIFY.read(VERIFY.BASE / "result.json")
        cls.labels = VERIFY.prior_verify.labels()
        cls.sources = VERIFY.replay.sources()

    def check(self, findings, events):
        return VERIFY.evaluate(findings, events, self.reports, self.baseline,
                               self.labels, self.sources)

    def test_missing_finding_rejected(self):
        with self.assertRaisesRegex(ValueError, "incomplete finding review"):
            self.check(self.findings[:-1], self.events)

    def test_uncovered_full_event_rejected(self):
        events = copy.deepcopy(self.events)
        event = next(row for row in events if row["event"] == "whole/p001-d02")
        event["status"] = "full"
        event["credited_finding_ids"] = [self.findings[0]["id"]]
        with self.assertRaisesRegex(ValueError, "unsupported full event"):
            self.check(self.findings, events)

    def test_inapplicable_advice_cannot_be_marked_conditional(self):
        findings = copy.deepcopy(self.findings)
        finding = next(row for row in findings if row["id"] == "acc17b97a56ec5af")
        finding["advice_status"] = "conditional_guidance"
        with self.assertRaisesRegex(ValueError, "advice disposition inconsistent"):
            self.check(findings, self.events)

    def test_actionable_finding_requires_repair(self):
        findings = copy.deepcopy(self.findings)
        finding = next(row for row in findings if row["id"] == "b4362a7123dcc3b2")
        finding["repair_action"] = None
        with self.assertRaisesRegex(ValueError, "repair proposal inconsistent"):
            self.check(findings, self.events)


if __name__ == "__main__":
    unittest.main()
