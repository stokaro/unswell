"""Negative checks for the source-bound disposition pilot gate."""

import copy
import json
import unittest

import packet


class PacketGateTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        packet.validate_sources_and_labels()
        cls.slug = "dagger"
        page_id, raw, record = packet.page(cls.slug)
        cls.raw = raw
        cls.record = record
        cls.inventory = packet.encoded({
            "page_id": page_id,
            "source_sha256": packet.sha(raw),
            "units": [{"ordinal": unit["ordinal"], "function": "other",
                       "proposition": "", "support": "absent"}
                      for unit in record["units"]],
        })
        packet.validate_inventory(cls.slug, json.loads(cls.inventory))

    def response(self):
        page_id, raw, record = packet.page(self.slug)
        return {"page_id": page_id, "source_sha256": packet.sha(raw),
                "inventory_sha256": packet.sha(self.inventory),
                "dispositions": [{"unit_ordinal": unit["ordinal"],
                                  "decision": "retain", "reason": "No supported correction."}
                                 for unit in record["units"]],
                "findings": []}

    def finding(self):
        label = packet.read(packet.ROOT / "labels.json")["events"][0]
        quote = label["quote"]
        start = self.raw.index(quote.encode("utf-8"))
        return {"unit_ordinal": 2, "line": self.raw[:start].count(b"\n") + 1,
                "quote": quote, "diagnosis": "Universal scope is unsupported.",
                "evidence": "The README later gives a runtime requirement.",
                "source_function": "capability", "deletion_test": "Remove the universal claim."}

    def test_complete_retained_response_passes(self):
        bound = packet.validate_treatment(self.slug, self.response(), self.inventory)
        self.assertEqual(len(bound["dispositions"]), 51)
        self.assertEqual(bound["findings"], [])

    def test_valid_revision_binds_original_bytes(self):
        response = self.response()
        response["dispositions"][2]["decision"] = "revise"
        response["findings"] = [self.finding()]
        bound = packet.validate_treatment(self.slug, response, self.inventory)
        self.assertEqual(bound["findings"][0]["start_byte"], 88)

    def test_missing_or_duplicate_disposition_rejected(self):
        for mutation in ("missing", "duplicate"):
            with self.subTest(mutation=mutation):
                response = self.response()
                if mutation == "missing":
                    response["dispositions"].pop()
                else:
                    response["dispositions"][1] = copy.deepcopy(response["dispositions"][0])
                with self.assertRaises(ValueError):
                    packet.validate_treatment(self.slug, response, self.inventory)

    def test_revise_without_bound_finding_rejected(self):
        response = self.response()
        response["dispositions"][2]["decision"] = "revise"
        with self.assertRaises(ValueError):
            packet.validate_treatment(self.slug, response, self.inventory)

    def test_finding_without_revise_rejected(self):
        response = self.response()
        response["findings"] = [self.finding()]
        with self.assertRaises(ValueError):
            packet.validate_treatment(self.slug, response, self.inventory)

    def test_unbound_quote_rejected(self):
        response = self.response()
        response["dispositions"][2]["decision"] = "revise"
        finding = self.finding()
        finding["quote"] = "This text does not occur in the source."
        response["findings"] = [finding]
        with self.assertRaisesRegex(ValueError, "not uniquely bound"):
            packet.validate_treatment(self.slug, response, self.inventory)

    def test_malformed_finding_rejected(self):
        response = self.response()
        response["findings"] = [None]
        with self.assertRaises(ValueError):
            packet.validate_treatment(self.slug, response, self.inventory)


if __name__ == "__main__":
    unittest.main()
