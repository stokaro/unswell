"""Reject missing review grades and inflated FULL credit."""

import copy
import json
import tempfile
import unittest
from pathlib import Path

from score_model_pilot import grade, template

ROOT = Path(__file__).resolve().parent


class ScoreModelPilotTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.run_dir = Path(self.temporary.name)
        self.run_dir.joinpath("run.json").write_text(json.dumps({
            "format": "unswell-349-atomic-paired-run-v1", "complete": True,
            "status": "complete", "calls": [
                {"request": f"page-{page:03d}-{arm}", "status": "valid",
                 "bound_sha256": ""} for page in range(1, 10)
                for arm in ("freeform", "atomic")]}))
        journal = json.loads((self.run_dir / "run.json").read_text())
        for row in journal["calls"]:
            target = self.run_dir / row["request"]
            target.mkdir()
            arm = row["request"].rsplit("-", 1)[1]
            value = {"arm": arm, "diagnoses": []}
            (target / "bound.json").write_text(json.dumps(value))
            from score_model_pilot import sha
            row["bound_sha256"] = sha(target / "bound.json")
        (self.run_dir / "run.json").write_text(json.dumps(journal))

    def test_complete_empty_run_is_not_a_pilot_pass(self):
        draft = template(self.run_dir)
        for arm in draft["arms"]:
            for event in arm["events"]:
                event.update(status="MISSED", basis="No delivered diagnosis overlaps this event.")
        result = grade(self.run_dir, draft)
        self.assertFalse(result["advance"])
        self.assertEqual(result["arms"]["atomic"]["MISSED"], 15)
        self.assertEqual(result["arms"]["atomic"]["delivered"], 0)

    def test_missing_event_grade_rejected(self):
        with self.assertRaisesRegex(ValueError, "event review changed or incomplete"):
            grade(self.run_dir, template(self.run_dir))

    def test_partial_run_rejected(self):
        journal = json.loads((self.run_dir / "run.json").read_text())
        journal["complete"] = False
        (self.run_dir / "run.json").write_text(json.dumps(journal))
        with self.assertRaisesRegex(ValueError, "partial or invalid run"):
            template(self.run_dir)

    def test_changed_review_identity_rejected(self):
        draft = template(self.run_dir)
        changed = copy.deepcopy(draft)
        changed["reference_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "review identity changed"):
            grade(self.run_dir, changed)


if __name__ == "__main__":
    unittest.main()
