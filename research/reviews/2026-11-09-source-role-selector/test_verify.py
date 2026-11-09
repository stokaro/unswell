"""Negative controls for the offline cue-transfer evidence check."""

from __future__ import annotations

import copy
import unittest

import verify


class TransferEvidenceTest(unittest.TestCase):
    def setUp(self) -> None:
        self.inventory, self.inventory_sha = verify.read_json("transfer_v2_inventory.json")
        self.labels, self.labels_sha = verify.read_json("transfer_v2_labels.json")
        self.result, _ = verify.read_json("transfer_v2_result.json")
        self.selector_sha = self.result["selector_sha256"]

    def check(self, labels: dict, result: dict) -> None:
        verify.verify_transfer_result(self.inventory, labels, result,
                                      self.inventory_sha, self.labels_sha,
                                      self.selector_sha)

    def test_frozen_result(self) -> None:
        self.check(self.labels, self.result)

    def test_rejects_changed_gold_role(self) -> None:
        labels = copy.deepcopy(self.labels)
        labels["labels"][0]["role"] = "admit"
        with self.assertRaisesRegex(ValueError, "frozen role"):
            self.check(labels, self.result)

    def test_rejects_changed_decision(self) -> None:
        result = copy.deepcopy(self.result)
        result["rows"][0]["decision"] = "withhold"
        with self.assertRaisesRegex(ValueError, "matrix differs"):
            self.check(self.labels, result)

    def test_rejects_changed_target_span(self) -> None:
        inventory = copy.deepcopy(self.inventory)
        inventory["targets"][0]["end"] += 1
        with self.assertRaisesRegex(ValueError, "quote length"):
            verify.verify_transfer_result(inventory, self.labels, self.result,
                                          self.inventory_sha, self.labels_sha,
                                          self.selector_sha)


if __name__ == "__main__":
    unittest.main()
