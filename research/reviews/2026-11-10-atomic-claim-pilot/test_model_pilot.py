"""Mechanical negative tests for the frozen paired-model packet."""

import copy
import json
import tempfile
import unittest
from pathlib import Path

from model_pilot import (ARMS, bind_quote, check_response, order, pages,
                         prepare, request, schema, verify)
from verify_reference import verify as verify_reference

ROOT = Path(__file__).resolve().parent


class ModelPilotTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.pages = pages()
        cls.reference = verify_reference(ROOT)

    def test_order_and_same_inputs(self):
        self.assertEqual(len(order()), 18)
        self.assertEqual(order()[:4], [(1, "freeform"), (1, "atomic"),
                                        (2, "atomic"), (2, "freeform")])
        for arm in ARMS:
            sent = request(self.pages[0], arm).split(b"INPUT JSON:\n", 1)[1]
            self.assertEqual(json.loads(sent), self.pages[0])
            self.assertEqual(set(schema(arm)["required"]),
                             {"page_id", "source_sha256", "diagnoses"} |
                             ({"units"} if arm == "atomic" else set()))

    def test_exact_event_binding(self):
        event = self.reference["events"][0]
        page = self.pages[event["page"] - 1]
        segments = page["units"][event["unit_ordinal"]]["segments"]
        index = next(i for i, span in enumerate(segments)
                     if span["start"] <= event["byte_span"]["start"] and
                     event["byte_span"]["end"] <= span["end"])
        self.assertEqual(bind_quote(page, event["unit_ordinal"], index, event["quote"]),
                         event["byte_span"])

    def test_bad_or_protected_quote_rejected(self):
        page = self.pages[0]
        with self.assertRaisesRegex(ValueError, "missing or ambiguous"):
            bind_quote(page, 0, 0, "fabricated quote")
        excluded = page["excluded"][0]["span"]
        protected = page["source"].encode()[excluded["start"]:excluded["end"]].decode()
        with self.assertRaises(ValueError):
            bind_quote(page, 0, 0, protected)

    def test_freeform_response_binding(self):
        page = self.pages[0]
        event = self.reference["events"][0]
        response = {"page_id": page["page_id"], "source_sha256": page["source_sha256"],
                    "diagnoses": [{"unit_ordinal": event["unit_ordinal"],
                                   "segment_index": 0, "quote": event["quote"],
                                   "problem": "Nested lead-in hides the check.",
                                   "preserve": "Check for underfitting or overfitting."}]}
        bound = check_response(response, page, "freeform")
        self.assertEqual(bound["diagnoses"][0]["byte_span"], event["byte_span"])
        changed = copy.deepcopy(response)
        changed["source_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source identity"):
            check_response(changed, page, "freeform")
        changed = copy.deepcopy(response)
        changed["diagnoses"][0]["quote"] = "not in source"
        with self.assertRaisesRegex(ValueError, "quote missing"):
            check_response(changed, page, "freeform")
        changed = copy.deepcopy(response)
        changed["diagnoses"].append(copy.deepcopy(changed["diagnoses"][0]))
        with self.assertRaisesRegex(ValueError, "duplicate diagnosis"):
            check_response(changed, page, "freeform")

    def test_atomic_unit_denominator(self):
        page = self.pages[0]
        response = {"page_id": page["page_id"], "source_sha256": page["source_sha256"],
                    "units": [{"ordinal": i, "empty_reason": "no independent claim",
                               "claims": []} for i in range(len(page["units"]))],
                    "diagnoses": []}
        self.assertEqual(len(check_response(response, page, "atomic")["units"]),
                         len(page["units"]))
        changed = copy.deepcopy(response)
        changed["units"].pop()
        with self.assertRaisesRegex(ValueError, "omitted"):
            check_response(changed, page, "atomic")
        changed = copy.deepcopy(response)
        changed["units"][0]["empty_reason"] = ""
        with self.assertRaisesRegex(ValueError, "empty reason"):
            check_response(changed, page, "atomic")
        changed = copy.deepcopy(response)
        changed["units"][0]["ordinal"] = True
        with self.assertRaisesRegex(ValueError, "unit order"):
            check_response(changed, page, "atomic")

    def test_packet_freeze_rejects_request_change(self):
        with tempfile.TemporaryDirectory() as directory:
            packet = Path(directory) / "packet"
            prepare(packet)
            verify(packet)
            path = next((packet / "requests").iterdir())
            path.write_bytes(path.read_bytes() + b" ")
            with self.assertRaisesRegex(ValueError, "packet changed"):
                verify(packet)


if __name__ == "__main__":
    unittest.main()
