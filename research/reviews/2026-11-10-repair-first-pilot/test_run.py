"""Negative source-binding tests for the repair-first research packet."""

import copy
from pathlib import Path
import tempfile
import unittest

import run


class RepairBindingTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.row = next(row for row in run.rows() if row["page_id"] == "eza.md")
        unit = next(unit for unit in cls.row["eligible_units"]
                    if unit["ordinal"] == 5)
        cls.item = {
            "unit_ordinal": 5,
            "quote": "one single binary",
            "replacement": "single binary",
            "category": "wordiness",
            "diagnosis": "One and single repeat the same quantity.",
            "reader_benefit": "The count remains clear without duplication.",
            "retained_meaning": "The build produces one binary.",
        }
        assert cls.item["quote"] in unit["raw_slice"]

    def response(self):
        return {"page_id": self.row["page_id"],
                "source_sha256": self.row["source_sha256"],
                "edits": [copy.deepcopy(self.item)]}

    def test_exact_source_edit_binds(self):
        bound = run.bind(self.response(), self.row)
        self.assertEqual(len(bound["edits"]), 1)
        target = bound["edits"][0]["span"]
        self.assertEqual(self.row["source"].encode("utf-8")
                         [target["start"]:target["end"]], b"one single binary")

    def test_wrong_source_hash_rejected(self):
        response = self.response()
        response["source_sha256"] = "0" * 64
        with self.assertRaisesRegex(ValueError, "source identity"):
            run.bind(response, self.row)

    def test_edit_outside_unit_rejected(self):
        response = self.response()
        response["edits"][0]["unit_ordinal"] = 1
        with self.assertRaisesRegex(ValueError, "quote not unique"):
            run.bind(response, self.row)

    def test_expansion_rejected(self):
        response = self.response()
        response["edits"][0]["replacement"] = "one single binary plus invented facts"
        with self.assertRaisesRegex(ValueError, "quote or replacement length"):
            run.bind(response, self.row)

    def test_unknown_fields_rejected(self):
        response = self.response()
        response["edits"][0]["old_score"] = 1
        with self.assertRaisesRegex(ValueError, "edit fields"):
            run.bind(response, self.row)

    def test_frozen_request_tamper_rejected(self):
        original = run.PACKET
        with tempfile.TemporaryDirectory(dir=run.ROOT) as scratch:
            try:
                run.PACKET = Path(scratch) / "packet"
                run.prepare()
                run.verify()
                request = run.PACKET / "requests" / "eza.md.txt"
                request.write_bytes(request.read_bytes() + b"changed")
                with self.assertRaisesRegex(ValueError, "packet freeze changed"):
                    run.verify()
            finally:
                run.PACKET = original


if __name__ == "__main__":
    unittest.main()
