"""Negative contract tests for the frozen evidence-flip packet."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

import run


class PacketTests(unittest.TestCase):
    def test_requests_exclude_answer_key(self):
        pairs = run.entries()
        self.assertEqual(len(pairs), 16)
        for payload, key in pairs:
            self.assertEqual(set(payload), {"item_id", "source_sha256", "source",
                                            "target_quote", "target_start_byte",
                                            "target_end_byte", "proposed_diagnosis"})
            body = run.request(payload).decode("utf-8")
            sent = json.loads(body.split("INPUT JSON:\n", 1)[1])
            self.assertEqual(sent, payload)
            self.assertNotIn("applies", sent)
            self.assertNotIn("variant_id", sent)
            self.assertNotIn("rationale", sent)
            self.assertNotIn(key["variant_id"], body)

    def test_packet_verifies(self):
        self.assertEqual(run.verify()["calls_max"], 16)

    def test_packet_rejects_modified_request(self):
        with tempfile.TemporaryDirectory() as temp:
            copy = Path(temp) / "packet"
            shutil.copytree(run.PACKET, copy)
            first = next((copy / "requests").iterdir())
            first.write_bytes(first.read_bytes() + b"\n")
            with self.assertRaisesRegex(ValueError, "request set or bytes changed"):
                run.verify(copy)

    def test_packet_rejects_modified_freeze(self):
        with tempfile.TemporaryDirectory() as temp:
            copy = Path(temp) / "packet"
            shutil.copytree(run.PACKET, copy)
            (copy / "freeze.json").write_text("{}\n")
            with self.assertRaisesRegex(ValueError, "packet freeze changed"):
                run.verify(copy)

    def test_response_rejects_wrong_byte_binding(self):
        payload = run.entries()[0][0]
        response = self.valid_response(payload)
        response["target_start_byte"] += 1
        with self.assertRaisesRegex(ValueError, "source binding changed"):
            run.check_response(response, payload)

    def test_response_rejects_extra_field(self):
        payload = run.entries()[0][0]
        response = self.valid_response(payload)
        response["answer_key"] = True
        with self.assertRaisesRegex(ValueError, "response fields changed"):
            run.check_response(response, payload)

    def test_response_rejects_advice_on_reject(self):
        payload = run.entries()[0][0]
        response = self.valid_response(payload)
        response["replacement"] = "Delete this sentence."
        with self.assertRaisesRegex(ValueError, "nonadmission contains edit"):
            run.check_response(response, payload)

    def test_response_rejects_unchanged_admitted_edit(self):
        payload = run.entries()[0][0]
        response = self.valid_response(payload)
        response["decision"] = "admit"
        response["replacement"] = payload["target_quote"]
        with self.assertRaisesRegex(ValueError, "leaves target unchanged"):
            run.check_response(response, payload)

    @staticmethod
    def valid_response(payload):
        return {"item_id": payload["item_id"],
                "source_sha256": payload["source_sha256"],
                "target_start_byte": payload["target_start_byte"],
                "target_end_byte": payload["target_end_byte"],
                "decision": "reject", "replacement": "",
                "basis": "The complete source supports the target function."}


if __name__ == "__main__":
    unittest.main()
