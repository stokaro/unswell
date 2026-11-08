"""Negative checks for the pinned #349 pilot evidence gate."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import packet
import verify_result

ROOT = Path(__file__).resolve().parent


class VerifyResultTest(unittest.TestCase):
    def test_complete_evidence(self):
        result = verify_result.verify()
        self.assertEqual(result["full_recalled"], 0)
        self.assertEqual(result["accepted"], 3)
        self.assertFalse(result["advance"])

    def test_missing_inventory_id_rejected(self):
        response = json.loads((ROOT / "run" / "uv" / "inventory" / "response.json").read_text())
        response["units"].pop()
        with self.assertRaisesRegex(ValueError, "omitted or added"):
            packet.validate_inventory("uv", response)

    def test_unbound_diagnosis_rejected(self):
        inventory_raw = (ROOT / "run" / "atuin" / "inventory" / "response.json").read_bytes()
        response = json.loads((ROOT / "run" / "atuin" / "diagnosis" / "response.json").read_text())
        response["findings"][0]["quote"] = "not present in the source"
        with self.assertRaisesRegex(ValueError, "not uniquely bound"):
            packet.validate_diagnosis("atuin", response, inventory_raw)

    def test_mutated_result_rejected(self):
        with tempfile.TemporaryDirectory() as temporary:
            copy = Path(temporary) / "pilot"
            shutil.copytree(ROOT, copy)
            result_path = copy / "result.json"
            result = json.loads(result_path.read_text())
            result["full_recalled"] = 4
            result_path.write_text(json.dumps(result))
            with self.assertRaisesRegex(ValueError, "pinned artifact changed"):
                verify_result.verify(copy)

    def test_prior_failed_attempt_cannot_gain_a_response(self):
        with tempfile.TemporaryDirectory() as temporary:
            copy = Path(temporary) / "pilot"
            shutil.copytree(ROOT, copy)
            (copy / "prior-v1" / "events.jsonl").write_text('{}\n')
            with self.assertRaisesRegex(ValueError, "first stopped attempt changed"):
                verify_result.verify(copy)


if __name__ == "__main__":
    unittest.main()
