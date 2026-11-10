#!/usr/bin/env python3
"""Negative tests for the frozen #349 editorial delta."""

import copy
import importlib.util
from pathlib import Path
import unittest


HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("delta_verify", HERE / "verify.py")
delta = importlib.util.module_from_spec(spec)
spec.loader.exec_module(delta)


class DeltaVerificationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.old_reports = {
            p: delta.read(delta.BASE / (p + ".json.gz"))
            for p in ("technical", "strict")
        }
        cls.current_reports = {
            p: delta.read(HERE / (p + ".json.gz"))
            for p in ("technical", "strict")
        }
        cls.old_findings = delta.read(delta.OLD / "findings.json")
        cls.old_events = delta.read(delta.OLD / "events.json")
        cls.new_review = delta.read(HERE / "new-review.json")
        cls.baseline = delta.read(delta.BASE / "result.json")
        cls.sources = delta.replay.sources()
        cls.labels = delta.labels_verify.labels()
        cls.commit = delta.read(HERE / "freeze.json")["runtime_commit"]

    def evaluate(self, current=None, new_review=None, labels=None):
        return delta.evaluate(
            self.old_reports,
            self.current_reports if current is None else current,
            self.old_findings,
            self.old_events,
            self.new_review if new_review is None else new_review,
            self.baseline,
            self.sources,
            self.labels if labels is None else labels,
            self.commit,
        )

    def test_preserved_finding_change_rejected(self):
        current = dict(self.current_reports)
        strict = dict(current["strict"])
        strict["findings"] = list(strict["findings"])
        strict["findings"][0] = dict(strict["findings"][0])
        strict["findings"][0]["message"] += " altered"
        current["strict"] = strict
        with self.assertRaisesRegex(ValueError, "previously reviewed finding changed"):
            self.evaluate(current=current)

    def test_unreviewed_new_finding_rejected(self):
        with self.assertRaisesRegex(ValueError, "unexpected new finding count"):
            self.evaluate(new_review=self.new_review[:-1])

    def test_frozen_target_quote_change_rejected(self):
        labels = copy.deepcopy(self.labels)
        labels["context/c02-d03"]["targets"][0]["quote"] += " altered"
        with self.assertRaisesRegex(ValueError, "target quote drift"):
            self.evaluate(labels=labels)


if __name__ == "__main__":
    unittest.main()
