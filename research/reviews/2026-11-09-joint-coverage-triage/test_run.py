"""Negative controls for the frozen pilot's response and source contract."""

import copy
import unittest

import run


class BinderTest(unittest.TestCase):
    def response(self, number):
        inp = run.payload(number)
        return {
            "page_id": inp["page_id"],
            "source_sha256": inp["source_sha256"],
            "reviews": [{"id": finding["id"], "decision": "keep",
                         "reason": "The full page supports this correction."}
                        for finding in inp["baseline_findings"]],
            "new_findings": [],
        }

    def test_complete_baseline_review_and_original_byte_binding(self):
        response = self.response(8)
        quote = "the easiest way"
        raw = run.source_path(8).read_bytes()
        start = raw.index(quote.encode("utf-8"))
        response["new_findings"].append({
            "line": raw[:start].count(b"\n") + 1,
            "quote": quote,
            "diagnosis": "An absolute ease claim lacks a comparison basis.",
            "source_function": "promotional ranking",
            "reader_burden": "Readers cannot assess the ranking.",
            "retain": "The product supports Vue forms.",
        })
        bound = run.bind(8, response)
        self.assertEqual(bound["new_findings"][0]["id"], "page-008/new-01")
        self.assertEqual(bound["new_findings"][0]["start_byte"], start)
        self.assertEqual(bound["new_findings"][0]["end_byte"], start + len(quote))

    def test_missing_baseline_decision_is_invalid(self):
        response = self.response(2)
        response["reviews"].pop()
        with self.assertRaisesRegex(ValueError, "coverage"):
            run.bind(2, response)

    def test_duplicate_baseline_id_is_invalid(self):
        response = self.response(2)
        response["reviews"][1]["id"] = response["reviews"][0]["id"]
        with self.assertRaisesRegex(ValueError, "duplicate"):
            run.bind(2, response)

    def test_unrecognized_decision_is_invalid(self):
        response = self.response(2)
        response["reviews"][0]["decision"] = "pass"
        with self.assertRaisesRegex(ValueError, "invalid review decision"):
            run.bind(2, response)

    def test_quote_must_bind_to_original_line(self):
        response = self.response(8)
        response["new_findings"] = [{
            "line": 1,
            "quote": "the easiest way",
            "diagnosis": "Unsupported ranking",
            "source_function": "promotion",
            "reader_burden": "The comparison has no basis.",
            "retain": "The product supports Vue forms.",
        }]
        with self.assertRaisesRegex(ValueError, "not uniquely bound"):
            run.bind(8, response)

    def test_protected_source_bytes_are_rejected(self):
        response = self.response(2)
        response["new_findings"] = [{
            "line": 1,
            "quote": '<div align="center">',
            "diagnosis": "Unsupported structure",
            "source_function": "HTML wrapper",
            "reader_burden": "The wrapper is redundant.",
            "retain": "The heading remains.",
        }]
        with self.assertRaisesRegex(ValueError, "protected bytes"):
            run.bind(2, response)

    def test_source_identity_change_is_invalid(self):
        response = copy.deepcopy(self.response(8))
        response["source_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "identity changed"):
            run.bind(8, response)

    def test_frozen_packet_rejects_unlisted_request(self):
        if not (run.PACKET / "freeze.json").exists():
            self.skipTest("packet not prepared yet")
        extra = run.PACKET / "requests" / "unexpected.txt"
        self.assertFalse(extra.exists())
        try:
            extra.write_text("unexpected", encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "packet hash mismatch|request set changed"):
                run.verify()
        finally:
            extra.unlink(missing_ok=True)


if __name__ == "__main__":
    unittest.main()
