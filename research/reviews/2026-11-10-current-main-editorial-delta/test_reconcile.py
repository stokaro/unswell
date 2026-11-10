#!/usr/bin/env python3
"""Negative tests for the #349 cross-review reconciliation."""

import importlib.util
from pathlib import Path
import unittest


HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location("reconcile", HERE / "reconcile.py")
reconcile = importlib.util.module_from_spec(spec)
spec.loader.exec_module(reconcile)


class ReconciliationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        profiles = ("technical", "strict")
        cls.claim_reports = {
            profile: reconcile.read(reconcile.CLAIMS / (profile + ".json.gz"))
            for profile in profiles
        }
        cls.replay_reports = {
            profile: reconcile.read(HERE / (profile + ".json.gz"))
            for profile in profiles
        }
        cls.scoped_result = reconcile.read(reconcile.SCOPED / "result.json")
        cls.claim_result = reconcile.read(reconcile.CLAIMS / "result.json")
        cls.replay_result = reconcile.read(HERE / "summary.json")
        cls.second_review = reconcile.read(HERE / "new-review.json")

    def evaluate(self, claim=None, replay=None, review=None):
        return reconcile.evaluate(
            self.claim_reports if claim is None else claim,
            self.replay_reports if replay is None else replay,
            self.scoped_result,
            self.claim_result,
            self.replay_result,
            self.second_review if review is None else review,
        )

    def test_changed_diagnosis_is_rejected(self):
        replay = dict(self.replay_reports)
        strict = dict(replay["strict"])
        strict["findings"] = list(strict["findings"])
        strict["findings"][0] = dict(strict["findings"][0])
        strict["findings"][0]["message"] += " changed"
        replay["strict"] = strict
        with self.assertRaisesRegex(ValueError, "finding diagnosis changed"):
            self.evaluate(replay=replay)

    def test_changed_source_is_rejected(self):
        claim = dict(self.claim_reports)
        technical = dict(claim["technical"])
        technical["documents"] = list(technical["documents"])
        technical["documents"][0] = dict(technical["documents"][0])
        technical["documents"][0]["source_hash"] = "0" * 64
        claim["technical"] = technical
        with self.assertRaisesRegex(ValueError, "source identity changed"):
            self.evaluate(claim=claim)

    def test_disagreement_cannot_be_silently_removed(self):
        review = [dict(row) for row in self.second_review]
        row = next(row for row in review if row["event"] == "whole/p023-d01")
        row["event_status"] = "full"
        with self.assertRaisesRegex(ValueError, "review disagreement identities changed"):
            self.evaluate(review=review)


if __name__ == "__main__":
    unittest.main()
