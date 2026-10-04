"""Negative controls for missing scope, false acceptance, and changed artifacts."""
import copy
import json
from pathlib import Path
import shutil
import tempfile
import unittest
import verify

ROOT = Path(__file__).resolve().parent


class AccountingTests(unittest.TestCase):
    def setUp(self):
        self.args = [json.loads((ROOT / name).read_text()) for name in
                     ["evaluation.json", "execution.json", "manifest.json", "protocol.json"]]

    def reject(self, mutate):
        values = copy.deepcopy(self.args)
        mutate(*values)
        with self.assertRaises(ValueError):
            verify.check_counts(*values)

    def test_complete_actual_accounting(self):
        verify.check_counts(*self.args)

    def test_changed_denominator(self):
        self.reject(lambda r, *_: r["cohorts"]["whole"].update(events=55))

    def test_removed_cohort(self):
        self.reject(lambda r, *_: r["cohorts"].pop("confirmation"))

    def test_false_development_acceptance(self):
        self.reject(lambda r, *_: r.update(development_screen_passed=True))

    def test_invented_available_page(self):
        self.reject(lambda r, *_: r["cohorts"]["whole"].update(available_pages=26))

    def test_lost_uncertain_finding(self):
        self.reject(lambda r, *_: r["cohorts"]["whole"].update(uncertain=83))

    def test_false_human_review(self):
        self.reject(lambda r, *_: r.update(independent_human_annotation=True))

    def test_changed_operator_count(self):
        self.reject(lambda r, *_: r["operators"]["evaluation"].update(accepted=57))

    def test_source_loss(self):
        self.reject(lambda r, e, m, p: m["pages"].pop())

    def test_repaired_original_status(self):
        self.reject(lambda r, *_: r["original_strict_trace_execution"].update(passed=True))

    def test_artifact_change(self):
        with tempfile.TemporaryDirectory() as temp:
            folder = Path(temp)
            for name in [*json.loads((ROOT / "public-artifacts.json").read_text()), "public-artifacts.json"]:
                shutil.copyfile(ROOT / name, folder / name)
            verify.verify(folder)
            with (folder / "prompt.md").open("a") as stream:
                stream.write("changed\n")
            with self.assertRaisesRegex(ValueError, "Changed public artifact"):
                verify.verify(folder)


if __name__ == "__main__":
    unittest.main()
