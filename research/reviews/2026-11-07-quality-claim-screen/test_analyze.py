"""Negative checks for the frozen quality-claim screen."""

import unittest

import analyze


class ScreenTests(unittest.TestCase):
    def test_committed_audit_matches_reports(self):
        self.assertEqual(analyze.analyze()["profiles"]["technical"]["new_accepted"], 5)

    def test_unreviewed_addition_is_rejected(self):
        report = analyze.read(analyze.HERE / "technical.json.gz")
        prior = analyze.read(analyze.PRIOR / "technical.json.gz")
        old_locations = {analyze.key(finding) for finding in prior["findings"]}
        additions = [finding for finding in report["findings"]
                     if analyze.key(finding) not in old_locations]
        target = next(index for index, finding in enumerate(additions)
                      if analyze.location(finding) == ("whole/p019.md", 598, 628))
        altered = dict(additions[target])
        altered["primary"] = dict(altered["primary"])
        altered["primary"]["span"] = {"start": 598, "end": 627}
        additions[target] = altered
        with self.assertRaisesRegex(ValueError, "Added locations differ"):
            analyze.validate_additions(additions, {}, {}, {})


if __name__ == "__main__":
    unittest.main()
