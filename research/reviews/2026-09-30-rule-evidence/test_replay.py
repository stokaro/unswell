"""Reject invented evidence, revised source, and silently changed rule actions."""

import copy
import hashlib
import unittest

import replay


class RuleReplayTests(unittest.TestCase):
    def fixture(self):
        raw = "The café client retries."
        source_hash = hashlib.sha256(raw.encode()).hexdigest()
        span = {"start": 0, "end": len(raw.encode())}
        location = {"path": "example.md", "span": span, "segments": [span]}
        evidence = {"kind": "exact", "message": "", "suggestion": "", "metrics": [
            {"name": "length", "value": 4, "unit": "prose-words", "onset": 3, "saturation": 6}]}
        finding = {"id": "f1", "rule_id": "rule-id", "rule_version": "1", "message": "Review its length.",
                   "primary": location, "related": [], "evidence": evidence, "suppressed": False}
        observation = {"id": "f1", "rule_id": "rule-id", "rule_version": "1", "diagnostic": finding["message"],
                       "source_sha256": source_hash, "primary": location, "related": [], "evidence": evidence,
                       "suppressed": False, "reason": "The reported length is four prose-words.",
                       "action": finding["message"], "action_basis": "engine_diagnostic", "editorial_verdict": "unreviewed"}
        report = {"status": "complete", "manifest": {"tool_commit": "original", "config_hash": "policy"},
                  "documents": [{"name": "example.md", "source_hash": source_hash}], "findings": [finding]}
        observations = {"version": "unswell-rule-evidence-v1", "tool_commit": "original", "config_hash": "policy",
                        "observations": [observation], "editorial_qualified": False}
        candidate = {"id": "rule/f1", "origin": "rule", "category": "rule-id", "diagnostic": finding["message"],
                     "rationale": "", "suggestion": "", "targets": [{"span": span, "quote": raw, "units": ["u1"]}]}
        document = {"source": raw, "source_sha256": source_hash}
        return report, observations, candidate, document

    def test_explicit_child_preserves_original_and_exact_targets(self):
        report, observations, candidate, document = self.fixture()
        before = copy.deepcopy(candidate)
        findings, rows = replay.index_report(report, observations)
        child = replay.bind_original(candidate, document, "example.md", findings["f1"], rows["f1"])
        self.assertEqual(candidate, before)
        self.assertEqual(child["targets"], before["targets"])
        self.assertEqual(child["rationale"], observations["observations"][0]["reason"])
        self.assertEqual(child["suggestion"], report["findings"][0]["message"])

    def test_missing_or_repeated_observations_are_not_coverage(self):
        report, observations, _, _ = self.fixture()
        for rows in [[], observations["observations"] * 2]:
            modified = copy.deepcopy(observations)
            modified["observations"] = rows
            with self.assertRaises(ValueError):
                replay.index_report(report, modified)

    def test_source_metrics_version_action_and_verdict_must_survive(self):
        report, observations, _, _ = self.fixture()
        mutations = [lambda r: r.update(source_sha256="different"),
                     lambda r: r.update(rule_version="new-version"),
                     lambda r: r["evidence"]["metrics"][0].update(value=99),
                     lambda r: r.update(action="Delete the whole sentence."),
                     lambda r: r.update(editorial_verdict="accepted"),
                     lambda r: r["primary"]["span"].update(start=1)]
        for mutation in mutations:
            modified = copy.deepcopy(observations)
            mutation(modified["observations"][0])
            with self.assertRaises(ValueError):
                replay.index_report(report, modified)

    def test_source_quote_or_target_changes_are_rejected(self):
        report, observations, candidate, document = self.fixture()
        finding, observation = report["findings"][0], observations["observations"][0]
        for field, changed in [("quote", "A different sentence."), ("span", {"start": 1, "end": 3})]:
            modified = copy.deepcopy(candidate)
            modified["targets"][0][field] = changed
            with self.assertRaises(ValueError):
                replay.bind_original(modified, document, "example.md", finding, observation)
        document["source"] += " Another sentence."
        with self.assertRaises(ValueError):
            replay.bind_original(candidate, document, "example.md", finding, observation)

    def test_incomplete_or_qualified_report_cannot_pass(self):
        report, observations, _, _ = self.fixture()
        report["status"] = "incomplete"
        with self.assertRaises(ValueError):
            replay.index_report(report, observations)
        report["status"] = "complete"
        observations["editorial_qualified"] = True
        with self.assertRaises(ValueError):
            replay.index_report(report, observations)


if __name__ == "__main__":
    unittest.main()
