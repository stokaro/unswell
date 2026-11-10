"""Reject altered NLI evidence and unsupported success claims."""

import copy
import importlib.util
import json
from pathlib import Path
import unittest


HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("nli_verify_v2", HERE / "verify_v2.py")
VERIFY = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VERIFY)
PREPARE_SPEC = importlib.util.spec_from_file_location("nli_prepare", HERE / "prepare.py")
PREPARE = importlib.util.module_from_spec(PREPARE_SPEC)
PREPARE_SPEC.loader.exec_module(PREPARE)


class EvidenceStateVerificationTest(unittest.TestCase):
    def setUp(self):
        self.freeze = json.loads((HERE / "freeze-v2.json").read_text())
        self.packet = json.loads((HERE / "packet.json").read_text())
        self.key = json.loads((HERE / "key.json").read_text())
        self.result = json.loads((HERE / "result-v2.json").read_text())

    def test_frozen_source_packet_rebuilds_exactly(self):
        packet, key = PREPARE.build()
        self.assertEqual(self.packet, packet)
        self.assertEqual(self.key, key)

    def test_negative_result_is_recomputed(self):
        summary = VERIFY.score(self.result, self.freeze, self.packet, self.key)
        self.assertEqual(summary, json.loads((HERE / "summary-v2.json").read_text()))
        self.assertFalse(summary["directional_gate_passed"])
        self.assertEqual(summary["relation_correct"], 9)
        self.assertEqual(summary["contradiction_false_admissions"], 2)

    def test_rejects_forged_relation(self):
        result = copy.deepcopy(self.result)
        result["rows"][0]["relation"] = "support"
        with self.assertRaisesRegex(ValueError, "disagrees with logits"):
            VERIFY.score(result, self.freeze, self.packet, self.key)

    def test_rejects_missing_case(self):
        result = copy.deepcopy(self.result)
        result["rows"].pop()
        with self.assertRaisesRegex(ValueError, "incomplete"):
            VERIFY.score(result, self.freeze, self.packet, self.key)

    def test_rejects_changed_logit_shape(self):
        result = copy.deepcopy(self.result)
        result["rows"][0]["logits"] = [1.0, 2.0]
        with self.assertRaisesRegex(ValueError, "invalid or missing scores"):
            VERIFY.score(result, self.freeze, self.packet, self.key)

    def test_rejects_false_model_identity(self):
        result = copy.deepcopy(self.result)
        result["model_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "model identity"):
            VERIFY.score(result, self.freeze, self.packet, self.key)


if __name__ == "__main__":
    unittest.main()
