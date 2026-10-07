"""Negative controls for public accounting; these do not adjudicate wording."""
from copy import deepcopy
import json
import unittest

import verify


def records():
    return [json.loads((verify.ROOT / name).read_text()) for name in
        ("evaluation.json", "execution.json", "manifest.json", "protocol.json", "comparison.json", "component-ablation.json")]


class PublicAccounting(unittest.TestCase):
    def test_completed_record(self):
        self.assertTrue(verify.check(*records())["public_identities_and_counts_verified"])

    def test_cannot_claim_failed_cohort_or_product_passes(self):
        for key in ("development_screen_passed", "product_qualified", "objective_achieved", "standalone_generation_qualified", "prospective_confirmation_measured"):
            values = records()
            values[0][key] = True
            with self.assertRaises(ValueError):
                verify.check(*values)

    def test_original_scope_and_aliases_cannot_be_reduced(self):
        for position, change in ((0, "scope"), (2, "page"), (2, "alias")):
            values = records()
            if change == "scope":
                values[position]["original_scope"]["events"] = 122
            elif change == "page":
                values[position]["pages"].pop()
            else:
                values[position]["pages"][0]["actual_input_aliases"] -= 1
            with self.assertRaises(ValueError):
                verify.check(*values)

    def test_uncertain_and_partial_counts_cannot_disappear(self):
        for field in ("uncertain", "missed", "nonempty_supplied_advice"):
            values = records()
            values[0]["cohorts"]["whole"][field] += 1
            with self.assertRaises(ValueError):
                verify.check(*values)

    def test_ratios_thresholds_and_advice_gate_are_checked(self):
        values = records()
        values[0]["cohorts"]["whole"]["accepted_fraction"] = 0.9
        with self.assertRaises(ValueError):
            verify.check(*values)
        values = records()
        values[3]["numeric_targets"]["accepted_fraction"] = 0.75
        with self.assertRaises(ValueError):
            verify.check(*values)
        row = deepcopy(records()[0]["cohorts"]["whole"])
        row.update(accepted=row["delivered"], accepted_fraction=1, passed=True)
        rows = deepcopy(records()[0]["cohorts"])
        rows["whole"] = row
        with self.assertRaises(ValueError):
            verify.check_cohorts(rows)

    def test_execution_cannot_add_retries_or_restart_deadline(self):
        for key, value in (("retries", 1), ("new_full_calls", 33), ("elapsed_seconds_including_pilot_and_preparation", 7201)):
            values = records()
            values[1][key] = value
            with self.assertRaises(ValueError):
                verify.check(*values)

    def test_component_analysis_is_not_isolated_inference(self):
        values = records()
        values[5]["isolated_component_inference_measured"] = True
        with self.assertRaises(ValueError):
            verify.check(*values)


if __name__ == "__main__":
    unittest.main()
