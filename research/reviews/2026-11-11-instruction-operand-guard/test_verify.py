#!/usr/bin/env python3
"""Mutation tests for the exposed instruction-operand replay verifier."""

import copy
import unittest

import verify


class ReplayVerificationTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.record, cls.reports = verify.load_inputs()
        cls.archived = {profile: verify.read(verify.INPUTS[f"archived_{profile}"])
                        for profile in ("technical", "strict")}
        cls.review = verify.read(verify.INPUTS["prior_finding_review"])
        cls.events = verify.read(verify.INPUTS["prior_event_review"])
        cls.additions = verify.read(verify.INPUTS["prior_second_review"])
        cls.prior = verify.read(verify.INPUTS["prior_second_summary"])
        cls.sources = verify.expected_sources()

    def evaluate(self, after=None, review=None, events=None):
        return verify.evaluate(self.archived, self.reports["before"],
                               after if after is not None else self.reports["after"],
                               review if review is not None else self.review,
                               events if events is not None else self.events,
                               self.additions, self.prior, self.sources, self.record)

    def test_complete_paired_replay(self):
        result = self.evaluate()
        self.assertEqual(result["technical"]["after"], 160)
        self.assertEqual(result["strict"]["after"], 181)

    def test_rejects_unexpected_retained_warning(self):
        after = copy.deepcopy(self.reports["after"])
        removed = next(row for row in self.reports["before"]["technical"]["findings"]
                       if verify.location(row) in verify.REMOVED)
        after["technical"]["findings"].append(removed)
        with self.assertRaisesRegex(ValueError, "unexpected removed or added"):
            self.evaluate(after=after)

    def test_rejects_changed_diagnosis(self):
        after = copy.deepcopy(self.reports["after"])
        after["technical"]["findings"][0]["message"] += " changed"
        with self.assertRaisesRegex(ValueError, "retained diagnosis changed"):
            self.evaluate(after=after)

    def test_rejects_source_drift(self):
        after = copy.deepcopy(self.reports["after"])
        after["technical"]["documents"][0]["source_hash"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source coverage or identity changed"):
            self.evaluate(after=after)

    def test_rejects_removed_actionable_warning(self):
        review = copy.deepcopy(self.review)
        next(row for row in review if (row["rule_id"], row["path"],
             row["span"]["start"], row["span"]["end"]) in verify.REMOVED)["status"] = "actionable"
        with self.assertRaisesRegex(ValueError, "nonactionable review"):
            self.evaluate(review=review)

    def test_rejects_removed_event_witness(self):
        events = copy.deepcopy(self.events)
        reviewed = next(row for row in self.review if (row["rule_id"], row["path"],
                        row["span"]["start"], row["span"]["end"]) in verify.REMOVED)
        events[0]["credited_finding_ids"].append(reviewed["id"])
        with self.assertRaisesRegex(ValueError, "credited to an event"):
            self.evaluate(events=events)


if __name__ == "__main__":
    unittest.main()
